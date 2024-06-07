package work_pool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ReqHandler 是請求處理程序的 map，以請求類型為鍵。
var ReqHandler = map[int]RequestHandler{
	1: func(data interface{}) error {
		return nil
	},
}

// dispatcher 代表管理一組 workers 並在他們之間分配傳入的請求。
type dispatcher struct {
	inCh        chan WorkerRequest
	wg          *sync.WaitGroup
	mu          sync.Mutex
	workerCount int
	stopCh      chan struct{} // Channel to signal workers to stop
}

// NewDispatcher 建立一個帶有緩衝通道和等待群組的新調度程序
func NewDispatcher(b int, wg *sync.WaitGroup, maxWorkers int) *dispatcher {
	return &dispatcher{
		inCh:   make(chan WorkerRequest, b),
		wg:     wg,
		stopCh: make(chan struct{}, maxWorkers), // Buffered channel 緩衝通道可防止停止時阻塞
	}
}

// AddWorker 將新 worker 新增至池中並增加 worker 數量
func (d *dispatcher) AddWorker(w WorkerLauncher) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.workerCount++
	d.wg.Add(1)
	w.LaunchWorker(d.inCh, d.stopCh)
}

// RemoveWorker 如果 worker 數量大於 minWorkers，則從池中刪除 worker
func (d *dispatcher) RemoveWorker(minWorkers int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.workerCount > minWorkers {
		d.workerCount--
		d.stopCh <- struct{}{} // Signal a worker to stop
	}
}

// ScaleWorkers 根據負載動態調整 worker 數量
func (d *dispatcher) ScaleWorkers(minWorkers, maxWorkers, loadThreshold int, loadLimit float64) {
	// 定時器
	ticker := time.NewTicker(time.Microsecond)
	defer ticker.Stop()

	for range ticker.C {
		load := len(d.inCh) // load 代表當前負載，是通道中待處理請求的數量
		if load > loadThreshold && d.workerCount < maxWorkers {
			// 如果目前負載大於負載閾值 loadThreshold 且 worker 數量小於最大可容許數量 maxWorkers，則表示負載過高，需要增加 worker
			fmt.Println("Scaling Triggered")
			// 新增 worker
			newWorker := &Worker{
				Wg:         d.wg,
				Id:         d.workerCount,
				ReqHandler: ReqHandler,
			}
			d.AddWorker(newWorker)
		} else if float64(load) < loadLimit*float64(loadThreshold) && d.workerCount > minWorkers {
			//如果目前負載小於負載閾值的 loadLimit  且 worker 數量大於最小可容許數量，則表示負載過低，可以刪除 worker 
			fmt.Println("Reducing Triggered")
			d.RemoveWorker(minWorkers) // 刪除 worker
		}
	}
}

// LaunchWorker 啟動一個 worker 並增加 worker 數量
func (d *dispatcher) LaunchWorker(id int, w WorkerLauncher) {
	w.LaunchWorker(d.inCh, d.stopCh) // Pass stopCh to the worker
	d.mu.Lock()
	d.workerCount++
	d.mu.Unlock()
}

// MakeRequest 將請求新增至輸入通道，或在通道已滿時丟棄該請求
func (d *dispatcher) MakeRequest(r WorkerRequest) {
	select {
	case d.inCh <- r:
	default:
		// 通道滿時的處理
		fmt.Println("Request channel is full. Dropping request.")
		// 或者，您可以記錄、緩衝請求或採取其他操作
	}
}

// Stop 優雅地停止所有 workers，等待他們完成處理
func (d *dispatcher) Stop(ctx context.Context) {
	fmt.Println("\nstop called")
	close(d.inCh) // Close the input channel to signal no more requests will be sent
	done := make(chan struct{})

	go func() {
		d.wg.Wait() // Wait for all workers to finish
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("All workers stopped gracefully")
	case <-ctx.Done():
		fmt.Println("Timeout reached, forcing shutdown")
		// 如果逾時則強制停止所有 worker
		for i := 0; i < d.workerCount; i++ {
			d.stopCh <- struct{}{}
		}
	}

	d.wg.Wait()
}
