package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
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
	configFile := flag.String("config", "config.json", "path to a configuration file")
	flag.Parse()

	config, err := loadConfigFile(*configFile)
	if err != nil {
		log.Fatalf("Cannot load config file %s : %s", *configFile, err)
	}

	runningProcs, err := startProcesses(config.Applications, true)
	if err != nil {
		log.Fatalf("Error launching apps : %s", err)
	}

	time.Sleep(time.Duration(config.WaitCheck) * time.Second)
	checkRunningProcesses(runningProcs)

	time.Sleep(time.Duration(config.WaitExit) * time.Second)
	killProcesses(runningProcs)
	log.Printf("Done")
}

func loadConfigFile(configFile string) (ConfigFile, error) {
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error opening config %s : %s", configFile, err)
	}

	var jsonData ConfigFile
	err = json.Unmarshal([]byte(file), &jsonData)

	return jsonData, err
}

func startProcesses(apps []AppConfig, checkCurrentProcesses bool) (map[int]ProcessDetails, error) {
	procs := make(map[int]ProcessDetails)

	runningProcs, err := ps.Processes()
	if err != nil {
		log.Fatalf("Error listing processed (%s)", err)
	}

	for _, app := range apps {
		newProc := ProcessDetails{path: app.Path, pid: -1, killOnExit: app.KillOnExit}

		if app.UseExistingInstance {
			currentPid, err := findPidFromPath(app.Path, runningProcs)
			if err != nil {
				log.Fatalf("Error checking running processes : %s", err)
			}
			if currentPid != -1 {
				newProc.pid = currentPid
				log.Printf("Found running app %s [PID : %d]\n", newProc.path, newProc.pid)
			}
		}

		if newProc.pid == -1 {
			cmd := exec.Command(app.Path)
			err := cmd.Start()
			if err != nil {
				return nil, err
			}
			newProc.pid = cmd.Process.Pid
			log.Printf("Starting apps %s [PID : %d]\n", newProc.path, newProc.pid)
		}
		procs[newProc.pid] = newProc
	}

	return procs, nil
}

func findPidFromPath(path string, procs []ps.Process) (int, error) {
	for _, proc := range procs {
		if procPath, _ := proc.Path(); procPath == path {
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
	log.Printf("Process closed %s [PID: %d]", procs[closedProcess].path, closedProcess)
}

func checkRunningProcess(pid int, processes chan int) {
	process, err := os.FindProcess(pid)
	if err != nil {
		processes <- pid
	}
	processState, err := process.Wait()
	if err != nil {
		processes <- pid
	}
	if processState.Exited() {
		processes <- pid
	}
}

func killProcesses(procs map[int]ProcessDetails) {
	for _, proc := range procs {
		if proc.killOnExit {
			procKilled, err := killProcess(proc.pid)
			if err != nil {
				log.Printf("Error when killing process%s [PID: %d] : %s", proc.path, proc.pid, err)
			}
			if procKilled {
				log.Printf("Killed process %s [PID: %d]", proc.path, proc.pid)
			}
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
		log.Printf("Cannot kill process %d", pid)
		return false, err
	}
	return true, nil
}
