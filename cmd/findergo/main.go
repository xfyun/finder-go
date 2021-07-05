package main

import (
	"flag"
	"fmt"
	"git.iflytek.com/AIaaS/finder-go/v3/finderm"
	"io"
	"os"
)

var(
	companion = ""
	project = ""
	group =""
	service = ""
	version = ""
	filename = ""
	output = ""
)

func init(){
	flag.StringVar(&project,"p",project,"project")
	flag.StringVar(&companion,"url",project,"companion url")
	flag.StringVar(&group,"g",group,"group")
	flag.StringVar(&service,"s",service,"service")
	flag.StringVar(&version,"v",version,"version")
	flag.StringVar(&filename,"f",filename,"filename")
	flag.StringVar(&output,"o",output,"output file name")

}

func main(){
	flag.Parse()
	finderm.Init(companion,"x")
	data ,err := finderm.GetFile(project,group,service,version,filename)
	if err != nil{
		fatal("get file error:",err)
	}
	var out io.Writer

	switch output {
	case "":
		out,err = os.Create(filename)
	case "stdout":
		out = os.Stdout
	default:
		out,err = os.Create(output)
	}
	if err != nil{
		fatal("create output file error:",err)
	}
	out.Write(data)

}



func fatal(args ...interface{}){
	fmt.Println(args...)
	os.Exit(1)
}
