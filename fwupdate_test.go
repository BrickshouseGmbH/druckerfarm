package main

import (
	"encoding/json"
	"testing"
)

// The printer reports a module with a newer number in new_ver_list.
func TestUpgradeStateFindetOffenesUpdate(t *testing.T) {
	raw := `{"new_ver_list":[
		{"name":"ota","cur_ver":"01.07.00.00","new_ver":"01.08.02.00"},
		{"name":"ams/0","cur_ver":"00.00.06.49","new_ver":"00.00.07.89"}
	]}`
	var v interface{}
	json.Unmarshal([]byte(raw), &v)
	got := parseUpgradeState(v)
	if len(got) != 2 {
		t.Fatalf("erwarte 2 Meldungen, bekam %d: %+v", len(got), got)
	}
	if got[0].Modul != "ota" || got[0].Neu != "01.08.02.00" {
		t.Errorf("ota falsch: %+v", got[0])
	}
}

// If the running build already equals the "new" one, nothing is pending.
func TestUpgradeStateIgnoriertGleichstand(t *testing.T) {
	raw := `{"new_ver_list":[{"name":"ota","cur_ver":"01.08.02.00","new_ver":"01.08.02.00"}]}`
	var v interface{}
	json.Unmarshal([]byte(raw), &v)
	if got := parseUpgradeState(v); len(got) != 0 {
		t.Errorf("erwarte 0, bekam %+v", got)
	}
}

// Leere Liste = nichts offen.
func TestUpgradeStateLeer(t *testing.T) {
	var v interface{}
	json.Unmarshal([]byte(`{"new_ver_list":[]}`), &v)
	if got := parseUpgradeState(v); len(got) != 0 {
		t.Errorf("erwarte 0, bekam %+v", got)
	}
}

// nochOffen removes an update once the running firmware has reached it.
func TestNochOffenStreichtInstalliertes(t *testing.T) {
	list := []FwModuleUpdate{
		{Modul: "ota", Aktuell: "01.07.00.00", Neu: "01.08.02.00"},
		{Modul: "ams/0", Aktuell: "00.00.06.49", Neu: "00.00.07.89"},
	}
	st := &PrinterStatus{
		Online: true,
		Info: &DeviceInfo{
			Firmware: "01.08.02.00", // ota jetzt installiert
			AMS:      []ModuleVersion{{Name: "ams/0", SW: "00.00.06.49"}},
		},
	}
	got := nochOffen(list, st)
	if len(got) != 1 || got[0].Modul != "ams/0" {
		t.Fatalf("erwarte nur ams/0 offen, bekam %+v", got)
	}
}

// Without status info (offline) the list stays untouched.
func TestNochOffenOhneStatusUnveraendert(t *testing.T) {
	list := []FwModuleUpdate{{Modul: "ota", Aktuell: "01.07.00.00", Neu: "01.08.02.00"}}
	if got := nochOffen(list, nil); len(got) != 1 {
		t.Errorf("offline soll erhalten bleiben, bekam %+v", got)
	}
}

// If the program does not know the running firmware, the discovery state applies — the
// update stays open (no wrong cleanup).
func TestNochOffenOhneInfoBehaeltFund(t *testing.T) {
	list := []FwModuleUpdate{{Modul: "ota", Aktuell: "01.07.00.00", Neu: "01.08.02.00"}}
	st := &PrinterStatus{Online: true} // no info
	if got := nochOffen(list, st); len(got) != 1 {
		t.Errorf("ohne Info soll offen bleiben, bekam %+v", got)
	}
}

// X1-Reihe (X1E): keine new_ver_list, sondern Einzelfelder je Baugruppe.
func TestUpgradeStateX1EinzelfelderAMS(t *testing.T) {
	raw := `{"new_version_state":2,"ota_new_version_number":"","ams_new_version_number":"00.00.07.89","ahb_new_version_number":""}`
	var v interface{}
	json.Unmarshal([]byte(raw), &v)
	got := parseUpgradeState(v)
	if len(got) != 1 || got[0].Modul != "ams" || got[0].Neu != "00.00.07.89" {
		t.Fatalf("erwarte ein AMS-Update, bekam %+v", got)
	}
}

