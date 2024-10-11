package constants

import "flag"

const (
	FlagName_ConfigFile = "config"
)

var FlagValue_ConfigFile *string = nil

func InitFlags() {
	//解析命令行参数
	FlagValue_ConfigFile = flag.String(FlagName_ConfigFile, "", FlagName_ConfigFile)
	flag.Parse()
}
