package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// BoardEntry is one row of the board-planner table: for a combination of
// printer model and print file, the figures per board and the hours per board.
type BoardEntry struct {
	Model    string  `json:"model"`     // e.g. "H2D" (empty = any)
	File     string  `json:"file"`      // print file name (empty = any)
	PerPlate int     `json:"per_plate"` // Figuren pro Brett
	Hours    float64 `json:"hours"`     // Stunden pro Brett
}

// handleBoardTable: GET returns the table, POST replaces it entirely (the UI
// always sends the whole list). The table is settings data and therefore
// lives in config.json.
func handleBoardTable(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		list := make([]BoardEntry, len(state.BoardTable))
		copy(list, state.BoardTable)
		mu.Unlock()
		writeJSON(w, list)
	case http.MethodPost:
		var list []BoardEntry
		if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		clean := list[:0:0]
		for _, e := range list {
			e.Model = strings.TrimSpace(e.Model)
			e.File = strings.TrimSpace(e.File)
			if e.PerPlate < 0 {
				e.PerPlate = 0
			}
			if e.Hours < 0 {
				e.Hours = 0
			}
			// Skip empty rows (nothing filled in).
			if e.Model == "" && e.File == "" && e.PerPlate == 0 && e.Hours == 0 {
				continue
			}
			clean = append(clean, e)
		}
		mu.Lock()
		state.BoardTable = clean
		mu.Unlock()
		saveState()
		writeJSON(w, map[string]any{"ok": true, "count": len(clean)})
	default:
		http.Error(w, "GET oder POST erwartet", http.StatusMethodNotAllowed)
	}
}
