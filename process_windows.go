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

// zaehleGo2rtc zaehlt laufende go2rtc-Prozesse ueber die Windows-Prozessliste.
func zaehleGo2rtc() int {
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

// killAlleGo2rtc beendet ALLE go2rtc-Prozesse — auch verwaiste, die ein
// abgestuerzter oder ersetzter Programmlauf hinterlassen hat. taskkill /T nimmt
// die Kindprozesse (z. B. ffmpeg) gleich mit.
func killAlleGo2rtc() (int, error) {
	n := zaehleGo2rtc()
	if n == 0 {
		return 0, nil
	}
	cmd := exec.Command("taskkill", "/F", "/IM", "go2rtc.exe", "/T")
	hideProcessWindow(cmd)
	_ = cmd.Run() // Rueckgabe egal: nicht gefundene Prozesse sind kein Fehler
	return n, nil
}

// beendeGo2rtcBaum beendet einen go2rtc-Prozess SAMT seinen Kindern (ffmpeg)
// sofort. Nur den Elternprozess zu killen laesst die ffmpeg-Kinder laufen — die
// haengen dann noch rund zehn Sekunden an ihren RTSP-Verbindungen, bis sie von
// selbst aufgeben. taskkill /T raeumt den ganzen Baum in einem Rutsch, /F
// erzwingt es.
func beendeGo2rtcBaum(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	k := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid))
	hideProcessWindow(k)
	_ = k.Run()
	_ = cmd.Process.Kill() // Sicherheitsnetz, falls taskkill nicht griff
}
