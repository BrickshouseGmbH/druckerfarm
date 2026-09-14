package main

import (
	"testing"
	"time"
)

func TestMischeSessions(t *testing.T) {
	jetzt := time.Now()
	alt := []accessSession{
		{Host: "PC-1", User: "reinhard", Zuletzt: jetzt.Add(-30 * time.Second)}, // eigener, frisch
		{Host: "PC-2", User: "anna", Zuletzt: jetzt.Add(-1 * time.Minute)},      // fremd, frisch
		{Host: "PC-3", User: "alt", Zuletzt: jetzt.Add(-10 * time.Minute)},      // fremd, veraltet
	}
	neu, andere := mischeSessions(alt, "PC-1", "reinhard", "DFxyz", jetzt)

	// PC-3 (veraltet) fliegt raus -> 2 Sitzungen (PC-1, PC-2)
	if len(neu) != 2 {
		t.Fatalf("erwarte 2 Sitzungen nach Bereinigung, bekam %d: %+v", len(neu), neu)
	}
	// eigener Eintrag aufgefrischt
	var eigen *accessSession
	for i := range neu {
		if neu[i].Host == "PC-1" {
			eigen = &neu[i]
		}
	}
	if eigen == nil || eigen.Instanz != "DFxyz" || !eigen.Zuletzt.Equal(jetzt) {
		t.Fatalf("eigener Eintrag nicht aufgefrischt: %+v", eigen)
	}
	// andere: nur PC-2
	if len(andere) != 1 || andere[0] != "anna@PC-2" {
		t.Fatalf("erwarte nur anna@PC-2, bekam %v", andere)
	}
}

func TestMischeSessionsLegtEigenenAn(t *testing.T) {
	jetzt := time.Now()
	neu, andere := mischeSessions(nil, "PC-9", "u", "DFabc", jetzt)
	if len(neu) != 1 || neu[0].Host != "PC-9" {
		t.Fatalf("eigener Eintrag sollte angelegt werden, bekam %+v", neu)
	}
	if len(andere) != 0 {
		t.Fatalf("keine anderen erwartet, bekam %v", andere)
	}
}
