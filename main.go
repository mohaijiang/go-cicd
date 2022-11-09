package main

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"state-example/model"
	"state-example/pipeline"
)

func main() {
	job := getJob()
	engine(job)
}

// 流程引擎
func engine(job *model.Job) {

	engineContext := make(map[string]interface{})
	engineContext["hamsterRoot"] = "/tmp/example"
	engineContext["workdir"] = "/tmp/example"
	engineContext["name"] = job.Name

	ctx, _ := context.WithCancel(context.WithValue(context.Background(), "stack", engineContext))

	var stack Stack

	//actionGit := pipeline.NewGitAction("https://gitee.com/mohaijiang/spring-boot-example.git", "master", ctx)
	//actionEnv := pipeline.NewDockerEnv("maven:3.5-jdk-8", ctx)
	//actionShell := pipeline.NewShellAction("mvn clean package -Dmaven.test.skip=true", ctx)

	// 1： 执行中 2：执行失败， 3： 执行成功
	status := 1

	for key, stage := range job.Stages {
		fmt.Println("stage : ", key)
		for _, step := range stage.Steps {
			if step.RunsOn != "" {
				action := pipeline.NewDockerEnv(step.RunsOn, ctx)
				err := action.Pre()
				if err != nil {
					status = 2
					fmt.Println(err)
					break
				}
				stack.push(action)
				err = action.Hook()
				if err != nil {
					status = 2
					break
				}
			}
			if step.Uses == "" {
				action := pipeline.NewShellAction(step.Run, ctx)
				err := action.Pre()
				if err != nil {
					status = 2
					fmt.Println(err)
					break
				}
				stack.push(action)
				err = action.Hook()
				if err != nil {
					status = 2
					break
				}
			}
			if step.Uses == "git-checkout" {
				action := pipeline.NewGitAction(step.With["url"], step.With["branch"], ctx)
				err := action.Pre()
				if err != nil {
					status = 2
					fmt.Println(err)
					break
				}
				stack.push(action)
				err = action.Hook()
				if err != nil {
					status = 2
					break
				}
			}

			for !stack.isEmpty() {
				action, _ := stack.pop()
				_ = action.Post()
			}
		}
	}

	fmt.Println("status: ", status)
	_ = os.RemoveAll(engineContext["hamsterRoot"].(string))
}

func getJob() *model.Job {
	yamlFile, err := os.ReadFile("./cicd.yml")
	if err != nil {
		panic(err)
	}
	var job model.Job

	err = yaml.Unmarshal(yamlFile, &job)

	if err != nil {
		fmt.Println(err.Error())
	}
	return &job
}
