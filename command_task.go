package gossh

/*
type CmdTask struct {
	TaskMetaData
	TaskJudge
	subTask []TaskDesc
	cmd     string
}

func (t *CmdTask) Exec(session *ssh.Session) TaskResult {
	out, err := session.CombinedOutput(t.cmd)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	res := TaskResult{
		output: string(out),
	}

	return res
}

func (t *CmdTask) SubTask() []TaskDesc {
	return t.subTask
}
*/
