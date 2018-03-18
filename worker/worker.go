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

	NewWorker(task.TaskDesc) Worker

	Timeout() time.Duration

	Evaluate(task.TaskResult) error
}
