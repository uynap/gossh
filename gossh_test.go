package gossh_test

import (
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/uynap/gossh"
	_ "github.com/uynap/gossh/cmdworker"
)

func TestGossh(t *testing.T) {
	go func() {
		ssh.Handle(func(s ssh.Session) {
			fmt.Println(s.Command())
			io.WriteString(s, fmt.Sprintf("hello %s\n", s.User()))
		})

		log.Println("starting ssh server on port 2222...")
		log.Fatal(ssh.ListenAndServe(":2222", nil))
	}()

	resultCh := gossh.LoadBP("./test.bp").Run()
	for result := range resultCh {
		if result.Err != nil {
			t.Error(result.Err)
		}
	}

}
