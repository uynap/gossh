package task

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
