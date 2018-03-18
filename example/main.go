package main

import (
	"fmt"
	"gossh"
	_ "gossh/cmd-worker"
	"log"
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
