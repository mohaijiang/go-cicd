package dispatcher

import "state-example/model"

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
}
