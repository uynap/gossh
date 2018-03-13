package worker

import (
	"time"

	"golang.org/x/crypto/ssh"
)

type Worker interface {
	// Task ID
	//	ID() string

	// Run Task
	Exec(chan TaskResult, *ssh.Session)

	// Sub tasks
	SubTask() []TaskDesc

	Init(TaskDesc) Task

	Timeout() time.Duration
}
