package executor

import (
	"context"
	"fmt"
	"io"
	"state-example/action"
	"state-example/model"
	"state-example/util"
)

type IExecutor interface {

	// FetchJob 获取任务
	FetchJob(name string) (io.Reader, error)

	// Execute 执行任务
	Execute(id int, job *model.Job) error

	// HandlerLog 处理日志
	HandlerLog(jobId int)

	//SendResultToQueue 发送结果到队列
	SendResultToQueue(channel chan any)

	Cancel(name string)
}

type Executor struct {
	cancelMap map[string]func()
}

// FetchJob 获取任务
func (e *Executor) FetchJob(name string) (io.Reader, error) {

	//TODO... 根据name 从rpc 或 直接内部调用获取job的pipeline文件
	return nil, nil
}

// Execute 执行任务
func (e *Executor) Execute(id int, job *model.Job) error {

	jobWrapper := model.JobWrapper{
		Job:    *job,
		Status: model.STATUS_NOTRUN,
	}

	// 1. 解析对pipeline 进行任务排序

	stages, err := job.StageSort()
	if err != nil {
		return err
	}

	// 2. 初始化 执行器的上下文

	engineContext := make(map[string]interface{})
	engineContext["hamsterRoot"] = "/tmp/example"
	engineContext["workdir"] = "/tmp/example"
	engineContext["name"] = job.Name

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), "stack", engineContext))

	// 将取消hook 记录到内存中,用于中断程序
	//TODO... 增加阻塞功能，一个 node 只允许执行一个pipeline 根据job.Name 阻塞（或加锁）
	e.cancelMap[job.Name] = cancel

	// 队列堆栈
	var stack util.Stack[action.ActionHandler]

	executeAction := func(ah action.ActionHandler, job model.JobWrapper) error {
		if jobWrapper.Status != model.STATUS_RUNNING {
			return nil
		}
		err := ah.Pre()
		if err != nil {
			job.Status = model.STATUS_FAIL
			fmt.Println(err)
			return err
		}
		stack.Push(ah)
		err = ah.Hook()
		if err != nil {
			job.Status = model.STATUS_FAIL
			return err
		}
		return nil
	}

	for _, stageWapper := range stages {
		//TODO ... stage的输出也需要换成堆栈方式
		fmt.Println("stage : ", stageWapper.Name)
		for _, step := range stageWapper.Stage.Steps {
			var ah action.ActionHandler
			if step.RunsOn != "" {
				ah = action.NewDockerEnv(step.RunsOn, ctx)
				err = executeAction(ah, jobWrapper)
			}
			if step.Uses == "" {
				ah = action.NewShellAction(step.Run, ctx)
				err = executeAction(ah, jobWrapper)
			}
			if step.Uses == "git-checkout" {
				ah = action.NewGitAction(step.With["url"], step.With["branch"], ctx)
				err = executeAction(ah, jobWrapper)
			}

		}
		for !stack.IsEmpty() {
			ah, _ := stack.Pop()
			_ = ah.Post()
		}

		if err != nil {
			cancel()
			break
		}
	}

	delete(e.cancelMap, job.Name)
	if err == nil {
		jobWrapper.Status = model.STATUS_SUCCESS
	} else {
		jobWrapper.Status = model.STATUS_FAIL
	}

	//TODO ... 发送结果到队列
	e.SendResultToQueue(nil)

	return err

}

// HandlerLog 处理日志
func (e *Executor) HandlerLog(jobId int) {

	//TODO ...
}

// SendResultToQueue 发送结果到队列
func (e *Executor) SendResultToQueue(channel chan any) {
	//TODO ...
}

// Cancel 取消
func (e *Executor) Cancel(name string) {

	cancel, ok := e.cancelMap[name]
	if ok {
		cancel()
	}
}
