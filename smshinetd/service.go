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

type MessageID struct {
	MessageID string
}

type TextStatus struct {
	Success bool
	Error   string
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

func errHandle(name string, err error) (stop bool) {
	// give up if neither network error nor EOF
	if _, netError := err.(net.Error); !netError && err != io.EOF {
		glog.Errorf("%s error: %s", name, err.Error())
		return true
	}
	glog.Warningf("%s network error: %s", name, err.Error())
	return false
}

func (t *SMSHiNet) SendTextSMS(r *http.Request, args *TextMsgArgs,
	ret *MessageID) (err error) {
	var msgId string
	for {
		id, v := t.pool.Get()
		c := v.(*smshinet.Client)
		msgId, err = c.SendTextInUTF8NowWithExpire(
			args.Recipient, args.Message, time.Minute)
		if err == nil || errHandle("SendTExtSMS", err) {
			goto BREAK
		}
		go t.keepReconnect(id, c)
		continue
	BREAK:
		t.pool.Put(id, v)
		break
	}
	if err == nil {
		*ret = MessageID{MessageID: msgId}
	}
	return
}

func (t *SMSHiNet) CheckTextStatus(r *http.Request, args *MessageID,
	ret *TextStatus) (err error) {
	for {
		id, v := t.pool.Get()
		c := v.(*smshinet.Client)
		err = c.CheckTextStatus(args.MessageID)
		if err == nil || errHandle("CheckTextStatus", err) {
			goto BREAK
		}
		go t.keepReconnect(id, c)
		continue
	BREAK:
		t.pool.Put(id, v)
		break
	}
	if err == nil {
		*ret = TextStatus{
			Success: true,
			Error:   "",
		}
	} else {
		*ret = TextStatus{
			Success: false,
			Error:   err.Error(),
		}
	}
	return
}

func (t *SMSHiNet) Initialize(size int) (err error) {
	clients := make([]interface{}, size)
	for i := 0; i < size; i++ {
		c := &smshinet.Client{Addr: t.config.Addr}
		err = c.DialAndAuth(t.config.Username, t.config.Password)
		if err != nil {
			goto ERROR_EXIT
		}
		clients[i] = c
	}
	err = t.pool.Initialize(clients)
	if err != nil {
		goto ERROR_EXIT
	}
	glog.Infof("Initialized with %d connections.", size)
	return nil
ERROR_EXIT:
	for _, v := range clients {
		if v == nil {
			continue
		}
		c := v.(*smshinet.Client)
		c.Close()
	}
	return err
}
