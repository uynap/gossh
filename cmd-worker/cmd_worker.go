package cmd

import (
	"fmt"
	"io/ioutil"
	"time"

	"gossh"
	"gossh/task"
	"gossh/worker"

	"golang.org/x/crypto/ssh"
)

type CmdTask struct {
	//	TaskMetaData
	//	TaskJudge
	task    task.TaskDesc
	subTask []task.TaskDesc
	cmd     string
	timeout time.Duration
}

func init() {
	fmt.Println("get init")
	gossh.Register("CmdTask", &CmdTask{})
}

func (t *CmdTask) InitWorker(tdesc task.TaskDesc) worker.Worker {
	return &CmdTask{
		/*
			TaskMetaData: TaskMetaData{
				id: tdesc.Id,
			},
		*/
		task:    tdesc,
		cmd:     tdesc.Cmd,
		subTask: tdesc.Tasks,
		timeout: time.Duration(tdesc.Timeout) * time.Second,
	}
}

func (t *CmdTask) Exec(res chan task.TaskResult, session *ssh.Session) {
	taskResult := task.TaskResult{Id: t.task.Id}
	defer func() { res <- taskResult }()

	r, err := session.StdoutPipe()
	if err != nil {
		taskResult.Err = err
		return
	}

	if err := session.Start(t.cmd); err != nil {
		taskResult.Err = err
		return
	}

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		taskResult.Err = err
		return
	}

	taskResult.Stdout = string(buf)
	/*
		br := bufio.NewReader(r)
		for {
			line, _, err := br.ReadLine()
			if err != nil {
				break
			}

			fmt.Println("oneLine:", string(line))
		}
	*/

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
