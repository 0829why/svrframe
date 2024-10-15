package logx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/0829why/svrframe/constants"

	"github.com/sirupsen/logrus"
)

const (
	RunTime constants.ContextKey = "runtime"
)

type log_t struct {
	msg   string
	ctx   context.Context
	level logrus.Level
}

var (
	logInfo  *logrus.Logger
	fullpath string
	filename string
	inited   bool
	buffer   chan *log_t
)

var (
	pid int
)

var JustPrint bool = false

const (
	ErrorLevel logrus.Level = iota + logrus.ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	MaxLevel
)
const (
	//文件保留2周
	file_WithMaxAge = time.Duration(time.Hour * 24 * 7 * 2)
	//文件一天一切换
	file_WithRotationTime = time.Duration(time.Hour * 24)
	//
	file_WithRotationCount = 0
	//
	file_WithRotationSize = -1
	//消息队列长度
	msg_queue_max_len = 4096
)

func init() {
	logInfo = logrus.New()
	fullpath = ""
	filename = "log.info"
	inited = false
	pid = os.Getpid()

	JustPrint = false

	initKernel32()

	buffer = make(chan *log_t, msg_queue_max_len)
}

func isLogRuntime() bool {
	log_runtime := os.Getenv(constants.Env_LogRuntime)
	log_runtime = strings.ToLower(log_runtime)
	if len(log_runtime) > 0 && log_runtime == "true" {
		return true
	}
	return false
}

func getLogPath() string {
	var _exists bool = false
	path := os.Getenv(constants.Env_LogPath)
	if len(path) <= 0 {
		return ""
		// p, _ := os.Getwd()
		// path = p + "/log/" + constants.ProjectName
	} else {
		if !filepath.IsAbs(path) {
			p, _ := os.Getwd()
			path = p + "/" + path
		}
		var is bool
		is, _exists = constants.IsDir(path)
		if _exists && !is {
			fmt.Println("env log_path must need a dir")
			// panic("env log_path must need a dir")
			return ""
		}
	}
	if !_exists {
		//创建文件夹
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Printf("%+v\n", err)
			// panic(err)
			return ""
		}
	}
	return path
}

func getRuntime(step int) string {

	if !isLogRuntime() {
		return ""
	}
	//function := "?"
	file := "?"
	line := 0

	_, file, line, ok := runtime.Caller(step)
	if !ok {
		//return fmt.Sprintf("%s %s:%d", file, function, line)
		return fmt.Sprintf("%s:%d", file, line)
	}
	//f := runtime.FuncForPC(pc)
	//if f != nil {
	//	function = f.Name()
	//}

	fName := filepath.Base(file)
	//return fmt.Sprintf("%s %s:%d", fName, function, line)
	return fmt.Sprintf("%s:%d", fName, line)
}

func getContext(_call_step int) context.Context {
	ctx := context.WithValue(context.Background(), RunTime, getRuntime(_call_step))
	//ctx = context.WithValue(ctx, "pid", pid)
	//ctx := context.WithValue(context.Background(), "pid", os.Getpid())

	return ctx
}

func log(msg *log_t) {
	if msg.level == InfoLevel {
		colorPrint(msg.msg, green)
		logInfo.WithContext(msg.ctx).Infoln(msg.msg)
	} else if msg.level == WarnLevel {
		colorPrint(msg.msg, yellow)
		logInfo.WithContext(msg.ctx).Warnln(msg.msg)
	} else if msg.level == ErrorLevel {
		colorPrint(msg.msg, red)
		logInfo.WithContext(msg.ctx).Errorln(msg.msg)
	} else {
		colorPrint(msg.msg, gray)
		logInfo.WithContext(msg.ctx).Debugln(msg.msg)
	}
}

func update() {
	defer func() {
		inited = false
		//close(buffer)
		constants.GetServiceStopWaitGroup().Done()
	}()

	exit_ch := constants.GetServiceStopListener().AddListener()
	constants.GetServiceStopWaitGroup().Add(1)

	_exit := false
	for {
		select {
		case <-exit_ch.Done():
			_exit = true
		case msg, ok := <-buffer:
			if !ok || msg == nil {
				colorPrint(msg.msg, red)
				logrus.Errorln("log buffer 关闭")
				_exit = true
			} else {
				log(msg)
			}
		default:
			if _exit {
				return
			}
			time.Sleep(time.Millisecond * 10)
		}
	}
}
