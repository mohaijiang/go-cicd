package main

import (
	"context"
	"fmt"
	"os"
	"state-example/logger"
	"state-example/model"
	"state-example/pipeline"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func main() {
	logger.Init().StdoutAndFile().Level(logrus.TraceLevel)
	job := getJob()
	for name := range job.Stages {
		logger.Info(name)
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

	// 1：执行中 2：执行失败，3：执行成功
	status := 1

	stagesList, err := StageSort(job)
	if err != nil {
		defer cancel()
		panic(err)
	}

	for _, stageWapper := range stagesList {
		fmt.Println("stage : ", stageWapper.Name)
		for _, step := range stageWapper.Stage.Steps {
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

func StageSort(job *model.Job) ([]model.StageWrapper, error) {
	stages := make(map[string]model.Stage)
	for key, stage := range job.Stages {
		stages[key] = stage
	}

	sortedMap := make(map[string]any)

	stageList := make([]model.StageWrapper, 0)
	for len(stages) > 0 {
		last := len(stages)
		for key, stage := range stages {
			allContains := true
			for _, needs := range stage.Needs {
				_, ok := sortedMap[needs]
				if !ok {
					allContains = false
				}
			}
			if allContains {
				sortedMap[key] = ""
				delete(stages, key)
				stageList = append(stageList, model.NewStageWrapper(key, stage))
			}
		}

		if len(stages) == last {
			return nil, fmt.Errorf("cannot resolve dependency, %v", stages)
		}

	}

	return stageList, nil

}
