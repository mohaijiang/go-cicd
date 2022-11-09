package pipeline

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

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
	return nil
}

func (a *GitAction) Hook() error {

	stack := a.ctx.Value(STACK).(map[string]interface{})

	pipelineName := stack["name"].(string)

	fmt.Println("git stack: ", stack)

	hamsterRoot := stack["hamsterRoot"].(string)

	_ = os.MkdirAll(hamsterRoot, os.ModePerm)
	_ = os.Remove(path.Join(hamsterRoot, pipelineName))

	commands := []string{"git", "clone", a.repository, "-b", a.branch, pipelineName}
	c := exec.CommandContext(a.ctx, commands[0], commands[1:]...) // mac linux
	c.Dir = hamsterRoot
	fmt.Println(strings.Join(commands, " "))

	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	errout, err := c.StderrPipe()
	if err != nil {
		return err
	}

	// var wg sync.WaitGroup
	// wg.Add(1)
	go func() {
		// defer wg.Done()
		reader := bufio.NewReader(stdout)
		errReader := bufio.NewReader(errout)
		for {
			// 其实这段去掉程序也会正常运行，只是我们就不知道到底什么时候 Command 被停止了，而且如果我们需要实时给 web 端展示输出的话，这里可以作为依据 取消展示
			select {
			// 检测到 ctx.Done() 之后停止读取
			case <-a.ctx.Done():
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
				return
			default:
				readString, err := reader.ReadString('\n')
				fmt.Printf("标准输出：%s\n", readString)
				if err != nil || err == io.EOF {
					return
				}

				errString, err := errReader.ReadString('\n')
				fmt.Printf("标准错误：%s\n", errString)
				if err != nil || err == io.EOF {
					return
				}

			}
		}

	}()

	err = c.Start()
	if err != nil {
		fmt.Println("command start error: ", err)
	}

	// wg.Wait()

	err = c.Wait()
	if err != nil {
		fmt.Println("command wait error: ", err)
	}

	if err == nil {
		a.workdir = path.Join(hamsterRoot, pipelineName)
		stack["workdir"] = a.workdir
	}
	return err
}

func (a *GitAction) Post() error {
	return os.Remove(a.workdir)
}
