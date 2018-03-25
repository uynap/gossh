# gossh
gossh is a SSH automation library.
It takes a task list(a file in JSON/TOML) and run the tasks in parallel.

## Features
* gossh can control the concurrency of the multiple tasks 
in an elegant way and collect the results. 
* gossh is a generic interface. It must be used in conjunction with a(or multiple) type of worker.
you can provide your own wokers to do a customised ssh-based work.
* Use a Business Process description file to describe the tasks.
* Support sub-tasks.

## Usage
There are many kinds of tasks can be done through SSH. For example, 
run a command to collect result, run a command to make a change to the remote system 
or even download a file through SSH.

`github.com/uynap/gossh/cmdworker` is the worker which is used for running a command remotely, especially for collecting the output.

```go
package main

import (
    "github.com/uynap/gossh"
    _ "github.com/uynap/gossh/cmdworker"
)

func main() {
    // LoadBP() loads tasks from a JSON string, a JSON file or a []JobDesc(`github.com/uynap/gossh/task`)
    // Run() returns a channel of TaskResult(`github.com/uynap/gossh/task`)
    resultCh := gossh.LoadBP("./test.bp").Run()
    for result := range resultCh {
        if result.Err != nil {
            println(result.Stderr)
        }

        println(result.Id)
        println(result.Stdout)
    }
}
```

A \*.bp file looks like: 
```
[
    {
        "host" : "127.0.0.1",
        "port" : "22",
        "user" : "USER",
        "pass" : "PASS",
        "concurrent" : 1,
        "timeout": 5,
        "tasks" : [
            {
                "cmd" : "uname -a",
                "type" : "CmdTask",
                "timeout" : 3,
            },
            {
                "cmd" : "uptime",
                "type" : "CmdTask",
                "timeout" : 3,
            }
        ]
    }
]
```

## Installation
`$ go get github.com/uynap/batRun`

## How to write your own worker
A worker is an interface defined at github.com/uynap/gossh/worker. Have a look 
at "github.com/uynap/gossh/cmdworker" as an example to show how to write a worker.

## License

This project is licensed under the GNU GENERAL PUBLIC LICENSE.

License can be found [here](LICENSE).
