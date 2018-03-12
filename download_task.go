package gossh

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

type DownloadTask struct {
	//	TaskMetaData
	//	TaskJudge
	subTask []TaskDesc
	cmd     string
}

func (t *DownloadTask) Exec(session *ssh.Session) TaskResult {
	r, err := session.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	if err := session.Start(t.cmd); err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	tmpfile, err := ioutil.TempFile("/tmp", "py")
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	fmt.Println("start reading...")
	gzr, err := gzip.NewReader(r)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	//	tr := tar.NewReader(gzr)
	//	tr.Next()
	br := bufio.NewReader(gzr)
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		_ = line
		fmt.Println(string(line))
	}
	/*
				for {
					hdr, err := tr.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Fatalln(err)
					}
					fmt.Printf("contents of %s:\n", hdr.Name)
		//			if _, err := io.Copy(os.Stdout, tr); err != nil {
				    if _, err := io.Copy(tmpfile, tr); err != nil {
						log.Fatalln(err)
					}
					fmt.Println("hhhhhhh========")
				}
	*/
	/*
		if _, err := io.Copy(tmpfile, r); err != nil {
			log.Fatalln(err)
		}
	*/

	if err := session.Wait(); err != nil {
		fmt.Println(err)
		os.Remove(tmpfile.Name())
		log.Fatal(err)
	}

	res := TaskResult{
		result: tmpfile.Name(),
	}

	return res
}

func (t *DownloadTask) SubTask() []TaskDesc {
	return t.subTask
}
