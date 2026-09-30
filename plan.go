package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// ─── JOB-PLANER ───────────────────────────────────────────────────────────────
//
// The job planner is a matrix: rows are design figures, columns are printer
// MODELS (X1C, P1S, H2C, A1 …). Each cell is the planned quantity of that figure
// on that model. Models are not fixed — they come from the printers actually in
// the farm, so the columns follow the fleet. The plan is saved in config.json.

type PlanFigure struct {
	ID   string         `json:"id"`
	Name string         `json:"name"`
	File string         `json:"file,omitempty"` // optional 3MF/gcode reference
	Qty  map[string]int `json:"qty,omitempty"`  // printer model -> planned quantity
}

// planModels returns the distinct printer models currently in the farm, sorted.
func planModels() []string {
	seen := map[string]bool{}
	mu.RLock()
	for _, p := range state.Printers {
		m := strings.TrimSpace(p.Model)
		if m != "" {
			seen[m] = true
		}
	}
	mu.RUnlock()
	out := make([]string, 0, len(seen))
	for m := range seen {
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}

// handlePlan: GET returns the plan + current models; POST replaces the plan.
func handlePlan(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var body struct {
			Figures []PlanFigure `json:"figures"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Clean up: assign IDs to new figures, drop empty rows, trim names.
		clean := make([]PlanFigure, 0, len(body.Figures))
		for _, f := range body.Figures {
			f.Name = strings.TrimSpace(f.Name)
			if f.Name == "" {
				continue
			}
			if strings.TrimSpace(f.ID) == "" {
				f.ID = fmt.Sprintf("fig_%d", time.Now().UnixNano())
				time.Sleep(time.Nanosecond)
			}
			if f.Qty != nil {
				for k, v := range f.Qty {
					if v <= 0 {
						delete(f.Qty, k)
					}
				}
			}
			clean = append(clean, f)
		}
		mu.Lock()
		state.JobPlan = clean
		mu.Unlock()
		saveState()
	}
	mu.RLock()
	figs := make([]PlanFigure, len(state.JobPlan))
	copy(figs, state.JobPlan)
	mu.RUnlock()
	if figs == nil {
		figs = []PlanFigure{}
	}
	writeJSON(w, map[string]any{"figures": figs, "models": planModels()})
}
