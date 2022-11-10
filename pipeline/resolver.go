package pipeline

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"state-example/model"
)

func GetJob(path string) (*model.Job, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return GetJobFromReader(file)
}

func GetJobFromReader(reader io.Reader) (*model.Job, error) {
	yamlFile, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var job model.Job

	err = yaml.Unmarshal(yamlFile, &job)

	return &job, err
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
