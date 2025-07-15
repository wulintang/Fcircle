package utils

import (
	"io"
	"log"
	"os"
	"sync"
)

var (
	Logger *log.Logger
	once   sync.Once
)

func InitLog(logFile string) error {
	var err error
	once.Do(func() {
		f, e := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if e != nil {
			err = e
			return
		}
		// io.MultiWriter 可以同时写入文件和控制台
		mw := io.MultiWriter(f, os.Stdout)
		Logger = log.New(mw, "", log.LstdFlags)
	})
	return err
}

func Info(v ...interface{}) {
	if Logger != nil {
		Logger.Println(v...)
	}
}

func Error(v ...interface{}) {
	if Logger != nil {
		args := make([]interface{}, 0, 1+len(v))
		args = append(args, "[ERROR]")
		args = append(args, v...)
		Logger.Println(args...)
	}
}

func Infof(format string, v ...interface{}) {
	if Logger != nil {
		Logger.Printf(format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if Logger != nil {
		Logger.Printf("[ERROR] "+format, v...)
	}
}