// X1: OTA and AMS pending at the same time.
func TestUpgradeStateX1BeideModule(t *testing.T) {
	raw := `{"new_version_state":2,"ota_new_version_number":"01.08.02.00","ams_new_version_number":"00.00.07.89"}`
	var v interface{}
	json.Unmarshal([]byte(raw), &v)
	if got := parseUpgradeState(v); len(got) != 2 {
		t.Fatalf("erwarte 2 Module, bekam %+v", got)
	}
}

// Echter X1E (woobly 7): new_version_state==1, ABER ota_new_version_number
// filled — that is an update. The state is not a suitable switch.
func TestUpgradeStateEchterX1E(t *testing.T) {
	raw := `{"ahb_new_version_number":"","ams_new_version_number":"","consistency_request":false,"dis_state":0,"err_code":0,"ext_new_version_number":"","force_upgrade":false,"idx":5,"lower_limit":"00.00.00.00","message":"","module":"","new_version_state":1,"ota_new_version_number":"01.03.00.00","progress":"0","sequence_id":0,"sn":"03W09C442102388","status":"IDLE"}`
	var v interface{}
	json.Unmarshal([]byte(raw), &v)
	got := parseUpgradeState(v)
	if len(got) != 1 || got[0].Modul != "ota" || got[0].Neu != "01.03.00.00" {
		t.Fatalf("erwarte OTA-Update 01.03.00.00, bekam %+v", got)
	}
}

// Real H2 (woobly h12): new_version_state==0, all number fields empty — nothing
// offen.
func TestUpgradeStateEchterH2Nichts(t *testing.T) {
	raw := `{"ahb_new_version_number":"","ams_new_version_number":"","consistency_request":false,"dis_state":0,"err_code":0,"ext_new_version_number":"","force_upgrade":false,"idx":2722,"lower_limit":"00.00.00.00","message":"","module":"","new_version_state":0,"ota_new_version_number":"","progress":"0","sequence_id":0,"sn":"31B8BP610600131","status":"IDLE","upgrade_fail_list":[],"upgrade_type":"unknown"}`
	var v interface{}
	json.Unmarshal([]byte(raw), &v)
	if got := parseUpgradeState(v); len(got) != 0 {
		t.Errorf("H2 ohne Update soll nichts liefern, bekam %+v", got)
	}
}

// Placeholder 00.00.00.00 is not an update.
func TestUpgradeStateX1NullVersion(t *testing.T) {
	raw := `{"new_version_state":2,"ota_new_version_number":"00.00.00.00","ams_new_version_number":"00.00.00.00"}`
	var v interface{}
	json.Unmarshal([]byte(raw), &v)
	if got := parseUpgradeState(v); len(got) != 0 {
		t.Errorf("Nullversionen sind kein Update, bekam %+v", got)
	}
}

// The generic "ams" (without index) is checked against the lowest AMS level
// : if any AMS is below, the update stays open.
func TestNochOffenAmsGenerischNiedrigsterStand(t *testing.T) {
	list := []FwModuleUpdate{{Modul: "ams", Neu: "00.00.07.89"}}
	st := &PrinterStatus{Online: true, Info: &DeviceInfo{AMS: []ModuleVersion{
		{Name: "ams/0", SW: "00.00.07.89"}, // already current
		{Name: "ams/1", SW: "00.00.06.49"}, // still old -> update pending
	}}}
	if got := nochOffen(list, st); len(got) != 1 {
		t.Errorf("mindestens ein AMS ist alt -> Update offen, bekam %+v", got)
	}
	// If all are current, nothing is pending.
	st.Info.AMS[1].SW = "00.00.07.89"
	if got := nochOffen(list, st); len(got) != 0 {
		t.Errorf("alle AMS aktuell -> nichts offen, bekam %+v", got)
	}
}
