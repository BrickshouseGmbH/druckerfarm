package main

import "strings"

// ─── MODELLKENNUNGEN ──────────────────────────────────────────────────────────
//
// In der Netzwerkantwort steht nicht der Modellname, sondern eine interne
// Kennung: ein X1E meldet sich als "C13". Ohne Uebersetzung steht diese Kennung
// in der Liste, und niemand weiss, welches Geraet gemeint ist.
//
// Belegt sind die folgenden Zuordnungen (Profildateien im Slicer des
// Herstellers, resources/printers/):
//
//	C11   → P1P        C12   → P1S        C13 → X1E
//	N1    → A1 mini    N2S   → A1         O1D → H2D    O1C → H2C
//	BL-P001 → X1C      BL-P002 → X1
//
// Die H2-Reihe meldet sich nach dem Muster O1x: O1D = H2D, O1C = H2C (in der
// Praxis taucht die Kennung auch als "O1C2" auf). Fuer H2S, X2D und P2S habe ich
// keine belastbare Quelle — diese Kennungen bleiben unveraendert stehen, lieber
// eine nachschlagbare Kennung als ein falscher Name. Das Modell laesst sich
// ohnehin von Hand berichtigen.
var modellKennungen = map[string]string{
	"C11":                 "P1P",
	"C12":                 "P1S",
	"C13":                 "X1E",
	"N1":                  "A1 MINI",
	"N2S":                 "A1",
	"O1D":                 "H2D",
	"O1C":                 "H2C",
	"O1C2":                "H2C",
	"N6":                  "X2D",
	"BL-P001":             "X1C",
	"BL-P002":             "X1",
	"3DPRINTER-X1-CARBON": "X1C",
	"3DPRINTER-X1":        "X1",
}

// modellName uebersetzt eine Kennung, laesst aber alles unangetastet, was schon
// wie ein Modellname aussieht — neuere Firmware meldet direkt "H2D".
func modellName(roh string) string {
	k := strings.ToUpper(strings.TrimSpace(roh))
	if k == "" {
		return ""
	}
	if name, ok := modellKennungen[k]; ok {
		return name
	}
	return k
}
