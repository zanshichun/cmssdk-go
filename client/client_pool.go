package client

import (
	"errors"
	"github.com/silenceper/pool"
	"time"
)

var errClientConvert = errors.New("client convert err")

type ClientPool struct {
	pool.Pool
}

func NewClientPool(s *SocketInfo, cap, maxIdle int) (*ClientPool, error) {
	if cap <= 0 {
		return nil, errors.New("client pool cap must greater than zero")
	}

	factory := func() (interface{}, error) {
		c := &Client{info: s}
		err := c.initConn()
		if err != nil {
			return nil, err
		}
		go c.output()
		return c, nil
	}

	close := func(v interface{}) error {
		client := v.(*Client)
		return client.conn.Close()
	}
	if maxIdle <= 0 {
		maxIdle = cap
	}
	//创建一个连接池： 初始化5，最大空闲连接是20，最大并发连接30
	poolConfig := &pool.Config{
		InitialCap: cap,     //资源池初始连接数
		MaxIdle:    maxIdle, //最大空闲连接数
		MaxCap:     cap,     //最大并发连接数
		Factory:    factory,
		Close:      close,
		//Ping:       ping,
		//连接最大空闲时间，超过该时间的连接 将会关闭，可避免空闲时连接EOF，自动失效的问题
		IdleTimeout: 300 * time.Second,
	}
	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil, err
	}
	return &ClientPool{p}, nil
}

func (cp *ClientPool) GetClient() (*Client, error) {
	c, err := cp.Get()
	if err != nil {
		return nil, err
	}
	client, ok := c.(*Client)
	if !ok {
		return nil, errClientConvert
	}
	return client, nil
}

func (cp *ClientPool) PutClient(c *Client) error {
	return cp.Put(c)
}
