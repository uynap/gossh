package gossh

import (
	"bufio"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"time"

	"gopkg.in/olivere/elastic.v3"
)

type ESTask struct {
	TaskMetaData
	TaskJudge
	subTask []TaskDesc
	cmd     string
	timeout time.Duration
}

type Sip struct {
	_id string

	// From log header
	Aparty string `json:"aparty"`
	Bparty string `json:"bparty"`
	Time   string `json:"time"`

	// From SIP body
	Protocol     string `json:"protocol"`
	Direction    string `json:"direction"`
	Sip_type     string `json:"sip_type"`
	Sip_code     string `json:"sip_code"`
	From         string `json:"from"`
	To           string `json:"to"`
	Via          string `json:"via"`
	Call_id      string `json:"call-id"`
	User_agent   string `json:"user-agent"`
	Content_type string `json:"content-type"`

	/*
		Accept              string `json:"accept"`
		Accept_language     string `json:"accept-language"`
		Alert_info          string `json:"alert-info"`
		Allow               string `json:"allow"`
		Allow_events        string `json:"allow-events"`
		Authorization       string `json:"authorization"`
		Call_id             string `json:"call-id"`
		Call_info           string `json:"call-info"`
		Cisco_guid          string `json:"cisco-guid"`
		Contact             string `json:"contact"`
		Content_disposition string `json:"content-disposition"`
		Content_length      string `json:"content-length"`
		Content_type        string `json:"content-type"`
		Cseq                string `json:"cseq"`
		Date                string `json:"date"`
		Diversion           string `json:"diversion"`
		Event               string `json:"event"`
		Expires             string `json:"expires"`
		From                string `json:"from"`
		History_info        string `json:"history-info"`
		Max_forwards        string `json:"max-forwards"`
		Mime_version        string `json:"mime-version"`
		Min_se              string `json:"min-se"`
		P_asserted_identity string `json:"p-asserted-identity"`
		P_charging_vector   string `json:"p-charging-vector"`
		P_early_media       string `json:"p-early-media"`
		P_location          string `json:"p-location"`
		P_rtp_stat          string `json:"p-rtp-stat"`
		P_station_name      string `json:"p-station-name"`
		Privacy             string `json:"privacy"`
		Proxy_require       string `json:"proxy-require"`
		Rack                string `json:"rack"`
		Reason              string `json:"reason"`
		Record_route        string `json:"record-route"`
		Recv_info           string `json:"recv-info"`
		Refer_to            string `json:"refer-to"`
		Referred_by         string `json:"referred-by"`
		Remote_party_id     string `json:"remote-party-id"`
		Require             string `json:"require"`
		Route               string `json:"route"`
		Rseq                string `json:"rseq"`
		Sema_dpm_target     string `json:"sema-dpm-target"`
		Server              string `json:"server"`
		Session             string `json:"session"`
		Session_expires     string `json:"session-expires"`
		Subscription_state  string `json:"subscription-state"`
		Supported           string `json:"supported"`
		Timestamp           string `json:"timestamp"`
		To                  string `json:"to"`
		Total               string `json:"total"`
		User_agent          string `json:"user-agent"`
		User_to_user        string `json:"user-to-user"`
		Via                 string `json:"via"`
		Www_authenticate    string `json:"www-authenticate"`
	*/
}

func (t *ESTask) Init(tdesc TaskDesc) Task {
	return &ESTask{
		TaskMetaData: TaskMetaData{
			id: tdesc.ID,
		},
		cmd:     tdesc.Cmd,
		subTask: tdesc.Task,
		timeout: time.Duration(tdesc.Timeout) * time.Second,
	}
}

func (t *ESTask) Timeout() time.Duration {
	return t.timeout
}

func (t *ESTask) Exec(res chan TaskResult, session *ssh.Session) {
	taskResult := TaskResult{ID: t.ID()}
	defer func() { res <- taskResult }()

	r, err := session.StdoutPipe()
	if err != nil {
		taskResult.Err = err
		return
	}

	if err := session.Start(t.cmd); err != nil {
		taskResult.Err = err
		return
	}

	gzr, err := gzip.NewReader(r)
	if err != nil {
		taskResult.Err = err
		return
	}

	// Create an elastic client
	client, err := elastic.NewClient()
	if err != nil {
		taskResult.Err = err
		fmt.Println("break at elastic.newclient()")
		return
	}
	p, err := client.BulkProcessor().Name("MyBackgroundWorker-1").Workers(3).Do()
	if err != nil {
		taskResult.Err = err
		fmt.Println("break at client.bulkprocessor()")
		return
	}

	println("start reading...")
	br := bufio.NewReader(gzr)
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}

		sip := Sip{}
		if err := json.Unmarshal(line, &sip); err != nil {
			taskResult.Err = err
			return
		}

		id := sip.Time + sip.Aparty + sip.Bparty
		sip._id = fmt.Sprintf("%x", md5.Sum([]byte(id)))
		//		fmt.Println("date:", sip.Time)
		//		continue
		r := elastic.NewBulkIndexRequest().Index("xslog").Type("sip").Doc(sip)
		//		r := elastic.NewBulkUpdateRequest().Index("xslog").Type("sip").Upsert(sip)
		if err != nil {
			taskResult.Err = err
			return
		}
		p.Add(r)
		/*
			_, err = client.Index().
				Index("xslog").
				Type("sip").
				BodyJson(sip).
				Refresh(true).
				Do()
			if err != nil {
				panic(err)
			}
		*/
	}
	err = p.Flush()
	if err != nil {
		taskResult.Err = err
		return
	}

	if err := session.Wait(); err != nil {
		taskResult.Err = err
		return
	}
	return
}

func (t *ESTask) SubTask() []TaskDesc {
	return t.subTask
}
