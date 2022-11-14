package action

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path"
	"state-example/pkg/logger"
	"state-example/pkg/stream"
	"strings"
	"syscall"
)

// GitAction git clone
type GitAction struct {
	repository string
	branch     string
	workdir    string
	ctx        context.Context
}

func NewGitAction(repository, branch string, ctx context.Context) *GitAction {

	return &GitAction{
		repository: repository,
		branch:     branch,
		ctx:        ctx,
	}
}

func (a *GitAction) Pre() error {

	//TODO ... 检验git 命令是否存在
	return nil
}

func (a *GitAction) Hook() error {

	stack := a.ctx.Value(STACK).(map[string]interface{})

	pipelineName := stack["name"].(string)

	logger.Infof("git stack: %v", stack)

	hamsterRoot := stack["hamsterRoot"].(string)

	_ = os.MkdirAll(hamsterRoot, os.ModePerm)
	_ = os.Remove(path.Join(hamsterRoot, pipelineName))

	commands := []string{"git", "clone", "--progress", a.repository, "-b", a.branch, pipelineName}
	c := exec.CommandContext(a.ctx, commands[0], commands[1:]...) // mac linux
	c.Dir = hamsterRoot
	logger.Debugf("execute git clone command: %s", strings.Join(commands, " "))

	stdout, err := c.StdoutPipe()
	if err != nil {
		logger.Errorf("stdout error: %v", err)
		return err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		logger.Errorf("stderr error: %v", err)
		return err
	}

	go func() {
		for {
			// 检测到 ctx.Done() 之后停止读取
			<-a.ctx.Done()
			if a.ctx.Err() != nil {
				logger.Errorf("git clone error: %v", a.ctx.Err())
			} else {
				p := c.Process
				if p == nil {
					return
				}
				// Kill by negative PID to kill the process group, which includes
				// the top-level process we spawned as well as any subprocesses
				// it spawned.
				_ = syscall.Kill(-p.Pid, syscall.SIGKILL)
				logger.Info("git clone process killed")
			}
		}
	}()

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stdoutScanner.Scan() {
			stream.OutputCh <- stdoutScanner.Text()
		}
	}()
	go func() {
		for stderrScanner.Scan() {
			stream.OutputCh <- stderrScanner.Text()
		}
	}()

	err = c.Start()
	if err != nil {
		logger.Errorf("git clone error: %v", err)
		return err
	}

	err = c.Wait()
	if err != nil {
		logger.Errorf("git clone error: %v", err)
		return err
	}
	logger.Info("git clone success")

	a.workdir = path.Join(hamsterRoot, pipelineName)
	stack["workdir"] = a.workdir
	return nil
}

func (a *GitAction) Post() error {
	return os.Remove(a.workdir)
}
