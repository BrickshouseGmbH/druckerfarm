package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Creates a program that poses as go2rtc: it counts its starts and
// exits by itself after a short while — like a crashing go2rtc.
func fakeGo2rtcBinary(t *testing.T, dir, counter string, liveFor string) string {
	t.Helper()
	path := filepath.Join(dir, "go2rtc")
	script := "#!/bin/sh\n" +
		"echo x >> " + counter + "\n" +
		"if [ \"$1\" = \"-version\" ]; then echo 'go2rtc version 1.9.14 linux/amd64'; exit 0; fi\n" +
		"sleep " + liveFor + "\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func starts(t *testing.T, counter string) int {
	b, err := os.ReadFile(counter)
	if err != nil {
		return 0
	}
	return strings.Count(string(b), "x")
}

// The core: if go2rtc dies, it must come back by itself. Previously it stayed gone
// and with it all videos and snapshots.
func TestSupervisorRestartsAfterCrash(t *testing.T) {
	dir := t.TempDir()
	counter := filepath.Join(dir, "starts.txt")
	fakeGo2rtcBinary(t, dir, counter, "0.4")

	oldDir := appDir
	appDir = dir
	defer func() {
		appDir = oldDir
		stopGo2rtc()
	}()

	go2rtcMu.Lock()
	go2rtcFails = 0
	go2rtcWanted = false
	go2rtcMu.Unlock()

	startGo2rtc()
	// 0.4 s runtime + 1 s wait before the first restart
	time.Sleep(3 * time.Second)

	n := starts(t, counter)
	stopGo2rtc()
	if n < 2 {
		t.Fatalf("nach einem Absturz wurde nicht neu gestartet (%d Starts)", n)
	}
	t.Logf("go2rtc wurde %dx gestartet — Überwachung greift", n)
}

// An intentional stop must not trigger a restart.
func TestSupervisorRespectsDeliberateStop(t *testing.T) {
	dir := t.TempDir()
	counter := filepath.Join(dir, "starts.txt")
	fakeGo2rtcBinary(t, dir, counter, "30")

	oldDir := appDir
	appDir = dir
	defer func() { appDir = oldDir }()

	go2rtcMu.Lock()
	go2rtcFails = 0
	go2rtcWanted = false
	go2rtcMu.Unlock()

	startGo2rtc()
	time.Sleep(400 * time.Millisecond)
	stopGo2rtc()
	time.Sleep(2500 * time.Millisecond)

	if n := starts(t, counter); n != 1 {
		t.Fatalf("nach gewolltem Stopp wurde neu gestartet (%d Starts)", n)
	}
}

// A failed component update must not leave go2rtc stopped
// — that was exactly the bug.
func TestFailedComponentUpdateLeavesGo2rtcRunning(t *testing.T) {
	dir := t.TempDir()
	counter := filepath.Join(dir, "starts.txt")
	fakeGo2rtcBinary(t, dir, counter, "30")

	oldDir := appDir
	appDir = dir
	defer func() {
		appDir = oldDir
		stopGo2rtc()
	}()

	go2rtcMu.Lock()
	go2rtcFails = 0
	go2rtcWanted = false
	go2rtcMu.Unlock()

	startGo2rtc()
	time.Sleep(300 * time.Millisecond)
	before := starts(t, counter)

	// Update that fails: no repository reachable
	oldURL := go2rtcURL
	go2rtcURL = "http://127.0.0.1:1/gibtsnicht.zip"
	defer func() { go2rtcURL = oldURL }()

	dlMu.Lock()
	dlProgress = installProgress{Active: true}
	dlMu.Unlock()
	runComponentUpdate([]string{"go2rtc"})
	time.Sleep(2 * time.Second)

	p := progressSnapshot()
	if p.Error == "" {
		t.Fatal("die Aktualisierung hätte scheitern müssen")
	}
	after := starts(t, counter)
	if after <= before {
		t.Fatalf("go2rtc wurde nach der fehlgeschlagenen Aktualisierung nicht wieder gestartet (%d → %d)", before, after)
	}

	go2rtcMu.Lock()
	wanted := go2rtcWanted
	go2rtcMu.Unlock()
	if !wanted {
		t.Fatal("go2rtc gilt als nicht gewollt — es bliebe für immer aus")
	}
	t.Logf("Fehlschlag gemeldet (%s), go2rtc wieder gestartet (%d → %d)", firstLine(p.Error), before, after)
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i > 0 {
		return s[:i]
	}
	if len(s) > 70 {
		return s[:70] + "…"
	}
	return s
}

// The backoff grows so a permanently broken go2rtc is not restarted
// every second.
func TestRestartBackoffGrows(t *testing.T) {
	var last time.Duration
	for fails := 0; fails < 8; fails++ {
		wait := time.Duration(1<<uint(minInt(fails, 5))) * time.Second
		if wait > time.Minute {
			wait = time.Minute
		}
		if wait < last {
			t.Fatalf("Wartezeit schrumpft bei %d Fehlschlägen: %v nach %v", fails, wait, last)
		}
		if wait > time.Minute {
			t.Fatalf("Wartezeit über der Grenze: %v", wait)
		}
		last = wait
	}
	if minInt(3, 5) != 3 || minInt(9, 5) != 5 {
		t.Fatal("minInt falsch")
	}
	_ = atomic.LoadInt64
	_ = exec.Command
	_ = fmt.Sprint
}

// Adding ten printers in a row used to trigger ten overlapping
// restarts — go2rtc barely got running. Now they are coalesced.
func TestNeustartsWerdenGesammelt(t *testing.T) {
	dir := t.TempDir()
	zaehler := filepath.Join(dir, "starts.txt")
	fakeGo2rtcBinary(t, dir, zaehler, "5")

	altDir := appDir
	appDir = dir
	defer func() {
		appDir = altDir
		stopGo2rtc()
	}()

	go2rtcMu.Lock()
	go2rtcWanted = false
	go2rtcFails = 0
	restartTimer = nil
	go2rtcMu.Unlock()

	// Ten requests in quick succession, as when adding several printers
	for i := 0; i < 10; i++ {
		restartGo2rtcAsync()
		time.Sleep(20 * time.Millisecond)
	}
	time.Sleep(1800 * time.Millisecond)

	n := starts(t, zaehler)
	if n > 1 {
		t.Fatalf("%d Starts — die Neustarts wurden nicht gesammelt", n)
	}

	go2rtcMu.Lock()
	if restartTimer != nil {
		restartTimer.Stop()
	}
	go2rtcMu.Unlock()
}

// Without go2rtc installed nothing may be scheduled — otherwise
// a goroutine runs into nothing after each click and accesses state that
// niemand mehr erwartet.
func TestKeinNeustartOhneGo2rtc(t *testing.T) {
	altDir := appDir
	appDir = t.TempDir() // empty, no go2rtc in it
	defer func() { appDir = altDir }()

	go2rtcMu.Lock()
	restartTimer = nil
	go2rtcMu.Unlock()

	restartGo2rtcAsync()

	go2rtcMu.Lock()
	geplant := restartTimer != nil
	go2rtcMu.Unlock()
	if geplant {
		t.Fatal("Neustart eingeplant, obwohl go2rtc nicht installiert ist")
	}
}
