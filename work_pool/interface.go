package work_pool

import "context"

// WorkerLauncher 是用來啟動 worker 的介面
type WorkerLauncher interface {
	LaunchWorker(in chan WorkerRequest, stopCh chan struct{})
}

// Dispatcher 是管理工作池的介面
type WorkerDispatcher interface {
	AddWorker(w WorkerLauncher)
	RemoveWorker(minWorkers int)
	LaunchWorker(id int, w WorkerLauncher)
	ScaleWorkers(minWorkers, maxWorkers, loadThreshold int, laodLimit float64)
	MakeRequest(WorkerRequest)
	Stop(ctx context.Context)
}
