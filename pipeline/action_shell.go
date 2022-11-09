package pipeline

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type ShellAction struct {
	command  string
	filename string
	ctx      context.Context
}

func NewShellAction(command string, ctx context.Context) *ShellAction {

	return &ShellAction{
		command: command,
		ctx:     ctx,
	}
}

func (a *ShellAction) Pre() error {

	stack := a.ctx.Value(STACK).(map[string]interface{})

	data, ok := stack["workdir"]

	var workdir string
	if ok {
		workdir = data.(string)
	} else {
		return errors.New("workdir error")
	}

	workdirTmp := workdir + "@tmp"

	_ = os.MkdirAll(workdirTmp, os.ModePerm)

	a.filename = workdirTmp + "/" + randSeq(10) + ".sh"

	content := []byte("#!/bin/sh\nset -ex\n" + a.command)
	err := os.WriteFile(a.filename, content, os.ModePerm)

	return err
}

func (a *ShellAction) Hook() error {

	stack := a.ctx.Value(STACK).(map[string]interface{})

	workdir, ok := stack["workdir"].(string)
	fmt.Println(stack)

	commands := []string{"sh", "-c", a.filename}
	val, ok := stack["withEnv"]
	if ok {
		precommand := val.([]string)
		shellCommand := make([]string, len(commands))
		copy(shellCommand, commands)
		commands = append([]string{}, precommand...)
		commands = append(commands, shellCommand...)
	}

	// c := exec.CommandContext(ctx, "cmd", "/C", cmd)
	c := exec.CommandContext(a.ctx, commands[0], commands[1:]...) // mac linux
	c.Dir = workdir
	fmt.Println(strings.Join(commands, " "))
	stdout, err := c.StdoutPipe()
	//stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		reader := bufio.NewReader(stdout)
		//errReader := bufio.NewReader(stderr)
		for {
			// 其实这段去掉程序也会正常运行，只是我们就不知道到底什么时候Command被停止了，而且如果我们需要实时给web端展示输出的话，这里可以作为依据 取消展示
			select {
			// 检测到ctx.Done()之后停止读取
			case <-a.ctx.Done():
				if a.ctx.Err() != nil {
					fmt.Printf("程序出现错误: %q", a.ctx.Err())
				} else {
					fmt.Println("程序被终止")
				}
				return
			default:
				//errString, err := errReader.ReadString('\n')
				//if err != nil || err == io.EOF {
				//	return
				//}
				//fmt.Print(errString)
				readString, err := reader.ReadString('\n')
				if err != nil || err == io.EOF {
					return
				}
				fmt.Print(readString)
			}
		}
	}(&wg)
	err = c.Start()
	wg.Wait()
	return err
}

func (a *ShellAction) Post() error {
	//return os.Remove(a.command)
	return nil
}
