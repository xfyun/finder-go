package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"git.iflytek.com/AIaaS/finder-go/v3/finderm"
	"io"
	"os"
)

var (
	companion = ""
	project   = ""
	group     = ""
	service   = ""
	version   = ""
	filename  = ""
	output    = ""
)

func init() {
	flag.StringVar(&project, "p", project, "project")
	flag.StringVar(&companion, "url", project, "companion url")
	flag.StringVar(&group, "g", group, "group")
	flag.StringVar(&service, "s", service, "service")
	flag.StringVar(&version, "v", version, "version")
	flag.StringVar(&filename, "f", filename, "filename")
	flag.StringVar(&output, "o", output, "output file name")

}

func main() {
	flag.Parse()
	finderm.Init(companion, "x")
	cmd := flag.Arg(0)
	switch cmd {
	case "":
		getConfig()
	case "subscribe-service":
		subscribeService()
	case "listen-service":
		listenService()
	case "listen-config":
		listenConfig()
	default:
		getConfig()
	}

}

func fatal(args ...interface{}) {
	fmt.Println(args...)
	os.Exit(1)
}


func getConfig(){
	data, err := finderm.GetFile(project, group, service, version, filename)
	if err != nil {
		fatal("get file error:", err)
	}
	var out io.Writer

	switch output {
	case "":
		out, err = os.Create(filename)
	case "stdout":
		out = os.Stdout
	default:
		out, err = os.Create(output)
	}
	if err != nil {
		fatal("create output file error:", err)
	}
	out.Write(data)
}

func subscribeService(){
	if version == ""{
		version = "1.0.0"
	}
	addrs,err := finderm.SubscribeService(project,group,"my",service,version)
	if err != nil{
		fatal("subscribe service error:",err)
	}
	outputJson(map[string]interface{}{
		"addresses":addrs,
	})
}

func listenService(){
	if version == ""{
		version = "1.0.0"
	}
	addrs,err := finderm.ListenService(project,group,service,version,0)
	if err != nil{
		fatal("subscribe service error",err)
	}
	outputJson(map[string]interface{}{
		"addresses":addrs,
	})

}

func outputJson(v interface{}){
	bs ,_:= json.Marshal(v)
	fmt.Println(string(bs))
}

func listenConfig()  {

	data ,err := finderm.ListenFile(project,group,service,version,filename,0)
	if err != nil {
		fatal("get file error:", err)
	}
	var out io.Writer

	switch output {
	case "":
		out, err = os.Create(filename)
	case "stdout":
		out = os.Stdout
	default:
		out, err = os.Create(output)
	}
	if err != nil {
		fatal("create output file error:", err)
	}
	out.Write(data)

}
