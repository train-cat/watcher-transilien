package main

import (
	"fmt"
	"sync"
	"syscall"
	"os"
	"os/signal"

	"github.com/Eraac/train-sniffer/utils"
)

func main() {
	quit := make(chan struct{})
	wg := sync.WaitGroup{}

	fmt.Println("Puller started")
	queue, err := pull(quit)

	if err != nil {
		utils.Error(err.Error())
		os.Exit(utils.ErrorInitQueue)
	}

	go func() {
		wg.Add(1)
		fmt.Println("Worker started")
		consume(queue)
		fmt.Println("Jobs finished")
		wg.Done()
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ch:
		fmt.Println("Gracefull quit...")
		quit <- struct{}{}
		wg.Wait()
	}
}
