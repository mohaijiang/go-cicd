package main

import (
	"errors"
	"state-example/pipeline"
)

type Stack []pipeline.ActionHandler

// 入栈
func (s *Stack) push(a pipeline.ActionHandler) {
	*s = append(*s, a)
}

// 出栈
func (s *Stack) pop() (pipeline.ActionHandler, error) {
	if len(*s) == 0 {
		return nil, errors.New("empty stack")
	}
	a := *s
	defer func() {
		*s = a[:len(a)-1]
	}()
	return a[len(a)-1], nil
}

func (s *Stack) isEmpty() bool {
	return len(*s) == 0
}
