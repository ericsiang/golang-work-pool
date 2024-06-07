package work_pool

import "time"

// RequestHandler 定義了一個用於處理請求的函數類型
type RequestHandler func(interface{}) error

// WrokerRequest 代表要由 worker 處理的請求。
type WorkerRequest struct {
	Handler    RequestHandler
	Type       int
	Data       interface{}
	TimeOut    time.Duration // 請求的超時時長
	Retries    int           // 重試次數
	MaxRetries int           // 最大重試次數
}
