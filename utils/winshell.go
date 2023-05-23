package utils

import (
	"errors"
	"os"
	"strings"

	ps "github.com/mitchellh/go-ps"
	log "github.com/sirupsen/logrus"
)

func mapProcessList() (map[int]ps.Process, error) {
	list, err := ps.Processes()
	if err != nil {
		return nil, err
	}
	var processMap = make(map[int]ps.Process)
	for _, p := range list {
		processMap[p.Pid()] = p
	}
	return processMap, nil
}

func GetProcess(pid int) (process ps.Process, err error) {
	pMap, err := mapProcessList()
	if err != nil {
		log.Errorf("Error getting process list: %s.\n", err)
	}
	process, ok := pMap[pid]
	if !ok {
		return process, os.ErrNotExist
	} else {
		return process, nil
	}
}

func GetHostShell() (string, error) {
	var shell string
	currentPid := os.Getpid()
	process, err := GetProcess(currentPid)
	if err != nil {
		return "", errors.New("Error getting process: " + err.Error())
	}
	pprocess, err := GetProcess(process.PPid())
	if err != nil {
		return "", errors.New("Error getting parent process: " + err.Error())
	}
	pdesc := pprocess.Executable()
	if strings.Contains(pdesc, "cmd") {
		shell = "cmd"
	} else if strings.Contains(pdesc, "powershell") {
		shell = "powershell"
	} else if strings.Contains(pdesc, "bash") {
		shell = "bash"
	} else if strings.Contains(pdesc, "zsh") {
		shell = "zsh"
	} else if strings.Contains(pdesc, "sh") {
		shell = "sh"
	} else {
		return "", errors.New("Error getting shell: " + pdesc)
	}
	return shell, nil
}
