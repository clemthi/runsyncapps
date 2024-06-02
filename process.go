package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	ps "github.com/keybase/go-ps"
)

type ProcessDetails struct {
	path       string
	pid        int
	killOnExit bool
}

type ProcessHander struct {
	logger *slog.Logger
}

func NewProcessHander(logger *slog.Logger) *ProcessHander {
	return &ProcessHander{logger: logger}
}

func (p *ProcessHander) StartProcesses(apps []AppConfig) (map[int]ProcessDetails, error) {
	procs := make(map[int]ProcessDetails)

	runningProcs, err := ps.Processes()
	if err != nil {
		panic(fmt.Sprintf("Error listing processed (%s)", err))
	}

	for _, app := range apps {
		newProc := ProcessDetails{path: app.Path, pid: -1, killOnExit: app.KillOnExit}

		if app.UseExistingInstance {
			// check existing processes
			currentPid, err := p.findPidFromPath(app.Path, runningProcs)
			if err != nil {
				p.logger.Warn(fmt.Sprintf("Error checking running processes : %s", err))
			}
			if currentPid != -1 {
				newProc.pid = currentPid
				p.logger.Info(fmt.Sprintf("Found running app %s [PID : %d]\n", newProc.path, newProc.pid))
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
			p.logger.Info(fmt.Sprintf("Starting apps %s [PID : %d]", newProc.path, newProc.pid))
		}
		procs[newProc.pid] = newProc
	}

	return procs, nil
}

func (p *ProcessHander) findPidFromPath(path string, procs []ps.Process) (int, error) {
	for _, proc := range procs {
		if procPath, _ := proc.Path(); procPath == path {
			p.logger.Debug(fmt.Sprintf("Found running app %s with PID %d", path, proc.Pid()))
			return proc.Pid(), nil
		}
	}
	return -1, nil
}

// check if a process is running
// it uses github.com/keybase/go-ps instead of os.FindProcess
// as the latter always returns something on Windows
func (p *ProcessHander) isProcessRunning(pid int) bool {
	process, err := ps.FindProcess(pid)
	return process != nil && err == nil
}

func (p *ProcessHander) CheckRunningProcesses(procs map[int]ProcessDetails) {
	chanProcesses := make(chan int)
	for pid := range procs {
		go p.checkRunningProcess(pid, chanProcesses)
	}
	closedProcess := <-chanProcesses
	p.logger.Info(fmt.Sprintf("Process closed %s [PID: %d]", procs[closedProcess].path, closedProcess))
}

func (p *ProcessHander) checkRunningProcess(pid int, processes chan int) {
	process, err := os.FindProcess(pid)
	if err != nil {
		processes <- pid
	}

	// Wait the process to exit
	p.logger.Debug(fmt.Sprintf("Waiting process [PID: %d] to exit", pid))
	processState, err := process.Wait()
	if err != nil {
		p.logger.Warn(fmt.Sprintf("Error while waiting process [PID: %d]", pid))
		processes <- pid
		return
	}

	if processState.Exited() {
		p.logger.Info(fmt.Sprintf("Process [PID: %d] exited with code %d", pid, processState.ExitCode()))
		processes <- pid
		return
	}

	// something went wrong (?), let's assume process is over
	p.logger.Warn(fmt.Sprintf("Process [PID: %d] exited but : %s ", pid, processState))
	processes <- pid
}

func (p *ProcessHander) KillProcesses(procs map[int]ProcessDetails) {
	p.logger.Info("Killing other apps")
	for _, proc := range procs {
		if proc.killOnExit {
			p.logger.Debug(fmt.Sprintf("Killing process %s [PID: %d]", proc.path, proc.pid))
			procKilled, err := p.killProcess(proc.pid)
			if err != nil {
				p.logger.Warn(fmt.Sprintf("Error when killing process %s [PID: %d] : %s", proc.path, proc.pid, err))
			}
			if procKilled {
				p.logger.Info(fmt.Sprintf("Killed process %s [PID: %d]", proc.path, proc.pid))
			}
		} else {
			p.logger.Debug(fmt.Sprintf("Skipping process %s [PID: %d]", proc.path, proc.pid))
		}
	}
}

func (p *ProcessHander) killProcess(pid int) (bool, error) {
	if !p.isProcessRunning(pid) {
		return false, nil
	}

	process, _ := os.FindProcess(pid)
	err := process.Kill()
	if err != nil {
		p.logger.Warn(fmt.Sprintf("Cannot kill process %d", pid))
		return false, err
	}
	return true, nil
}
