package smshinet

import (
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"testing"
	"time"
)

type config struct {
	Addr       string
	Username   string
	Password   string
	Mobile     string
	IntlMobile string
}

func waitStatus(t *testing.T, c *Client, msgId string) {
Loop:
	for {
		err := c.CheckTextStatus(msgId)
		switch err {
		case nil:
			break Loop
		case CheckRetCode[1], CheckRetCode[2], CheckRetCode[4], CheckRetCode[19]:
			glog.Infof("Got error `%s', wait 10 seconds before retrying.", err.Error())
			time.Sleep(10 * time.Second)
			continue
		default:
			t.Fatal(err)
		}
	}
	return
}

func TestAll(t *testing.T) {
	conf := config{}
	b, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		t.Fatal("Require `config.yaml'. Check `config.yaml.sample'.")
	}
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		t.Fatal(err)
	}
	c := Client{
		Addr: conf.Addr,
	}
	err = c.Dial()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	err = c.Auth(conf.Username, conf.Password)
	if err != nil {
		t.Fatal(err)
	}
	msgId, err := c.SendTextInUTF8Now(conf.Mobile, "smshinet 中文 UTF-8 測試")
	if err != nil {
		t.Fatal(err)
	}
	waitStatus(t, &c, msgId)
	msgId, err = c.SendIntlTextInUTF8Now(conf.IntlMobile,
		"smshinet 中文 國際 UTF-8 測試")
	if err != nil {
		t.Fatal(err)
	}
	waitStatus(t, &c, msgId)
}
