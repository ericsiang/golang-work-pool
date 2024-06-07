package work_pool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Worker 代表處理請求的 worker。
type Worker struct {
	Id         int
	Wg         *sync.WaitGroup
	ReqHandler map[int]RequestHandler
}

// LaunchWorker 啟動工作執行緒來處理傳入的請求。
// 它在一個單獨的 goroutine 中運行，持續監聽輸入通道上的傳入請求。
// 當輸入通道關閉或收到停止訊號時，工作執行緒會優雅地停止。
func (w *Worker) LaunchWorker(in chan WorkerRequest, stopCh chan struct{}) {
	go func() {
		defer w.Wg.Done()
		for {
			select {
			case msg, open := <-in:
				if !open {
					// 如果通道關閉，則停止處理並返回
					// 如果我們跳過關閉通道檢查，則在關閉通道後，worker 繼續從關閉的通道讀取空值。
					fmt.Println("Stopping worker:", w.Id)
					return
				}
				w.processRequest(msg)
				time.Sleep(1 * time.Microsecond) // 延遲小，防止環路過緊
			case <-stopCh:
				fmt.Println("Stopping worker:", w.Id)
				return
			}
		}
	}()
}

// processRequest 處理單一請求。
func (w *Worker) processRequest(msg WorkerRequest) {
	fmt.Printf("Worker %d processing request: %v\n", w.Id, msg)
	var handler RequestHandler
	var ok bool
	if handler, ok = w.ReqHandler[msg.Type]; !ok {
		fmt.Println("Handler not implemented: workerID:", w.Id)
	} else {
		if msg.TimeOut == 0 {
			msg.TimeOut = time.Duration(10 * time.Millisecond) // Default timeout
		}
		// 在 MaxRetries 數量內，嘗試去執行 ReqHandler
		for attempt := 0; attempt <= msg.MaxRetries; attempt++ {
			var err error
			done := make(chan struct{})
			ctx, cancel := context.WithTimeout(context.Background(), msg.TimeOut)
			defer cancel()

			go func() {
				err = handler(msg.Data)
				close(done)
			}()

			select {
			case <-done:
				if err == nil {
					return // Successfully processed
				}
				fmt.Printf("Worker %d: Error processing request: %v\n", w.Id, err)
			case <-ctx.Done():
				fmt.Printf("Worker %d: Timeout processing request: %v\n", w.Id, msg.Data)
			}

			fmt.Printf("Worker %d: Retry %d for request %v\n", w.Id, attempt, msg.Data)
		}
		fmt.Printf("Worker %d: Failed to process request %v after %d retries\n", w.Id, msg.Data, msg.MaxRetries)
	}
}
