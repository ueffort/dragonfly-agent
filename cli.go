package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/Sirupsen/logrus"
	flag "github.com/ueffort/goutils/mflag"
)

type EnvFlags struct {
	FlagSet   *flag.FlagSet
	PostParse func() error

	Debug    bool
	LogLevel string
}

type CommonFlags struct {
	FlagSet   *flag.FlagSet
	PostParse func() error

	ConfigFileName string
	Discovery      string
	Advertise      string
}

var (
	envFlags    = &EnvFlags{FlagSet: new(flag.FlagSet)}
	commonFlags = &CommonFlags{FlagSet: new(flag.FlagSet)}
	enoughFlag  = false
)

func homepath(p string) string {
	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	}
	return filepath.Join(home, p)
}

func init() {
	envFlags.PostParse = postParseEnv

	cmd := envFlags.FlagSet
	cmd.BoolVar(&envFlags.Debug, []string{"D", "-debug"}, true, "Enable debug mode")
	cmd.StringVar(&envFlags.LogLevel, []string{"l", "-log-level"}, "info", "Set the logging level")

	commonFlags.PostParse = postParseCommon

	cmd = commonFlags.FlagSet
	cmd.StringVar(&commonFlags.ConfigFileName, []string{"-config"}, "config.json", "Location of config file")
	cmd.StringVar(&commonFlags.Discovery, []string{"-discovery"}, "", "Discovery server url scheme.")
	cmd.StringVar(&commonFlags.Advertise, []string{"-advertise"}, "", "Address of the Agent joining the cluster.")
}

func Run() error {
	exit := make(chan bool)
	discovery, err := parseDiscovery(&commonFlags.Discovery)
	if err != nil {
		return err
	}
	advertise, err := parseAdvertise(&commonFlags.Advertise)
	if err != nil {
		return err
	}
	sm, err := StartServer(discovery, advertise)
	if err != nil {
		return err
	}
	closeHandler := func(s os.Signal, arg interface{}) error {
		return StopServer(sm, exit)
	}
	go SignalHandle(closeHandler)
	<-exit
	return nil
}

// 分析discovery
func parseDiscovery(discoveryStr *string) (*Discovery, error) {
	discovery := new(Discovery)
	err := discovery.parse(*discoveryStr)
	if err != nil {
		return nil, err
	}
	return discovery, nil
}

// 分析advertise
func parseAdvertise(advertiseStr *string) (*Advertise, error) {
	advertise := new(Advertise)
	err := advertise.parse(*advertiseStr)
	if err != nil {
		return nil, err
	}
	return advertise, nil
}

func postParseEnv() error {
	if envFlags.LogLevel != "" {
		lvl, err := logrus.ParseLevel(envFlags.LogLevel)
		if err != nil {
			return fmt.Errorf("Unable to parse logging level: %s\n", envFlags.LogLevel)
		}
		logrus.SetLevel(lvl)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if envFlags.Debug {
		os.Setenv("DEBUG", "1")
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}

// 解析基本参数
func postParseCommon() error {
	if commonFlags.ConfigFileName != "" {
		file := ""
		if path.IsAbs(commonFlags.ConfigFileName) {
			file = commonFlags.ConfigFileName
		} else {
			current_path, _ := os.Getwd()
			file = path.Join(current_path, commonFlags.ConfigFileName)
		}
		configFile, err := Load(file)
		if err == nil {
			logger.Debug(fmt.Sprintf("Config info: %s", configFile))
			if configFile.Discovery != "" && commonFlags.Discovery == "" {
				commonFlags.Discovery = configFile.Discovery
			}
			if configFile.Advertise != "" && commonFlags.Advertise == "" {
				commonFlags.Advertise = configFile.Advertise
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("Unable to parse config file: %s\n%s", commonFlags.ConfigFileName, err)
		} else {
			logger.Debugln("Config file is not exist:%s", file)
		}
	}
	if commonFlags.Discovery != "" && commonFlags.Advertise != "" {
		enoughFlag = true
	} else {
		return fmt.Errorf("Required Flag: Discovery: %s, Advertise: %s \n", commonFlags.Discovery, commonFlags.Advertise)
	}
	return nil
}
