package client

import (
	"cmssdk-go/async"
	"cmssdk-go/model"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

type SocketInfo struct {
	Addr string
	Path string
}

const (
	defaultPath = "/"
	scheme      = "ws"
)

var (
	errSocketInfo          = errors.New("socket info error")
	errUninitializedSocket = errors.New("uninitialized socket")
	errMessage             = errors.New("send message is null")
)

type Client struct {
	info *SocketInfo
	conn *websocket.Conn
}

func (c *Client) initConn() error {
	if c.info == nil || c.info.Addr == "" {
		return errSocketInfo
	}
	if c.info.Path == "" {
		c.info.Path = defaultPath
	}
	u := url.URL{
		Scheme: scheme,
		Host:   c.info.Addr,
		Path:   c.info.Path,
	}

	var dialer *websocket.Dialer
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) SendMessage(id string, message *model.ReqMessage) error {
	if message == nil {
		return errMessage
	}
	if c.conn == nil {
		return errUninitializedSocket
	}
	msg := &model.SendReqMessage{
		Id:  id,
		Msg: message,
	}
	b, err := msg.Marshal()
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, b)
}

func (c *Client) output() error {
	if c.conn == nil {
		return errUninitializedSocket
	}
	for {
		msgType, message, err := c.conn.ReadMessage()
		if err != nil {
			if msgType == -1 {
				log.Println("服务端断开链接")
				break
			}
		}
		var msg model.SendRespMessage
		msg.Unmarshal(message)
		err = async.Call(msg)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}
