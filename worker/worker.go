package worker

import (
	"time"

	"gossh/task"

	"golang.org/x/crypto/ssh"
)

type Worker interface {
	// Task ID
	//	ID() string

	// Run Task
	Exec(chan task.TaskResult, *ssh.Session)

	// Sub tasks
	SubTask() []task.TaskDesc

	InitWorker(task.TaskDesc) Worker

	Timeout() time.Duration
}
