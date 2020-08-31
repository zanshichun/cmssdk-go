package model

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"
)

const (
	Add = iota
	Check
	Get
)

type SendReqMessage struct {
	Id  string
	Msg *ReqMessage
	//_sendTime time.Time
}

type ReqMessage struct {
	Event   int
	StrId   string
	SetType string
	DevId   string
	DealId  string
	Ta      int
	CrId    string
	Worker  string
	Ip      string
	Name    string
	D       uint
	W       uint
}

func (a *SendReqMessage) Marshal() ([]byte, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (a *SendReqMessage) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, &a)
	return err
}

func (a *ReqMessage) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(time.Now().Format("2006-01-02 15:04:05"))
	buffer.WriteString(",")
	buffer.WriteString(a.StrId)
	buffer.WriteString(",")
	buffer.WriteString(a.SetType)
	buffer.WriteString(",")
	buffer.WriteString(a.DevId)
	buffer.WriteString(",")
	buffer.WriteString(a.DealId)
	buffer.WriteString(",")
	buffer.WriteString(strconv.Itoa(a.Ta))
	buffer.WriteString(",")
	buffer.WriteString(a.CrId)
	buffer.WriteString(",")
	buffer.WriteString(a.Ip)
	buffer.WriteString(",")
	buffer.WriteString(a.Worker)
	return buffer.String()
}

type SendRespMessage struct {
	Id  string
	Msg *RespMessage
	//_sendTime time.Time
}

type RespMessage struct {
	Event   int
	Freq    uint64
	Message string
}

func (r *SendRespMessage) Marshal() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *SendRespMessage) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, &r)
	return err
}
