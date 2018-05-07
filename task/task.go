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
	Id         string            `json:"id"`
	Cmd        string            `json:"cmd"`
	Type       string            `json:"type"`
	Timeout    int               `json:"timeout"`
	Concurrent int               `json:"concurrent"`
	Tasks      []TaskDesc        `json:"tasks"`
	Option     interface{}       `json:"option"`
	Metadata   map[string]string `json:"metadata"`
}

type TaskResult struct {
	TaskDesc

	// Standard output
	Stdout string

	// Standard error
	Stderr string

	// Output is the combination of Stdout and Stderr
	Output string

	// Task's error if any
	Err error

	// The conclusion of the task
	result string

	// Sub tasks result
	SubTask []TaskResult
}
