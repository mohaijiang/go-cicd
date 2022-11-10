package queue

import "state-example/model"

type IQueue interface {
	Push(job *model.Job, node *model.Node)
	Listener() chan *model.Job
}

type Queue struct {
}
