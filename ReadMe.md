# Work Pool
參考:
[Mastering Concurrent Processing: A Step-by-Step Guide to Building a Scalable Worker Pool in Go](https://medium.com/@souravchoudhary0306/mastering-concurrent-processing-a-step-by-step-guide-to-building-a-scalable-worker-pool-in-go-54093074c612)

### Worker
Worker 結構代表處理請求的 workers。每個工作線程都在自己的 goroutine 中運行，並偵聽通道上傳入的請求。

* LaunchWorker：在單獨的 goroutine 中啟動工作執行緒。工作執行緒處理傳入的請求，直到輸入通道關閉或收到停止訊號。
* processRequest：處理單一請求。如果發生錯誤或請求逾時，它將重試請求，最多達到指定的最大重試次數。

### Dispatcher
Dispatcher 負責管理 workers 並在他們之間分配傳入的請求。它可以根據當前負載動態新增或刪除 workers，並確保所有 workers 正常關閉。

* AddWorker：將新 workers 新增至池中並增加 workers 數量。啟動工作程序以開始處理請求。
* RemoveWorker：如果 workers 數量超過最低要求，則從池中刪除工作人員。透過 stopCh 頻道向 workers 發出停止訊號。
* ScaleWorkers：根據負載動態調整 worker 數量。如果負載超過閾值且 workers 數量少於允許的最大數量，則會新增 workers。如果負載低於閾值且 workers 數量超過所需的最低 workers 數量，則會刪除workers。
* LaunchWorker：啟動工作程序並增加工作程序計數。這通常用於初始 workers。
* MakeRequest：向輸入通道新增請求。如果通道已滿，請求將被丟棄，並記錄一則訊息。
* Stop：優雅地停止所有 workers。它等待所有 workers 完成當前請求的處理。如果達到超時，則強制停止所有 workers