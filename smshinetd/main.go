package main

import (
	"flag"
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
	glog.V(1).Infof("Echo %s", args.In)
	return nil
}

func main() {
	conffile := flag.String("conf", "config.yaml", "Config YAML file")
	srvaddr := flag.String("addr", "localhost:3059",
		"Host address to listen on")
	poolsize := flag.Int("pool", 1, "Size of thread pool")
	flag.Parse()
	b, err := ioutil.ReadFile(*conffile)
	if err != nil {
		glog.Fatalf("Error reading config file %s: %s, \n"+
			"hint: check config.yaml.sample",
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
		pool:   &ResourcePool{},
	}
	err = sms.Initialize(*poolsize)
	if err != nil {
		glog.Fatal(err)
	}
	err = srv.RegisterService(&sms, "")
	if err != nil {
		glog.Fatal(err)
	}
	for _, name := range []string{
		"SMSHiNet.SendTextSMS",
		"SMSHiNet.CheckTextStatus"} {
		if srv.HasMethod(name) {
			continue
		}
		glog.Fatalf("Cannot find required JSON-RPC method: %s", name)
	}
	// mount to /jsonrpc
	http.Handle("/jsonrpc", srv)
	glog.Infof("Listening on %s", *srvaddr)
	glog.Fatal(http.ListenAndServe(*srvaddr, nil))
}
