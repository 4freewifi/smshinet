package main

import (
	"github.com/4freewifi/smshinet"
	"github.com/golang/glog"
	"net/http"
	"time"
)

type SMSHiNet struct {
	config *Config
	pool   *ResourcePool
}

type TextMsgArgs struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

type SendMsgRet struct {
	MsgId string
}

const (
	RETRY_INTERVAL = 10
)

func (t *SMSHiNet) reconnect(id int, c *smshinet.Client) error {
	glog.Info("Reconnect")
	c.Close()
	err := c.DialAndAuth(t.config.Username, t.config.Password)
	if err != nil {
		glog.Warningf("Reconnect error: %s", err.Error())
	}
	return err
}

func (t *SMSHiNet) keepReconnect(id int, c *smshinet.Client) {
	for {
		err := t.reconnect(id, c)
		if err == nil {
			break
		}
		glog.Infof("Retry after %d seconds.", RETRY_INTERVAL)
		time.Sleep(RETRY_INTERVAL * time.Second)
	}
	t.pool.Put(id, c)
}

func (t *SMSHiNet) SendTextSMS(r *http.Request, args *TextMsgArgs,
	ret *SendMsgRet) (err error) {
	// 1 second timeout
	id, v, err := t.pool.GetWithTimeout(time.Second)
	if err != nil {
		return
	}
	c := v.(*smshinet.Client)

	// try this twice
	i := 0
	var msgId string
	for {
		msgId, err = c.SendTextInUTF8Now(args.Recipient, args.Message)
		if err == nil || i > 1 {
			break
		}
		glog.Warningf("SendTextSMS error: %s", err.Error())
		i++
		t.reconnect(id, c) // ignore error
	}
	if i > 1 {
		go t.keepReconnect(id, c)
		return err
	}
	t.pool.Put(id, v)
	*ret = SendMsgRet{MsgId: msgId}
	return nil
}

func (t *SMSHiNet) Initialize(size int) error {
	clients := make([]interface{}, size)
	defer func() {
		for _, v := range clients {
			if v == nil {
				continue
			}
			c := v.(*smshinet.Client)
			c.Close()
		}
	}()
	for i := 0; i < size; i++ {
		c := &smshinet.Client{Addr: t.config.Addr}
		err := c.DialAndAuth(t.config.Username, t.config.Password)
		if err != nil {
			return err
		}
		clients[i] = c
	}
	err := t.pool.Initialize(clients)
	if err != nil {
		return err
	}
	glog.Infof("Initialized with %d connections.", size)
	return nil
}
