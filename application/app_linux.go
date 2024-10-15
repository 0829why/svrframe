//go:build linux
// +build linux

package application

import (
	"fmt"
	"os"
	"syscall"

	"github.com/0829why/svrframe/constants"
)

// func writePid() bool {
// 	// var f *os.File
// 	// var err1 error
// 	// path := "./pid/"
// 	// is, _exists := isDir(path)
// 	// if _exists && !is {
// 	// 	fmt.Println("pid不是文件夹")
// 	// 	return false
// 	// }
// 	// if !_exists {
// 	// 	err := os.MkdirAll(path, os.ModePerm)
// 	// 	if err != nil {
// 	// 		fmt.Printf("%v\n", err)
// 	// 		return false
// 	// 	}
// 	// }
// 	// fpath := fmt.Sprintf("%spid-%s-%s-%d", path, constants.ProjectName, constants.Service_Type, config.GetServiceInfo().ServiceID)
// 	// if checkFileIsExist(fpath) { //如果文件存在
// 	// 	os.Remove(fpath)
// 	// }
// 	// f, err1 = os.Create(fpath) //创建文件
// 	// check(err1)
// 	// pidinfo := fmt.Sprintf("%d", os.Getpid())
// 	// f.WriteString(pidinfo)

// 	return true
// }

func notifySubKill() {
	if constants.IsWindowsSystem() {
		return
	}
	//通知所有子进程,退出
	if len(sub_pids) > 0 {
		for _, pid := range sub_pids {
			p, err := os.FindProcess(pid)
			if err == nil && p != nil {
				err = p.Signal(syscall.SIGINT)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
func setTitle(title string) {}
