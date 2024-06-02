package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/getlantern/systray"
)

func main() {
	// Init default log handler
	logHandler := slog.NewTextHandler(io.Discard, nil)

	configFile := *flag.String("config", "config.json", "path to a configuration file")
	enableLog := *flag.Bool("log", false, "enable logging (console by default)")
	logInFile := *flag.Bool("logfile", false, "log events in a trace.log file")

	if enableLog {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		if logInFile {
			f, _ := os.Create("trace.log")
			defer f.Close()
			logHandler = slog.NewTextHandler(f, nil)
		} else {
			logHandler = slog.NewTextHandler(os.Stdout, nil)
		}
	}

	logger := slog.New(logHandler)

	config, err := loadConfigFile(configFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Cannot load config file %s : %s", configFile, err))
		os.Exit(1)
	}

	// Init systray icon
	go systray.Run(onReadyUI, func() { os.Exit(0) })

	p := NewProcessHander(logger)
	runningProcs, err := p.StartProcesses(config.Applications)
	if err != nil {
		logger.Error(fmt.Sprintf("Error launching apps : %s", err))
		os.Exit(1)
	}

	time.Sleep(time.Duration(config.WaitCheck) * time.Second)
	p.CheckRunningProcesses(runningProcs)

	time.Sleep(time.Duration(config.WaitExit) * time.Second)
	p.KillProcesses(runningProcs)
}
