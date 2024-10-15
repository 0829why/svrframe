package logx

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/0829why/svrframe/constants"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

// level ErrorLevel-DebugLevel
func InitLogx() {

	if inited {
		return
	}
	inited = true

	level := DebugLevel
	if !constants.IsDebug() {
		log_level := os.Getenv(constants.Env_LogLevel)
		if len(log_level) > 0 {
			l, e := strconv.Atoi(log_level)
			if e == nil {
				level = logrus.Level(l)
			}
		}
		if level < ErrorLevel || level >= MaxLevel {
			level = InfoLevel
		}
	}

	var mw_info io.Writer = nil

	fullpath = getLogPath()
	if len(fullpath) > 0 {
		fullpath += "/" + constants.Service_Type

		filename = "log.info"
		file_info := fullpath + "/" + filename

		writer_info, _ := rotatelogs.New(
			file_info+".%Y%m%d%H%M",
			rotatelogs.WithLinkName(file_info),
			rotatelogs.WithMaxAge(file_WithMaxAge),
			rotatelogs.WithRotationCount(file_WithRotationCount),
			rotatelogs.WithRotationTime(file_WithRotationTime),
			rotatelogs.WithRotationSize(file_WithRotationSize),
		)
		mw_info = io.MultiWriter(writer_info, os.Stdout)
	} else {
		mw_info = io.MultiWriter(os.Stdout)
	}
	logInfo.SetOutput(mw_info)

	logInfo.SetLevel(level)

	formatter := &formatter{
		//ForceColors:               true,
		//EnvironmentOverrideColors: true,
		//FullTimestamp:   true,
		//TimestampFormat: constants.TimeFormatString,
		//DisableSorting:            true,
		//DisableLevelTruncation:    true,
		//PadLevelText: true,
	}

	logInfo.SetFormatter(formatter)

	go update()
}

func GetTraceFile() string {
	return fullpath + "/trace"
}
func GetLogFullPath() string {
	return fullpath
}
func GetLogFileName() string {
	return filename
}

func GetLoggerWriter() io.Writer {
	return logInfo.Out
}

func TraceBack() {
	if !inited || JustPrint {
		fmt.Println(string(debug.Stack()))
		return
	}
	ctx := getContext(3)
	t := &log_t{
		msg:   string(debug.Stack()),
		ctx:   ctx,
		level: DebugLevel,
	}
	buffer <- t
}

func Debugln(args ...interface{}) {
	if !inited || JustPrint {
		fmt.Println(append([]interface{}{"[error]"}, args...)...)
		return
	}
	ctx := getContext(3)
	t := &log_t{
		msg:   fmt.Sprintln(args...),
		ctx:   ctx,
		level: DebugLevel,
	}
	buffer <- t
}
func DebugF(msg string, args ...interface{}) {
	if !inited || JustPrint {
		fmt.Printf("[debug]"+msg+"\n", args...)
		return
	}
	ctx := getContext(3)
	t := &log_t{
		msg:   fmt.Sprintf(msg, args...),
		ctx:   ctx,
		level: DebugLevel,
	}
	buffer <- t
}

func Infoln(args ...interface{}) {
	if !inited || JustPrint {
		fmt.Println(append([]interface{}{"[error]"}, args...)...)
		return
	}
	ctx := getContext(3)
	t := &log_t{
		msg:   fmt.Sprintln(args...),
		ctx:   ctx,
		level: InfoLevel,
	}
	buffer <- t
}
func InfoF(msg string, args ...interface{}) {
	if !inited || JustPrint {
		fmt.Printf("[info]"+msg+"\n", args...)
		return
	}
	ctx := getContext(3)
	t := &log_t{
		msg:   fmt.Sprintf(msg, args...),
		ctx:   ctx,
		level: InfoLevel,
	}
	buffer <- t
}

func Warnln(args ...interface{}) {
	if !inited || JustPrint {
		fmt.Println(append([]interface{}{"[error]"}, args...)...)
		return
	}
	ctx := getContext(3)
	t := &log_t{
		msg:   fmt.Sprintln(args...),
		ctx:   ctx,
		level: WarnLevel,
	}
	buffer <- t
}
func WarnF(msg string, args ...interface{}) {
	if !inited || JustPrint {
		fmt.Printf("[warn]"+msg+"\n", args...)
		return
	}
	ctx := getContext(3)
	t := &log_t{
		msg:   fmt.Sprintf(msg, args...),
		ctx:   ctx,
		level: WarnLevel,
	}
	buffer <- t
}

func Errorln(args ...interface{}) {
	if !inited || JustPrint {
		fmt.Println(append([]interface{}{"[error]"}, args...)...)
		return
	}
	ctx := getContext(3)
	t := &log_t{
		msg:   fmt.Sprintln(args...),
		ctx:   ctx,
		level: ErrorLevel,
	}
	buffer <- t
}
func ErrorF(msg string, args ...interface{}) {
	if !inited || JustPrint {
		fmt.Printf("[error]"+msg+"\n", args...)
		return
	}
	ctx := getContext(3)
	t := &log_t{
		msg:   fmt.Sprintf(msg, args...),
		ctx:   ctx,
		level: ErrorLevel,
	}
	buffer <- t
}
