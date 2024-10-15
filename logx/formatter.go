package logx

import (
	"bytes"
	"fmt"

	"github.com/0829why/svrframe/constants"

	"github.com/sirupsen/logrus"
)

const (
	red    = 4
	yellow = 6
	blue   = 1
	gray   = 8
	green  = 2
)

type formatter struct {
}

func (m *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timeformat := entry.Time.Format(constants.TimeFormatString)

	_runtime := ""
	_pid := pid

	if entry.Context != nil {
		//_pid = entry.Context.Value("pid").(int)
		_runtime = entry.Context.Value(RunTime).(string)
		if len(_runtime) <= 0 {
			_runtime = "unknow"
		}
	}

	newLog := fmt.Sprintf("[%s] [%s][%d] [%s] [pid:%d] %s\n",
		entry.Level,
		timeformat,
		entry.Time.UnixMilli(),
		_runtime,
		_pid,
		entry.Message)

	// var levelColor int
	// switch entry.Level {
	// case logrus.DebugLevel, logrus.TraceLevel:
	// 	levelColor = gray
	// case logrus.WarnLevel:
	// 	levelColor = yellow
	// case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
	// 	levelColor = red
	// case logrus.InfoLevel:
	// 	levelColor = green
	// default:
	// 	levelColor = green
	// }
	// colorPrint(newLog, levelColor)

	b.WriteString(newLog)

	return b.Bytes(), nil
}
