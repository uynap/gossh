package gossh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"gossh/worker"
)

/*
var TaskType = map[string]Task{
	"ESTask":  &ESTask{},
	"CmdTask": &CmdTask{},
	//	"DownloadTask": &DownloadTask{},
}
*/

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
}

type JobDesc struct {
	Id         string     `json:"id"`
	Host       string     `json:"host"`
	Port       string     `json:"port"`
	User       string     `json:"user"`
	Pass       string     `json:"pass"`
	Timeout    int        `json:"timeout"`
	Concurrent int        `json:"concurrent"`
	Tasks      []TaskDesc `json:"tasks"`
}

type TaskDesc struct {
	Id         string     `json:"id"`
	Cmd        string     `json:"cmd"`
	Type       string     `json:"type"`
	Timeout    int        `json:"timeout"`
	Concurrent int        `json:"concurrent"`
	Tasks      []TaskDesc `json:"tasks"`
}

func decodeJSON(filename string) ([]JobDesc, error) {
	var jobs []JobDesc

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

var rand uint32
var randmu sync.Mutex

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func NextID() string {
	randmu.Lock()
	r := rand
	if r == 0 {
		r = reseed()
	}

	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	randmu.Unlock()

	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

type BatchJob struct {
	Jobs []JobDesc
}

func LoadBP(file interface{}) *BatchJob {
	batchJob := &BatchJob{}

	switch file := file.(type) {
	default:
		panic("LoadBP: only support file path or []JobDesc")
	case []JobDesc:
		batchJob.Jobs = file
	case string:
		jobs, err := decodeJSON(file)
		if err != nil {
			panic(err)
		}

		// Add Task ID for each job
		for i, job := range jobs {
			for j, _ := range job.Tasks {
				jobs[i].Tasks[j].Id = NextID()
			}
		}

		batchJob.Jobs = jobs
	}

	return batchJob
}

func (bj *BatchJob) Run(concurrent int) <-chan TaskResult {
	done := make(chan struct{})
	defer close(done)

	outs := make([]<-chan TaskResult, len(bj.Jobs))
	for i, job := range bj.Jobs {
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
			out := make(chan TaskResult, 1)
			out <- TaskResult{Id: job.Id, Err: err}
			outs[i] = out
			close(out)
			continue
		}

		in := generator(job.Tasks)
		out := host.DoTasks(in, job.Concurrent)

		outs[i] = out
		/*
			for res := range out {
				fmt.Println("finish one task:", res.output)
			}
		*/
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

type Task interface {
	// Task ID
	//	ID() string

	// Run Task
	Exec(chan TaskResult, *ssh.Session)

	// Sub tasks
	SubTask() []TaskDesc

	Init(TaskDesc) Task

	Timeout() time.Duration
}

type TaskResult struct {
	Id string
	// Standard output
	Stdout string

	// Standard error
	Stderr string

	// 如果任务的输出能够分为标准输出和标准错误输出的话，
	// 这里就是两者的混合，就像是你在显示器上看到的一样；
	// 如果不能够区分的话，那么 output 就是任务的输出
	Output string

	// Task's error if any
	Err error

	// The conclusion of the task
	result string

	// Sub tasks result
	SubTask []TaskResult

	TaskDesc TaskDesc
}

/*
type TaskMetaData struct {
	id string
}

func (tmd *TaskMetaData) ID() string {
	return tmd.id
}
*/

func generator(tasks []TaskDesc) chan Task {
	out := make(chan Task, 1)
	go func() {
		defer close(out)
		for _, tdesc := range tasks {
			if t, ok := TaskType[tdesc.Type]; ok {
				t := t.Init(tdesc)
				//				fmt.Printf("%#v\n", t)
				out <- t
			} else {
				log.Error("gossh: Task type is not supported: " + tdesc.Type)
				continue
			}
		}
	}()
	return out
}

func (h *HostInfo) DoTasks(upstream chan Task, num int) <-chan TaskResult {
	outs := make([]<-chan TaskResult, num)
	for i := 0; i < num; i++ {
		outs[i] = h.DoTask(upstream)
	}
	return merge(outs)
}

func (h *HostInfo) DoTask(upstream chan Task) <-chan TaskResult {
	out := make(chan TaskResult)
	go func() {
		defer close(out)

		for task := range upstream {
			session, err := h.Handle.NewSession()
			if err != nil {
				panic(err)
			}

			// Excecute one Task
			res := make(chan TaskResult)
			go task.Exec(res, session)

			var result TaskResult
			select {
			case result = <-res:
				out <- result
			case <-time.After(task.Timeout()):
				fmt.Println("timeout cancelling...")
				return
			}

			// Excecute Sub-tasks if any
			if task.SubTask() != nil {
				if result.Err != nil {
					continue
				}

				subIn := generator(task.SubTask())
				subOut := h.DoTasks(subIn, 1)
				for res := range subOut {
					out <- res
				}
			}
		}
	}()
	return out
}

func merge(cs []<-chan TaskResult) <-chan TaskResult {
	var wg sync.WaitGroup
	out := make(chan TaskResult, 10)

	output := func(c <-chan TaskResult) {
		defer wg.Done()
		for n := range c {
			out <- n
			/*
				select {
				case out <- n:
				case <-done:
					fmt.Println("merge: done is called")
					return
				}
			*/
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
