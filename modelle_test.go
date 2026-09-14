package main

import "testing"

// Ein X1E meldet sich als "C13". Stand die Kennung ungeuebersetzt in der Liste,
// wusste niemand, welches Geraet gemeint war.
func TestModellName(t *testing.T) {
	belegt := map[string]string{
		"C11": "P1P", "C12": "P1S", "C13": "X1E",
		"N1": "A1 MINI", "N2S": "A1", "O1D": "H2D", "O1C": "H2C", "O1C2": "H2C",
		"BL-P001": "X1C", "BL-P002": "X1",
		"3DPrinter-X1-Carbon": "X1C", "3DPrinter-X1": "X1",
		"c13": "X1E", "  C13  ": "X1E",
	}
	for ein, will := range belegt {
		if got := modellName(ein); got != will {
			t.Fatalf("%q -> %q, erwartet %q", ein, got, will)
		}
	}
}

// Was schon wie ein Modellname aussieht, bleibt unangetastet — neuere Firmware
// meldet direkt den Namen.
func TestModellNameLaesstBekanntesStehen(t *testing.T) {
	for _, x := range []string{"H2D", "X1C", "P1S", "A1"} {
		if got := modellName(x); got != x {
			t.Fatalf("%q wurde veraendert zu %q", x, got)
		}
	}
}

// Unbekannte Kennungen werden NICHT geraten. Lieber eine Kennung zum
// Nachschlagen als ein falscher Name, dem man glaubt.
func TestUnbekannteKennungBleibtStehen(t *testing.T) {
	for _, x := range []string{"O1S", "X2D", "P2S", "ZZ99"} {
		if got := modellName(x); got != x {
			t.Fatalf("%q wurde zu %q geraten — das war nicht belegt", x, got)
		}
	}
	if modellName("") != "" {
		t.Fatal("leere Eingabe muss leer bleiben")
	}
}

func TestModellN6X2D(t *testing.T) {
	if got := modellName("N6"); got != "X2D" {
		t.Fatalf("N6 sollte X2D sein, war %q", got)
	}
	if got := modellName("n6"); got != "X2D" {
		t.Fatalf("n6 (klein) sollte X2D sein, war %q", got)
	}
}
