package gossh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/uynap/gossh/task"
	"github.com/uynap/gossh/worker"
	"golang.org/x/crypto/ssh"
)

var (
	workersMu sync.RWMutex
	workers   = make(map[string]worker.Worker)
)

func Register(name string, worker worker.Worker) {
	workersMu.Lock()
	defer workersMu.Unlock()

	if worker == nil {
		panic("gossh: Register worker is nil")
	}

	if _, dup := workers[name]; dup {
		panic("gossh: Register called twice for driver " + name)
	}

	workers[name] = worker
}

func decodeJSON(str string) ([]task.JobDesc, error) {
	var jobs []task.JobDesc
	var bytes []byte

	if _, err := os.Stat(str); err == nil {
	    // str is a filename
		bytes, err = ioutil.ReadFile(str)
		if err != nil {
			println("ioutil err:", str)
			return nil, err
		}
	}else{
	    // str is a JSON string
	    bytes = []byte(str)
	}

	if err := json.Unmarshal(bytes, &jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

type Epic []task.JobDesc

func LoadBP(file interface{}) *Epic {
	var epic Epic

	switch file := file.(type) {
	default:
		panic("LoadBP: only support file path or []JobDesc")
	case []task.JobDesc:
		epic = file
	case string:
		jobs, err := decodeJSON(file)
		if err != nil {
			panic(err)
		}

		// Add ID for each task
		for i, job := range jobs {
			for j, _ := range job.Tasks {
				jobs[i].Tasks[j].Id = strconv.Itoa(i+1) + "-" + strconv.Itoa(j+1)
			}
		}

		epic = jobs
	}

	return &epic
}

func (epic *Epic) Run() <-chan task.TaskResult {
	done := make(chan struct{})
	defer close(done)

	outs := make([]<-chan task.TaskResult, len(*epic))
	for i, job := range *epic {
		host := HostInfo{
			Host: job.Host,
			Port: job.Port,
		}
		account := AccountInfo{
			User: job.User,
			Pass: job.Pass,
		}
		err := host.ConnectAs(account, time.Duration(job.Timeout)*time.Second)
		if err != nil {
			out := make(chan task.TaskResult, 1)
			out <- task.TaskResult{Err: err}
			outs[i] = out
			close(out)
			continue
		}

		in := generator(job.Tasks)
		out := host.DoTasks(in, job.Concurrent)

		outs[i] = out
	}
	return merge(outs)
}

type HostInfo struct {
	Host   string
	Port   string
	Handle *ssh.Client // SSH handler
}

type AccountInfo struct {
	User string
	Pass string
}

func (h *HostInfo) ConnectAs(acc AccountInfo, timeout time.Duration) error {
	config := &ssh.ClientConfig{
		User: acc.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(acc.Pass),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: timeout,
	}

	addr := net.JoinHostPort(h.Host, h.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}

	// Todo: verify the host's identity
	h.Handle = client
	return nil
}

func generator(tasks []task.TaskDesc) chan worker.Worker {
	out := make(chan worker.Worker, 1)
	go func() {
		defer close(out)
		for _, tdesc := range tasks {
			if t, ok := workers[tdesc.Type]; ok {
				t := t.NewWorker(tdesc)
				out <- t
			} else {
				log.Fatal("gossh: Task type is not supported: " + tdesc.Type)
				continue
			}
		}
	}()
	return out
}

func (h *HostInfo) DoTasks(upstream chan worker.Worker, num int) <-chan task.TaskResult {
	outs := make([]<-chan task.TaskResult, num)
	for i := 0; i < num; i++ {
		outs[i] = h.DoTask(upstream)
	}
	return merge(outs)
}

func (h *HostInfo) DoTask(upstream chan worker.Worker) <-chan task.TaskResult {
	out := make(chan task.TaskResult)
	go func() {
		defer close(out)

		for worker := range upstream {
			session, err := h.Handle.NewSession()
			if err != nil {
				panic(err)
			}

			// Excecute one Task
			resCh := make(chan task.TaskResult)
			go worker.Exec(resCh, session)

			var result task.TaskResult
			select {
			case result = <-resCh:
				result.Output = result.Stdout + result.Stderr
				result.Err = worker.Evaluate(result)

				out <- result
			case <-time.After(worker.Timeout()):
				fmt.Println("timeout cancelling...")
				return
			}

			// Excecute Sub-tasks if any
			if worker.SubTask() != nil {
				if result.Err != nil {
					continue
				}

				subIn := generator(worker.SubTask())
				subOut := h.DoTasks(subIn, 1)
				for res := range subOut {
					out <- res
				}
			}
		}
	}()
	return out
}

func merge(cs []<-chan task.TaskResult) <-chan task.TaskResult {
	var wg sync.WaitGroup
	out := make(chan task.TaskResult, 10)

	output := func(c <-chan task.TaskResult) {
		defer wg.Done()
		for n := range c {
			out <- n
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
