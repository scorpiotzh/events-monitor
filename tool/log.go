package tool

import (
	"github.com/scorpiotzh/mylog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func GetLog(name string, level int) *mylog.Logger {
	return mylog.NewLoggerDefault(name, level, GetLogWrite())
}

func GetLogWrite() *lumberjack.Logger {
	fileOut := &lumberjack.Logger{
		Filename:   "./logs/out.log", // log path
		MaxSize:    100,              // log file size, M
		MaxBackups: 30,               // backups num
		MaxAge:     7,                // log save days
		LocalTime:  true,
		Compress:   false,
	}
	return fileOut
}
