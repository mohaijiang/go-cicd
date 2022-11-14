package queue

import (
	model2 "state-example/pkg/model"
)

type IQueue interface {
	Push(job *model2.Job, node *model2.Node)
	Listener() chan *model2.Job
}

type Queue struct {
}
