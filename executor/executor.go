package executor

import "state-example/model"

type IExecutor interface {

	// FetchJob 获取任务
	FetchJob(name string) ([]byte, error)

	// Execute 执行任务
	Execute(job *model.Job)

	// HandlerLog 处理日志
	HandlerLog(jobId int)

	//SendResultToQueue 发送结果到队列
	SendResultToQueue(channel chan any)
}
