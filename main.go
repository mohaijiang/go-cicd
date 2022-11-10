package main

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"state-example/action"
	"state-example/model"
	"state-example/pipeline"
)

func main() {
	job := getJob()
	for name := range job.Stages {
		fmt.Println(name)
	}

	engine(job)
}

// 流程引擎
func engine(job *model.Job) {

	engineContext := make(map[string]interface{})
	engineContext["hamsterRoot"] = "/tmp/example"
	engineContext["workdir"] = "/tmp/example"
	engineContext["name"] = job.Name

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), "stack", engineContext))

	var stack Stack

	// 1： 执行中 2：执行失败， 3： 执行成功
	status := 1

	stagesList, err := pipeline.StageSort(job)
	if err != nil {
		panic(err)
	}

	for _, stageWapper := range stagesList {
		fmt.Println("stage : ", stageWapper.Name)
		for _, step := range stageWapper.Stage.Steps {
			if step.RunsOn != "" {
				action := action.NewDockerEnv(step.RunsOn, ctx)
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
				action := action.NewShellAction(step.Run, ctx)
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
				action := action.NewGitAction(step.With["url"], step.With["branch"], ctx)
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
		}
		for !stack.isEmpty() {
			action, _ := stack.pop()
			_ = action.Post()
		}

		if status == 2 {
			stageWapper.Status = 2
			cancel()
			break
		}
		stageWapper.Status = 3
	}

	fmt.Println("status: ", status)
	os.RemoveAll(engineContext["hamsterRoot"].(string))
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
