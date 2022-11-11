package dispatcher

import (
	"state-example/executor"
	"state-example/model"
)

type IDispatcher interface {
	// DispatchNode 选择节点
	DispatchNode(job *model.Job) *model.Node
	// Register 节点注册
	Register(node *model.Node)
	// UnRegister 节点注销
	UnRegister(node *model.Node)

	// HealthcheckNode 节点心跳
	HealthcheckNode(node *model.Node)

	// SendJob 发送任务
	SendJob(job *model.Job, node *model.Node)

	// CancelJob 取消任务
	CancelJob(job *model.Job, node *model.Node)

	// GetExecutor 根据节点获取执行器
	// TODO ... 这个方法设计的不好，分布式机构后应当用api代替
	GetExecutor(node *model.Node) executor.IExecutor
}

type Dispatcher struct {
	Channel chan model.QueueMessage
}

func NewDispatcher(channel chan model.QueueMessage) *Dispatcher {
	return &Dispatcher{
		channel,
	}
}

// DispatchNode 选择节点
func (d *Dispatcher) DispatchNode(job *model.Job) *model.Node {

	//TODO ... 单机情况直接返回 本地
	return nil
}

// Register 节点注册
func (d *Dispatcher) Register(node *model.Node) {
	return
}

// UnRegister 节点注销
func (d *Dispatcher) UnRegister(node *model.Node) {
	return
}

// HealthcheckNode 节点心跳
func (d *Dispatcher) HealthcheckNode(*model.Node) {
	// TODO  ... 检查注册的心跳信息，超过3分钟没有更新的节点，踢掉
	return
}

// SendJob 发送任务
func (d *Dispatcher) SendJob(job *model.Job, node *model.Node) {

	// TODO ... 单机情况下 不考虑节点，直接发送本地
	// TODO ... 集群情况下 通过注册的ip 地址进行api接口调用

	d.Channel <- model.NewStartQueueMsg(job.Name, 1)

	return
}

// CancelJob 取消任务
func (d *Dispatcher) CancelJob(job *model.Job, node *model.Node) {

	d.Channel <- model.NewStopQueueMsg(job.Name, 1)
	return
}
