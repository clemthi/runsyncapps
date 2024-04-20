package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	ps "github.com/keybase/go-ps"
)

type AppConfig struct {
	Path                string `json:"path"`
	UseExistingInstance bool   `json:"useExistingInstance"`
	KillOnExit          bool   `json:"killOnExit"`
}

type ConfigFile struct {
	WaitCheck    int         `json:"waitCheck"`
	WaitExit     int         `json:"waitExit"`
	Applications []AppConfig `json:"applications"`
}

type ProcessDetails struct {
	path       string
	pid        int
	killOnExit bool
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)

	slog.Debug("Loading flags")
	configFile := *flag.String("config", "config.json", "path to a configuration file")
	verbose := *flag.Bool("verbose", false, "show all logs")

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	config, err := loadConfigFile(configFile)
	if err != nil {
		exitWithMsg(fmt.Sprintf("Cannot load config file %s : %s", configFile, err))
	}

	runningProcs, err := startProcesses(config.Applications)
	if err != nil {
		exitWithMsg(fmt.Sprintf("Error launching apps : %s", err))
	}

	time.Sleep(time.Duration(config.WaitCheck) * time.Second)
	checkRunningProcesses(runningProcs)

	time.Sleep(time.Duration(config.WaitExit) * time.Second)
	killProcesses(runningProcs)
	slog.Debug("Exiting")
}

func loadConfigFile(configFile string) (*ConfigFile, error) {
	slog.Debug(fmt.Sprintf("Loading config file %s", configFile))
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var jsonData ConfigFile
	err = json.Unmarshal([]byte(file), &jsonData)

	return &jsonData, err
}

func startProcesses(apps []AppConfig) (map[int]ProcessDetails, error) {
	procs := make(map[int]ProcessDetails)

	runningProcs, err := ps.Processes()
	if err != nil {
		exitWithMsg(fmt.Sprintf("Error listing processed (%s)", err))
	}

	for _, app := range apps {
		newProc := ProcessDetails{path: app.Path, pid: -1, killOnExit: app.KillOnExit}

		if app.UseExistingInstance {
			// check existing processes
			currentPid, err := findPidFromPath(app.Path, runningProcs)
			if err != nil {
				slog.Warn(fmt.Sprintf("Error checking running processes : %s", err))
			}
			if currentPid != -1 {
				newProc.pid = currentPid
				slog.Info(fmt.Sprintf("Found running app %s [PID : %d]\n", newProc.path, newProc.pid))
			}
		}

		// start app if not found in existing processes
		if newProc.pid == -1 {
			cmd := exec.Command(app.Path)
			err := cmd.Start()
			if err != nil {
				return nil, err
			}
			newProc.pid = cmd.Process.Pid
			slog.Info(fmt.Sprintf("Starting apps %s [PID : %d]", newProc.path, newProc.pid))
		}
		procs[newProc.pid] = newProc
	}

	return procs, nil
}

func exitWithMsg(msg string) {
	slog.Error(msg)
	os.Exit(1)
}

func findPidFromPath(path string, procs []ps.Process) (int, error) {
	for _, proc := range procs {
		if procPath, _ := proc.Path(); procPath == path {
			slog.Debug(fmt.Sprintf("Found running app %s with PID %d", path, proc.Pid()))
			return proc.Pid(), nil
		}
	}
	return -1, nil
}

// check if a process is running
// it uses github.com/keybase/go-ps instead of os.FindProcess
// as the latter always returns something on Windows
func isProcessRunning(pid int) bool {
	process, err := ps.FindProcess(pid)
	return process != nil && err == nil
}

func checkRunningProcesses(procs map[int]ProcessDetails) {
	chanProcesses := make(chan int)
	for pid := range procs {
		go checkRunningProcess(pid, chanProcesses)
	}
	closedProcess := <-chanProcesses
	slog.Info(fmt.Sprintf("Process closed %s [PID: %d]", procs[closedProcess].path, closedProcess))
}

func checkRunningProcess(pid int, processes chan int) {
	process, err := os.FindProcess(pid)
	if err != nil {
		processes <- pid
	}

	// Wait the process to exit
	slog.Debug(fmt.Sprintf("Waiting process [PID: %d] to exit", pid))
	processState, err := process.Wait()
	if err != nil {
		slog.Warn(fmt.Sprintf("Error while waiting process [PID: %d]", pid))
		processes <- pid
		return
	}

	if processState.Exited() {
		slog.Info(fmt.Sprintf("Process [PID: %d] exited with code %d", pid, processState.ExitCode()))
		processes <- pid
		return
	}

	// something went wrong (?), let's assume process is over
	slog.Warn(fmt.Sprintf("Process [PID: %d] exited but : %s ", pid, processState))
	processes <- pid
}

func killProcesses(procs map[int]ProcessDetails) {
	slog.Info("Killing other apps")
	for _, proc := range procs {
		if proc.killOnExit {
			slog.Debug(fmt.Sprintf("Killing process %s [PID: %d]", proc.path, proc.pid))
			procKilled, err := killProcess(proc.pid)
			if err != nil {
				slog.Warn(fmt.Sprintf("Error when killing process %s [PID: %d] : %s", proc.path, proc.pid, err))
			}
			if procKilled {
				slog.Info(fmt.Sprintf("Killed process %s [PID: %d]", proc.path, proc.pid))
			}
		} else {
			slog.Debug(fmt.Sprintf("Skipping process %s [PID: %d]", proc.path, proc.pid))
		}
	}
}

func killProcess(pid int) (bool, error) {
	if !isProcessRunning(pid) {
		return false, nil
	}

	process, _ := os.FindProcess(pid)
	err := process.Kill()
	if err != nil {
		slog.Warn(fmt.Sprintf("Cannot kill process %d", pid))
		return false, err
	}
	return true, nil
}
