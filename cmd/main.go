package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	i "github.com/clemthi/runsyncapps/internal"
	"github.com/getlantern/systray"
)

const (
	traceFile string = "trace.log"
)

func main() {
	configFile := flag.String("config", "config.json", "path to a configuration file")
	enableLog := flag.Bool("log", false, "enable logging")
	flag.Parse()

	// Init default log handler
	var logHandler slog.Handler
	if *enableLog {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		f, _ := os.Create(addTimeSuffix(traceFile))
		defer f.Close()
		logHandler = slog.NewTextHandler(f, nil)
	} else {
		logHandler = slog.NewTextHandler(io.Discard, nil)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Handle interrupt
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	// Init systray icon
	go systray.Run(i.OnReadyUI, func() { os.Exit(0) })

	if err := run(ctx, *configFile, logHandler); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, configFile string, logHandler slog.Handler) error {
	logger := slog.New(logHandler)

	config, err := i.LoadConfigFile(configFile)
	if err != nil {
		logger.Error("cannot load config file", "error", err)
		return err
	}

	p := i.NewProcessHander(logger)
	runningProcs, err := p.StartProcesses(config.Applications)
	if err != nil {
		logger.Error("cannot start app", "error", err)
		return fmt.Errorf("error launching apps : %w", err)
	}

	time.Sleep(time.Duration(config.WaitCheck) * time.Second)
	p.CheckRunningProcesses(runningProcs)

	time.Sleep(time.Duration(config.WaitExit) * time.Second)
	p.KillProcesses(runningProcs)

	return nil
}

func addTimeSuffix(filePath string) string {
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	name := strings.TrimSuffix(filepath.Base(filePath), ext)

	return filepath.Join(dir, fmt.Sprintf("%s_%s%s", name, time.Now().Format("20060102150405"), ext))
}
