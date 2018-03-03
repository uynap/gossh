package gossh_test

import (
	"github.com/uynap/gossh"
	"testing"
//	"fmt"
)

func TestGoSSH(t *testing.T) {
	batchJob := gossh.LoadJobs("/home/d388118/go/test/ssh/scp.bp")
	batchJob.Run(5)
}

