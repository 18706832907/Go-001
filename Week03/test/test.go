package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	done := make(chan bool, 1)
	var mu sync.Mutex

	i, j := 0, 0

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				mu.Lock()
				i++
				time.Sleep(100 * time.Microsecond)
				mu.Unlock()
			}
		}
	}()

	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Microsecond)
		mu.Lock()
		j++
		mu.Unlock()
	}

	done <- true
	fmt.Println(i, " -- ", j)
}
