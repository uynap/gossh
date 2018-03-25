package worker

import (
	"time"

	"github.com/uynap/gossh/task"
	"golang.org/x/crypto/ssh"
)

type Worker interface {
	// Exec() is the heart of a worker. The main task of Exec() is to create a
	// "task.TaskResult" struct and send it to the "chan task.TaskResult". Use
	// "*ssh.Session" to handle the ssh connection.
	Exec(chan task.TaskResult, *ssh.Session)

	// SubTask() returns the sub-tasks if there's any.
	SubTask() []task.TaskDesc

	// NewWorker() intakes ONE "task.TaskDesc" and create a valid specific worker
	// which implements the Worker interface.
	NewWorker(task.TaskDesc) Worker

	// Timeout() returns the time duration for the task timeout.
	Timeout() time.Duration

	// Evaluate() is used for evaluating a taske.TaskResult's success/failure.
	Evaluate(task.TaskResult) error
}
