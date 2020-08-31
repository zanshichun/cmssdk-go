package async

import (
	"cmssdk-go/model"
	"cmssdk-go/promise"
	"cmssdk-go/sync/container"
	"errors"
	uuid "github.com/satori/go.uuid"
)

var errResultConvert = errors.New("send resp convert err")
var errPromiseGet = errors.New("error get promise")
var table = container.NewConcurrentTable()

func Call(s model.SendRespMessage) error {
	i, ok := table.Get(s.Id)
	if !ok {
		return errPromiseGet
	}
	p := i.(*promise.Promise)
	return p.Resolve(s)
}

func SetPromise(id string, p *promise.Promise) {
	table.Set(id, p)
}

func GetPromise(id string) (*promise.Promise, error) {
	i, ok := table.Get(id)
	if !ok {
		return nil, errPromiseGet
	}
	p := i.(*promise.Promise)
	return p, nil
}

func Feedback() (string, *promise.Promise) {
	p := promise.NewPromise()
	p.OnSuccess(func(v interface{}) {
		resp, ok := v.(model.SendRespMessage)
		if !ok {
			p.Reject(errResultConvert)
		}
		p.Resolve(resp)
	}).OnFailure(func(v interface{}) {
		resp, ok := v.(model.SendRespMessage)
		if !ok {
			p.Reject(errResultConvert)
		}
		table.Remove(resp.Id)
	}).OnComplete(func(v interface{}) {
		resp, ok := v.(model.SendRespMessage)
		if !ok {
			p.Reject(errResultConvert)
		}
		table.Remove(resp.Id)
	})
	id := uuid.NewV4().String()
	SetPromise(id, p)
	return id, p
}
