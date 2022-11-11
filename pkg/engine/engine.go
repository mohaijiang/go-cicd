package engine

import (
	"context"
	"fmt"
	"os"
	"state-example/pkg/logger"
	"state-example/pkg/model"
	"state-example/pkg/pipeline"

	"gopkg.in/yaml.v2"
)

// Engine 流程引擎
func Engine(job *model.Job) {

	engineContext := make(map[string]interface{})
	engineContext["hamsterRoot"] = "/tmp/example"
	engineContext["workdir"] = "/tmp/example"
	engineContext["name"] = job.Name

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), "stack", engineContext))
	defer cancel()

	var stack Stack

	// 1：执行中 2：执行失败，3：执行成功
	status := 1

	stagesList, err := StageSort(job)
	if err != nil {
		cancel()
		logger.Panicf("stage sort error: %v", err)
	}

	for _, stageWapper := range stagesList {
		logger.Infof("current stage : %s", stageWapper.Name)
		for _, step := range stageWapper.Stage.Steps {
			if step.RunsOn != "" {
				action := pipeline.NewDockerEnv(step.RunsOn, ctx)
				err := action.Pre()
				if err != nil {
					status = 2
					logger.Errorf("docker env pre error: %v", err)
					break
				}
				stack.push(action)
				err = action.Hook()
				if err != nil {
					status = 2
					logger.Errorf("docker env hook error: %v", err)
					break
				}
			}
			if step.Uses == "" {
				action := pipeline.NewShellAction(step.Run, ctx)
				err := action.Pre()
				if err != nil {
					status = 2
					logger.Errorf("shell action pre error: %v", err)
					break
				}
				stack.push(action)
				err = action.Hook()
				if err != nil {
					status = 2
					logger.Errorf("shell action hook error: %v", err)
					break
				}
			}
			if step.Uses == "git-checkout" {
				action := pipeline.NewGitAction(step.With["url"], step.With["branch"], ctx)
				err := action.Pre()
				if err != nil {
					status = 2
					logger.Errorf("git action pre error: %v", err)
					break
				}
				stack.push(action)
				err = action.Hook()
				if err != nil {
					status = 2
					logger.Errorf("git action hook error: %v", err)
					break
				}
			}

			for !stack.isEmpty() {
				action, _ := stack.pop()
				_ = action.Post()
			}
		}
	}

	logger.Infof("job %s status: %d", job.Name, status)
	_ = os.RemoveAll(engineContext["hamsterRoot"].(string))
}

func GetJob(file string) *model.Job {
	logger.Debugf("get job from yaml file: %s", file)

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		logger.Fatalf("yaml config file read error: %v", err)
	}

	var job model.Job
	err = yaml.Unmarshal(yamlFile, &job)
	if err != nil {
		logger.Errorf("yaml config file unmarshal error: %v", err)
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
