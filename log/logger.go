package log

import (
	"log"
	"os"
)

var Log Logger
type Logger interface {

	Infof(fmt string, v ...interface{})
	Debugf(fmt string, v ...interface{})
	Errorf(fmt string, v ...interface{})
	Printf(fmt string, v ...interface{})
}

func init(){
	logFile,err:=os.OpenFile("findergo.log",os.O_CREATE|os.O_WRONLY|os.O_APPEND,0666)
	if err!=nil{
		log.Fatalln("打开日志文件失败：",err)
	}
	log.SetPrefix("【findergo】")
	log.SetOutput(logFile)

}
type DefaultLogger struct {
}

func NewDefaultLogger() Logger {
	return &DefaultLogger{}
}

func (l *DefaultLogger) Info(v ...interface{}) {
	log.Println(v)
}

func (l *DefaultLogger) Debug(v ...interface{}) {
	log.Println(v)
}
func (l *DefaultLogger) Printf(fmt string, v ...interface{}) {
	log.Printf(fmt, v)
}
func (l *DefaultLogger) Error(v ...interface{}) {
	log.Println(v)
}

func (l *DefaultLogger) Infof(fmt string, v ...interface{}) {
	log.Printf(fmt, v)
}

func (l *DefaultLogger) Debugf(fmt string, v ...interface{}) {
	log.Printf(fmt, v)
}

func (l *DefaultLogger) Errorf(fmt string, v ...interface{}) {
	log.Printf(fmt, v)
}
