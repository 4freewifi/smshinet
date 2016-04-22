package main

import (
	"flag"
	"github.com/4freewifi/smshinet"
	"github.com/golang/glog"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

type Config struct {
	Addr     string
	Username string
	Password string
}

// echo RPC for testing
type Echo int

type EchoArgs struct {
	In string `json:"in"`
}

func (t Echo) Echo(r *http.Request, args *EchoArgs, ret *string) error {
	*ret = args.In
	return nil
}

type SMSHiNet struct {
	config *Config
	client *smshinet.Client
}

type TextMsgArgs struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

type SendMsgRet struct {
	MsgId string
}

func (t *SMSHiNet) SendTextSMS(r *http.Request, args *TextMsgArgs,
	ret *SendMsgRet) error {
	s, err := t.client.SendTextInUTF8Now(args.Recipient, args.Message)
	if err != nil {
		return err
	}
	*ret = SendMsgRet{MsgId: s}
	return nil
}

func main() {
	conffile := flag.String("conf", "config.yaml", "Config YAML file")
	srvaddr := flag.String("addr", "localhost:3059", "Host address to listen on")
	flag.Parse()
	b, err := ioutil.ReadFile(*conffile)
	if err != nil {
		glog.Fatalf("Error reading config file %s: %s, \nhint: check config.yaml.sample",
			conffile, err.Error())
	}
	conf := Config{}
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		glog.Fatalf("Error parsing config file %s: %s",
			conffile, err.Error())
	}
	srv := rpc.NewServer()
	srv.RegisterCodec(json2.NewCodec(), "application/json")
	err = srv.RegisterService(new(Echo), "")
	if err != nil {
		glog.Fatal(err)
	}
	if !srv.HasMethod("Echo.Echo") {
		glog.Fatal("Cannot find required JSON-RPC method: Echo.Echo")
	}
	sms := SMSHiNet{
		config: &conf,
		client: &smshinet.Client{Addr: conf.Addr},
	}
	err = srv.RegisterService(&sms, "")
	if err != nil {
		glog.Fatal(err)
	}
	if !srv.HasMethod("SMSHiNet.SendTextSMS") {
		glog.Fatal("Cannot find required JSON-RPC method: SMGP.Submit")
	}
	// mount to /jsonrpc
	http.Handle("/jsonrpc", srv)
	// connect and authenticate to HiNet "Socket to Air" server
	err = sms.client.Dial()
	if err != nil {
		glog.Fatalf("Error connecting to HiNet server: %s", err.Error())
	}
	defer sms.client.Close()
	glog.Info("Connected to HiNet server")
	err = sms.client.Auth(conf.Username, conf.Password)
	if err != nil {
		glog.Fatalf("Authentication error: %s", err.Error())
	}
	glog.Info("Authenticated")

	glog.Infof("Listening on %s", *srvaddr)
	glog.Fatal(http.ListenAndServe(*srvaddr, nil))
}
