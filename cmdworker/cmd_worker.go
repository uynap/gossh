package cmd

import (
	"io/ioutil"
	"time"

	"github.com/uynap/gossh"
	"github.com/uynap/gossh/task"
	"github.com/uynap/gossh/worker"

	"golang.org/x/crypto/ssh"
)

type CmdTask struct {
	taskDesc task.TaskDesc
	subTask  []task.TaskDesc
	cmd      string
	timeout  time.Duration
	stdout   string
	option   interface{}
}

var (
	evaluator = func(result task.TaskResult) error {
		return result.Err
	}
)

func init() {
	gossh.Register("CmdTask", &CmdTask{})
}

func (t *CmdTask) NewWorker(tdesc task.TaskDesc) worker.Worker {
	return &CmdTask{
		taskDesc: tdesc,
		cmd:      tdesc.Cmd,
		subTask:  tdesc.Tasks,
		timeout:  time.Duration(tdesc.Timeout) * time.Second,
	}
}

func Evaluator(f func(task.TaskResult) error) {
	evaluator = f
}

func (t *CmdTask) Evaluate(result task.TaskResult) error {
	return evaluator(result)
}

func (t *CmdTask) Exec(res chan task.TaskResult, session *ssh.Session) {
	taskResult := task.TaskResult{TaskDesc: t.taskDesc}
	defer func() { res <- taskResult }()

	rOut, err := session.StdoutPipe()
	if err != nil {
		taskResult.Err = err
		return
	}

	rErr, err := session.StderrPipe()
	if err != nil {
		taskResult.Err = err
		return
	}

	if err := session.Start(t.cmd); err != nil {
		taskResult.Err = err
		return
	}

	// Read stdout
	buf, err := ioutil.ReadAll(rOut)
	if err != nil {
		taskResult.Err = err
		return
	}

	taskResult.Stdout = string(buf)

	// Read stderr
	buf, err = ioutil.ReadAll(rErr)
	if err != nil {
		taskResult.Err = err
		return
	}

	taskResult.Stderr = string(buf)

	// Wait for the command
	if err := session.Wait(); err != nil {
		taskResult.Err = err
		return
	}

	return
}

func (t *CmdTask) SubTask() []task.TaskDesc {
	return t.subTask
}

func (t *CmdTask) Timeout() time.Duration {
	return t.timeout
}
