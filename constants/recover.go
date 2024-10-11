package constants

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

func trace(message string) {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}

	p, _ := os.Getwd()
	fullPath := fmt.Sprintf("%s/log/%s/%s/crash_%d", p, ProjectName, Service_Type, time.Now().UnixMilli())
	filePtr, err := os.Create(fullPath)
	if err == nil {
		defer filePtr.Close()
		io.WriteString(filePtr, str.String())
	}
}

func Recover() func() {
	return func() {
		if err := recover(); err != nil {
			trace(fmt.Sprintf("%+v", err))

			ProgramExit()
		}
	}
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}
