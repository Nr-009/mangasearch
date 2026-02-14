package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mangasearch/internal/db"
	"mangasearch/internal/ocr"
	"mangasearch/internal/search"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client        *redis.Client
	ctx           context.Context
	mu            *sync.Mutex
	maxWorkers    int
	activeWorkers int
	workerCounter int
	queueName     string
	retries       int
	db            *db.DB
	es            *search.Client
	ocr           *ocr.Client
}

func NewRedisQueue(workers int, redisAddr string, database *db.DB, esClient *search.Client, ocrClient *ocr.Client) *RedisQueue {
	return &RedisQueue{
		client: redis.NewClient(&redis.Options{
			Addr: redisAddr,
			DB:   0,
		}),
		ctx:        context.Background(),
		mu:         &sync.Mutex{},
		maxWorkers: workers,
		queueName:  "ocr_queue",
		retries:    3,
		db:         database,
		es:         esClient,
		ocr:        ocrClient,
	}
}

func (queue *RedisQueue) Push(dataPath string) error {
	return queue.client.RPush(queue.ctx, queue.queueName, dataPath).Err()
}

func (queue *RedisQueue) IsEmpty() bool {
	length, err := queue.client.LLen(queue.ctx, queue.queueName).Result()
	if err != nil {
		return false
	}
	return length == 0
}

func (queue *RedisQueue) worker(id int) {
	defer func() {
		queue.mu.Lock()
		queue.activeWorkers--
		queue.mu.Unlock()
	}()

	result, err := queue.client.BRPop(queue.ctx, 5*time.Second, queue.queueName).Result()
	if err == redis.Nil {
		return
	}
	if err != nil {
		fmt.Printf("[worker %d] error: %v\n", id, err)
		return
	}

	dataPath := result[1]
	for idx := 0; idx < queue.retries; idx++ {
		if process(dataPath, queue.db, queue.es, queue.ocr, id) == nil {
			break
		}
	}
}

func (queue *RedisQueue) Start(paths []string) {
	for _, path := range paths {
		if err := queue.Push(path); err != nil {
			fmt.Println("push error:", err)
		}
	}

	for !queue.IsEmpty() {
		for queue.getActiveWorkers() < queue.getMaxWorkers() {
			queue.mu.Lock()
			queue.activeWorkers++
			queue.workerCounter++
			id := queue.workerCounter
			queue.mu.Unlock()
			go queue.worker(id)
		}
		time.Sleep(100 * time.Millisecond)
	}

	for queue.getActiveWorkers() > 0 {
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("all done")
}

func (queue *RedisQueue) getActiveWorkers() int {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	return queue.activeWorkers
}

func (queue *RedisQueue) getMaxWorkers() int {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	return queue.maxWorkers
}
