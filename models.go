package main

import "strings"

// ─── MODELLKENNUNGEN ──────────────────────────────────────────────────────────
//
// The network response does not contain the model name but an internal
// identifier: an X1E reports itself as "C13". Without translation this code
// shows up in the list and nobody knows which device is meant.
//
// The following mappings are confirmed (profile files in the vendor's slicer):
// Herstellers, resources/printers/):
//
//	C11   → P1P        C12   → P1S        C13 → X1E
//	N1    → A1 mini    N2S   → A1         O1D → H2D    O1C → H2C
//	BL-P001 → X1C      BL-P002 → X1
//
// The H2 series reports as O1x: O1D = H2D, O1C = H2C (in practice the code
// also appears as "O1C2"). For H2S, X2D and P2S I have no reliable source —
// these codes are left unchanged; a lookupable code is better than a wrong
// name. The model can be corrected by hand anyway.
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

// modellName translates a code but leaves anything untouched that already
// looks like a model name — newer firmware reports "H2D" directly.
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
