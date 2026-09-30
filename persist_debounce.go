package main

import (
	"sync"
	"time"
)

// ─── WRITE DEBOUNCING FOR runtime.json ────────────────────────────────────────
//
// Every saveState() also persists the small runtime caches (per-printer runtime,
// repair flags, pending firmware updates) to runtime.json. With ~50 printers,
// status handlers and blink/runtime bookkeeping can trigger saveState() many
// times per second, so writing runtime.json synchronously each time is wasteful
// disk churn. The debouncer coalesces a burst of requests into a single write:
// the first request arms a timer, later requests within the window are folded in,
// and one write happens when the window elapses. saveRuntime() itself always
// snapshots the *current* state under the lock, so a coalesced write never loses
// the latest values.
//
// On shutdown flushRuntimeSave() forces any pending write out immediately, so the
// last runtime values are never lost when the app exits.

const runtimeDebounce = 1500 * time.Millisecond

var (
	runtimeSaveMu    sync.Mutex
	runtimeSaveTimer *time.Timer
)

// scheduleRuntimeSave requests a runtime.json write soon, coalescing bursts.
func scheduleRuntimeSave() {
	runtimeSaveMu.Lock()
	defer runtimeSaveMu.Unlock()
	if runtimeSaveTimer != nil {
		return // a write is already pending; it will pick up the latest state
	}
	runtimeSaveTimer = time.AfterFunc(runtimeDebounce, func() {
		runtimeSaveMu.Lock()
		runtimeSaveTimer = nil
		runtimeSaveMu.Unlock()
		saveRuntime()
	})
}

// flushRuntimeSave cancels any pending debounced write and persists immediately.
// Called on shutdown so the final runtime values are always written.
func flushRuntimeSave() {
	runtimeSaveMu.Lock()
	if runtimeSaveTimer != nil {
		runtimeSaveTimer.Stop()
		runtimeSaveTimer = nil
	}
	runtimeSaveMu.Unlock()
	saveRuntime()
}
