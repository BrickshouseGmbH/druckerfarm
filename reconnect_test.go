package main

import "testing"

func TestReconnectDurchlaufOhneSerial(t *testing.T) {
	// Printers without a serial are skipped — no client is created.
	state.Printers = []Printer{{Name: "A", IP: "10.99.99.99", Model: "X2D"}} // no serial
	reconnectDurchlauf()
	if mqttMgr.IsConnected("10.99.99.99") {
		t.Fatal("ohne Serial darf keine Verbindung entstehen")
	}
}
