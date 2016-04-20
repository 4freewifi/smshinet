package smshinet

import (
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"testing"
	"time"
)

type config struct {
	Addr     string
	Username string
	Password string
	Mobile   string
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
	s := Server{
		Addr: conf.Addr,
	}
	err = s.Dial()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	err = s.Auth(conf.Username, conf.Password)
	if err != nil {
		t.Fatal(err)
	}
	msgId, err := s.SendTextInUTF8(conf.Mobile, "smshinet test")
	if err != nil {
		t.Fatal(err)
	}
Loop:
	for {
		err = s.CheckTextStatus(msgId)
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
