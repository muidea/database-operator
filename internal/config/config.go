package config

import (
	"encoding/json"
	fu "github.com/muidea/magicCommon/foundation/util"
	"os"
)

var defaultConfig = `
{
}`

var currentListenPort string
var currentNodePort string
var currentWorkPath string
var configItem *CfgItem

const cfgFile = "/var/app/config/cfg.json"

func init() {
	cfg := &CfgItem{}
	err := fu.LoadConfig(cfgFile, cfg)
	if err != nil {
		_ = json.Unmarshal([]byte(defaultConfig), cfg)
	}

	configItem = cfg

	currentWorkPath, _ = os.Getwd()
}

func SetListenPort(listenPort string) {
	currentListenPort = listenPort
}

func GetListenPort() string {
	return currentListenPort
}

func GetNodePort() string {
	nodePort, ok := os.LookupEnv("NODEPORT")
	if ok {
		return nodePort
	}

	return currentListenPort
}

func GetConfigFile() string {
	return cfgFile
}

func ReloadConfig() *CfgItem {
	cfg := &CfgItem{}
	err := fu.LoadConfig(cfgFile, cfg)
	if err != nil {
		return nil
	}

	configItem = cfg
	return configItem
}

type CfgItem struct {
}
