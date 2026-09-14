//go:build !windows

package main

import (
	"os/exec"
	"strconv"
	"strings"
)

// hideProcessWindow is a no-op on non-Windows platforms.
func hideProcessWindow(cmd *exec.Cmd) {}

// zaehleGo2rtc zaehlt laufende go2rtc-Prozesse.
func zaehleGo2rtc() int {
	out, err := exec.Command("pgrep", "-fc", "go2rtc").Output()
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return n
}

// killAlleGo2rtc beendet alle go2rtc-Prozesse.
func killAlleGo2rtc() (int, error) {
	n := zaehleGo2rtc()
	if n == 0 {
		return 0, nil
	}
	_ = exec.Command("pkill", "-9", "-f", "go2rtc").Run()
	return n, nil
}

// beendeGo2rtcBaum beendet den go2rtc-Prozess. Auf Nicht-Windows-Systemen (nur
// für Tests) genuegt der einfache Kill.
func beendeGo2rtcBaum(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
}
