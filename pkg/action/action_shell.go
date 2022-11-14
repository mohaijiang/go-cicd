package action

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"state-example/pkg/utils"
	"strings"
	"syscall"
)

// ShellAction 命令工作
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

	a.filename = workdirTmp + "/" + utils.RandSeq(10) + ".sh"

	content := []byte("#!/bin/sh\nset -ex\n" + a.command)
	err := os.WriteFile(a.filename, content, os.ModePerm)

	return err
}

func (a *ShellAction) Hook() error {

	stack := a.ctx.Value(STACK).(map[string]interface{})

	workdir, ok := stack["workdir"].(string)
	if !ok {
		return errors.New("workdir is empty")
	}
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
	if err != nil {
		return err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		for {
			// 其实这段去掉程序也会正常运行，只是我们就不知道到底什么时候 Command 被停止了，而且如果我们需要实时给 web 端展示输出的话，这里可以作为依据 取消展示
			// 检测到 ctx.Done() 之后停止读取
			<-a.ctx.Done()
			if a.ctx.Err() != nil {
				fmt.Printf("程序出现错误: %q", a.ctx.Err())
			} else {
				p := c.Process
				if p == nil {
					return
				}
				// Kill by negative PID to kill the process group, which includes
				// the top-level process we spawned as well as any subprocesses
				// it spawned.
				_ = syscall.Kill(-p.Pid, syscall.SIGKILL)
				fmt.Println("程序被终止")
			}
		}
	}()

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stdoutScanner.Scan() {
			fmt.Println(stdoutScanner.Text())
		}
	}()
	go func() {
		for stderrScanner.Scan() {
			fmt.Println(stderrScanner.Text())
		}
	}()

	err = c.Start()
	if err != nil {
		fmt.Println("command start error: ", err)
	}

	err = c.Wait()
	if err != nil {
		fmt.Println("command wait error: ", err)
	}
	return err
}

func (a *ShellAction) Post() error {
	return os.Remove(a.command)
}
