package gossh_test

import (
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
			io.WriteString(s, "helloworld\n")
		})

		log.Fatal(ssh.ListenAndServe("127.0.0.1:2222", nil))
	}()

	resultCh := gossh.LoadBP("example/test.bp").Run()
	for result := range resultCh {
		if result.Err != nil {
			t.Error(result.Err)
		}
	}

}
