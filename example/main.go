package main

import (
	"fmt"

	"github.com/uynap/gossh"
	_ "github.com/uynap/gossh/cmdworker"
)

func init() {
	/*
		cmd.Evaluator(func(result task.TaskResult) error {
			return nil
		})
	*/
}

func main() {
	jobs := gossh.LoadBP("./test.bp")
	resCh := jobs.Run()
	for res := range resCh {
		if res.Err != nil {
			fmt.Println("Error:", res.Stderr)
		}
		fmt.Println(res.Id)
		fmt.Println(res.Output)
	}
}
