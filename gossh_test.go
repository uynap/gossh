package gossh_test

import (
	"gossh"
	"testing"
	//	"fmt"
)

func TestGoSSH(t *testing.T) {
	batchJob := gossh.LoadJobs("/home/d388118/go/test/ssh/scp.bp")
	batchJob.Run(5)
}
