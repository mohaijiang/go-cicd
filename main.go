package main

import (
	"context"
	"fmt"
	"os"
	"state-example/pipeline"
)

// 流程引擎
func main() {

	engineContext := make(map[string]interface{})
	engineContext["hamsterRoot"] = "/tmp/example"
	engineContext["workdir"] = "/tmp/example"
	engineContext["name"] = "test1"

	ctx, _ := context.WithCancel(context.WithValue(context.Background(), "stack", engineContext))

	var stack Stack

	actionGit := pipeline.NewGitAction("https://gitee.com/mohaijiang/spring-boot-example.gitt", "master", ctx)
	actionEnv := pipeline.NewDockerEnv("maven:3.5-jdk-8", ctx)
	actionShell := pipeline.NewShellAction("mvn clean package -Dmaven.test.skip=true", ctx)

	actions := []pipeline.ActionHandler{actionGit, actionEnv, actionShell}

	// 1：执行中 2：执行失败，3：执行成功
	status := 1
	for _, action := range actions {
		err := action.Pre()
		if err != nil {
			fmt.Println("error:", "action.Pre() error: %v", err)
			status = 2
			break
		}

		stack.push(action)
		err = action.Hook()
		if err != nil {
			fmt.Println("error:", "action.Hook() error: %v", err)
			status = 2
			break
		}
	}

	for !stack.isEmpty() {
		action, _ := stack.pop()
		_ = action.Post()
	}

	fmt.Println("status: ", status)
	_ = os.RemoveAll(engineContext["hamsterRoot"].(string))
}
