package main

import (
	"fmt"
	"log"

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
			log.Println("task error:", res.Err)
			fmt.Println("Error:", res.Stderr)
		}
		fmt.Println(res.Id)
		fmt.Println(res.Output)
	}
}
