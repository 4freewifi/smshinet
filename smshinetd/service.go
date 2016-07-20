package main

import (
	"github.com/4freewifi/smshinet"
	"github.com/golang/glog"
	"io"
	"net"
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
	var msgId string
	for {
		id, v := t.pool.Get()
		c := v.(*smshinet.Client)
		msgId, err = c.SendTextInUTF8Now(args.Recipient, args.Message)
		if err == nil {
			goto BREAK
		}
		// give up if neither network error nor EOF
		if _, netError := err.(net.Error); !netError && err != io.EOF {
			glog.Errorf("SendTextSMS error: %s", err.Error())
			goto BREAK
		}
		glog.Warningf("SendTextSMS network error: %s", err.Error())
		go t.keepReconnect(id, c)
		continue
	BREAK:
		t.pool.Put(id, v)
		break
	}
	if err == nil {
		*ret = SendMsgRet{MsgId: msgId}
	}
	return
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
