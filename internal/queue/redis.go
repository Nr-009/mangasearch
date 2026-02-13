package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mangasearch/internal/db"
	"mangasearch/internal/search"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client        *redis.Client
	ctx           context.Context
	mu            *sync.Mutex
	maxWorkers    int
	activeWorkers int
	queueName     string
	retries       int
	db            *db.DB
	es            *search.Client
}

func NewRedisQueue(workers int, database *db.DB, esClient *search.Client) *RedisQueue {
	return &RedisQueue{
		client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}),
		ctx:        context.Background(),
		mu:         &sync.Mutex{},
		maxWorkers: workers,
		queueName:  "ocr_queue",
		retries:    3,
		db:         database,
		es:         esClient,
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

func (queue *RedisQueue) worker() {
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
		fmt.Println("worker error:", err)
		return
	}

	dataPath := result[1]
	for idx := 0; idx < queue.retries; idx++ {
		if process(dataPath, queue.db, queue.es) == nil {
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
		if queue.getActiveWorkers() < queue.getMaxWorkers() {
			queue.addWorker()
			go queue.worker()
		} else {
			time.Sleep(100 * time.Millisecond)
		}
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

func (queue *RedisQueue) setMaxWorkers(workers int) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	queue.maxWorkers = workers
}

func (queue *RedisQueue) addWorker() {
	queue.mu.Lock()
	queue.activeWorkers++
	queue.mu.Unlock()
}
