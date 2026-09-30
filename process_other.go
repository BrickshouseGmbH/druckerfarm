//go:build !windows

package main

import (
	"os/exec"
	"strconv"
	"strings"
)

// hideProcessWindow is a no-op on non-Windows platforms.
func hideProcessWindow(cmd *exec.Cmd) {}

// countGo2rtc zaehlt laufende go2rtc-Prozesse.
func countGo2rtc() int {
	out, err := exec.Command("pgrep", "-fc", "go2rtc").Output()
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return n
}

// go2rtcPIDs returns the PIDs of all running go2rtc processes.
func go2rtcPIDs() []int {
	out, err := exec.Command("pgrep", "-f", "go2rtc").Output()
	if err != nil {
		return nil
	}
	var pids []int
	for _, line := range strings.Fields(strings.TrimSpace(string(out))) {
		if n, err := strconv.Atoi(line); err == nil {
			pids = append(pids, n)
		}
	}
	return pids
}

// killAllGo2rtc beendet alle go2rtc-Prozesse.
func killAllGo2rtc() (int, error) {
	n := countGo2rtc()
	if n == 0 {
		return 0, nil
	}
	_ = exec.Command("pkill", "-9", "-f", "go2rtc").Run()
	return n, nil
}

// killGo2rtcTree terminates the go2rtc process. On non-Windows systems (test
// only) a simple kill is enough.
func killGo2rtcTree(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
}
