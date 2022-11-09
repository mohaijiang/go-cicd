package pipeline

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const STACK = "stack"

type DockerEnv struct {
	ctx         context.Context
	Image       string
	containerId string
}

func NewDockerEnv(image string, ctx context.Context) *DockerEnv {
	return &DockerEnv{
		ctx:   ctx,
		Image: image,
	}
}

func (e *DockerEnv) Pre() error {

	stack := e.ctx.Value(STACK).(map[string]interface{})

	data, ok := stack["workdir"]

	var workdir string
	if ok {
		workdir = data.(string)
	} else {
		return errors.New("workdir error")
	}

	workdirTmp := workdir + "@tmp"

	_ = os.MkdirAll(workdirTmp, os.ModePerm)

	commands := []string{"docker", "run", "-u", "501:20", "-t", "-d", "-v", workdir + ":" + workdir, "-v", workdirTmp + ":" + workdirTmp, "-w", workdir, e.Image, "cat"}
	fmt.Println(strings.Join(commands, " "))
	c := exec.Command(commands[0], commands[1:]...)
	output, err := c.CombinedOutput()
	if err != nil {
		return err
	}
	containerId := string(output)
	fmt.Println(containerId)

	e.containerId = strings.Fields(containerId)[0]
	return err
}

func (e *DockerEnv) Hook() error {

	c := exec.Command("docker", "top", e.containerId, "-eo", "pid,comm")
	output, err := c.CombinedOutput()
	fmt.Println(string(output))
	if err != nil {
		return err
	}

	stack := e.ctx.Value(STACK).(map[string]interface{})
	stack["withEnv"] = []string{"docker", "exec", e.containerId}
	return nil
}

func (e *DockerEnv) Post() error {

	c := exec.Command("docker", "stop", "--time=1", e.containerId)
	_, err := c.CombinedOutput()
	if err != nil {
		return err
	}

	c = exec.Command("docker", "rm", "-f", e.containerId)
	_, err = c.CombinedOutput()

	stack := e.ctx.Value(STACK).(map[string]interface{})
	stack["withEnv"] = []string{}

	if err != nil {
		return err
	}
	return nil
}
