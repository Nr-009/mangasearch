package main

import (
	"fmt"
	"mangasearch/internal/queue"
)

func main() {
	fmt.Println("starting...")

	q := queue.NewRedisQueue(4)

	paths := []string{}
	for i := 1; i <= 1000; i++ {
		paths = append(paths, fmt.Sprintf("/manga/Berserk/Ch%d/%03d.jpg", i, i))
	}

	q.Start(paths)
}
