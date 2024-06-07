package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
	"wp/work_pool"
)

func main() {
	// GOMAXPROCS 設定為可用 CPU 的數量
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	fmt.Printf("Running with %d CPUs\n", numCPU)

	// Configuration
	bufferSize := 50000
	maxWorkers := 20
	minWorkers := 3
	loadThreshold := 10000
	requests := 50000

	var wg sync.WaitGroup
	dispatcher := work_pool.NewDispatcher(bufferSize, &wg, maxWorkers)

	// Start 初始化 workers
	for i := 0; i < minWorkers; i++ {
		fmt.Printf("Starting worker with id %d\n", i)
		w := &work_pool.Worker{
			Wg:         &wg,
			Id:         i,
			ReqHandler: work_pool.ReqHandler,
		}
		dispatcher.AddWorker(w)
		fmt.Println("ReqHandler : ", w.ReqHandler)
	}

	// 建立 goroutine 來動態調整
	go dispatcher.ScaleWorkers(minWorkers, maxWorkers, loadThreshold, 0.75)

	// Send requests to the dispatcher
	for i := 0; i < requests; i++ {
		var handler work_pool.RequestHandler
		handler = Handler
		req := work_pool.WorkerRequest{
			Data:    fmt.Sprintf("(Msg_id: %d) -> Hello", i),
			Handler: handler,
			Type:    1,
			TimeOut: 5 * time.Second,
		}
		dispatcher.MakeRequest(req)
	}

	// Gracefully stop the dispatcher
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dispatcher.Stop(ctx)
	fmt.Println("Exiting main!")
}

func Handler(test any) error {
	fmt.Println("Handler:", test)
	return nil
}
