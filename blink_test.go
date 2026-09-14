package main

import "testing"

func TestPrinterWantsBlink(t *testing.T) {
	faelle := []struct {
		name string
		s    *PrinterStatus
		want bool
	}{
		{"fertig blinkt nicht", &PrinterStatus{Online: true, GcodeState: "FINISH"}, false},
		{"fertig trotz Restmeldung nicht", &PrinterStatus{Online: true, GcodeState: "FINISH", HmsErrors: []string{"X"}}, false},
		{"pause blinkt", &PrinterStatus{Online: true, GcodeState: "PAUSE"}, true},
		{"paused blinkt", &PrinterStatus{Online: true, GcodeState: "PAUSED"}, true},
		{"fehler FAILED blinkt", &PrinterStatus{Online: true, GcodeState: "FAILED"}, true},
		{"print_error blinkt", &PrinterStatus{Online: true, GcodeState: "RUNNING", PrintError: 5}, true},
		{"offene HMS blinkt", &PrinterStatus{Online: true, GcodeState: "RUNNING", HmsErrors: []string{"07FF0001"}}, true},
		{"laufend blinkt nicht", &PrinterStatus{Online: true, GcodeState: "RUNNING"}, false},
		{"idle blinkt nicht", &PrinterStatus{Online: true, GcodeState: "IDLE"}, false},
		{"offline blinkt nicht", &PrinterStatus{Online: false, GcodeState: "FAILED"}, false},
	}
	for _, c := range faelle {
		if got := printerWantsBlink(c.s); got != c.want {
			t.Errorf("%s: printerWantsBlink() = %v, erwartet %v", c.name, got, c.want)
		}
	}
}
