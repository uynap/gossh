package cmd

import (
	"io/ioutil"
	"time"

	"golang.org/x/crypto/ssh"

	"gossh"
)

type CmdWorker struct {
	//	subTask []TaskDesc
	subTask  []interface{}
	cmd      string
	timeout  time.Duration
	evaluate func(*TaskResult) error
}

func init() {
	gossh.Register("CmdWorker", &CmdWorker{})
}

func (t *CmdWorker) Init(tdesc TaskDesc) Task {
	return &CmdWorker{
		TaskMetaData: TaskMetaData{
			id: tdesc.Id,
		},
		cmd:     tdesc.Cmd,
		subTask: tdesc.Task,
		timeout: time.Duration(tdesc.Timeout) * time.Second,
	}
}

func (t *CmdWorker) Timeout() time.Duration {
	return t.timeout
}

// Exec() will be run in a separated goroutine.
func (t *CmdWorker) Exec(res chan TaskResult, session *ssh.Session) {
	taskResult := TaskResult{Id: t.ID()}
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

func (t *CmdWorker) SubTask() []TaskDesc {
	return t.subTask
}