package main

import "testing"

func TestReconnectDurchlaufOhneSerial(t *testing.T) {
	// Drucker ohne Seriennummer werden übersprungen — kein Client wird angelegt.
	state.Printers = []Printer{{Name: "A", IP: "10.99.99.99", Model: "X2D"}} // kein Serial
	reconnectDurchlauf()
	if mqttMgr.IsConnected("10.99.99.99") {
		t.Fatal("ohne Serial darf keine Verbindung entstehen")
	}
}
