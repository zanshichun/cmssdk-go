package main

import (
	"cmssdk-go/async"
	"cmssdk-go/client"
	"cmssdk-go/model"
	"log"
	"testing"
	"time"
)

func client_test(t *testing.T) {
	s := &client.SocketInfo{Addr: "localhost:13232"}
	cp, _ := client.NewClientPool(s, 10, 10)
	c, _ := cp.GetClient()
	r := &model.ReqMessage{
		Event:   2,
		StrId:   "dfsa",
		SetType: "dsf",
		DevId:   "dfsaf",
		DealId:  "dfsaf",
		Ta:      1,
		CrId:    "dfsa",
	}

	r1 := &model.ReqMessage{
		Event:   0,
		StrId:   "dfsa",
		SetType: "dsf",
		DevId:   "dfsaf",
		DealId:  "dfsaf",
		Ta:      1,
		CrId:    "dfsa",
	}
	id, p := async.Feedback()
	c.SendMessage(id, r1)
	c.SendMessage(id, r)

	time.Sleep(3 * time.Second)
	//re, err, t := p.GetOrTimeout(1000000)
	//p, err = async.GetPromise(id)
	//log.Println(err)
	//log.Println(t)
	//log.Println(re)
	//cp.PutClient(c)
}
