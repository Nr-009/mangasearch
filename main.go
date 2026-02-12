package main

import (
	"fmt"
	"mangasearch/internal/queue"
)

func main() {
	fmt.Println("starting...")
	q := queue.NewRedisQueue(4)
	paths := []string{
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-001.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-002.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-003.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-004.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-005.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-006.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-007.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-008.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-009.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-010.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-011.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-012.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-013.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-014.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-015.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-016.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-017.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-018.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-019.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-020.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-021.jpg",
    "/Users/naranjitanaranjoso/Downloads/Berzerk/130/0130-022.jpg",
	}
	
	q.Start(paths)
}
