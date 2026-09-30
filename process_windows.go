//go:build windows

package main

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// hideProcessWindow sets SysProcAttr so the child process gets no console window.
// CREATE_NO_WINDOW (0x08000000) prevents the black terminal from appearing.
func hideProcessWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		HideWindow:    true,
	}
}

// countGo2rtc counts running go2rtc processes via the Windows process list.
func countGo2rtc() int {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq go2rtc.exe", "/NH", "/FO", "CSV")
	hideProcessWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	n := 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(strings.ToLower(line), "go2rtc.exe") {
			n++
		}
	}
	return n
}

// go2rtcPIDs returns the PIDs of ALL running go2rtc processes (not just the
// selbst verwalteten). tasklist gibt CSV: "go2rtc.exe","1234","Console",...
func go2rtcPIDs() []int {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq go2rtc.exe", "/NH", "/FO", "CSV")
	hideProcessWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var pids []int
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(strings.ToLower(line), "go2rtc.exe") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 2 {
			if n, err := strconv.Atoi(strings.Trim(strings.TrimSpace(parts[1]), "\"")); err == nil {
				pids = append(pids, n)
			}
		}
	}
	return pids
}

// killAllGo2rtc terminates ALL go2rtc processes — including orphans left by a
// crashed or replaced run. taskkill /T takes
// the child processes (e.g. ffmpeg) along with it.
func killAllGo2rtc() (int, error) {
	n := countGo2rtc()
	if n == 0 {
		return 0, nil
	}
	cmd := exec.Command("taskkill", "/F", "/IM", "go2rtc.exe", "/T")
	hideProcessWindow(cmd)
	_ = cmd.Run() // return value irrelevant: processes not found are no error
	return n, nil
}

// killGo2rtcTree terminates a go2rtc process TOGETHER with its children (ffmpeg)
// immediately. Killing only the parent leaves the ffmpeg children running — they
// then hang on their RTSP connections for about ten seconds until they
// give up on their own. taskkill /T clears the whole tree at once, /F
// erzwingt es.
func killGo2rtcTree(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	k := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid))
	hideProcessWindow(k)
	_ = k.Run()
	_ = cmd.Process.Kill() // safety net in case taskkill did not take
}
