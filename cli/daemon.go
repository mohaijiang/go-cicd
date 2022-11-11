package cli

import (
	"state-example/dispatcher"
	"state-example/executor"
	"state-example/model"
)

func Main() {

	channel := make(chan model.QueueMessage)

	dispatch := dispatcher.NewDispatcher(channel)

	// 本地注册
	dispatch.Register(&model.Node{
		Name:    "localhost",
		Address: "127.0.0.1",
	})

	// 启动executor

	executeClient := executor.NewExecutorClient(channel)
	defer close(channel)

	go executeClient.Main()
}
