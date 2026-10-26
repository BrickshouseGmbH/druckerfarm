# Druckerfarm — Checkliste

> **REGEL (von Reinhard, dreifach bestätigt): Die Versionsnummer bleibt unter 1.0.**
> Stand jetzt **0.9.1**. Innerhalb von 0.9.x wird weitergezählt (0.9.2, 0.9.3, …).
> **Niemals** von selbst auf 1.0 oder höher gehen — nur auf ausdrücklichen Befehl.


Laufende Notizen zum Projekt. Erledigtes bleibt stehen, damit nachvollziehbar ist, was warum so gebaut ist.

## Erledigt

### Drucker-Discovery per SSDP/UDP

Notiz aus einem früheren Chat, jetzt umgesetzt in `discovery.go`:

- Aktive Suche via M-SEARCH auf **Port 2021** (nicht 1900/1990) funktioniert, Drucker antworten direkt per Unicast mit `HTTP/1.1 200 OK`
- Für den Produktivbetrieb mehrfach senden wegen UDP-Paketverlust bei 36 Geräten im Netz → **3 Durchgänge mit 400 ms Pause**, getestet gegen einen Stand-in, der die ersten beiden Pakete verwirft
- Antwort-Felder: `DevModel`, `DevName`, `DevConnect` (lan/cloud), `DevBind` (free/occupied), `USN` (Seriennummer), `DevVersion` — alle werden gelesen. Der Parser trifft auch die Variante mit Suffix (`DevModel.suffix`), weil das je nach Firmware unterschiedlich kommt
- Suchtyp ist bewusst `ssdp:all` statt eines herstellerspezifischen Typs: unabhängig von der Firmware-Variante, und gefiltert wird über die Kopfzeilen, die nur Drucker mitschicken. Router und Fernseher fallen dadurch raus (eigener Test)
- Zusätzlich zur Multicast-Adresse wird an die Broadcast-Adresse jedes aktiven IPv4-Netzes gesendet — hilft, wenn Switches Multicast nicht weiterreichen
- Passives Lauschen auf periodische `NOTIFY`-Broadcasts (`ssdp:alive`) auf demselben Port: der Parser akzeptiert sie bereits, ein Dauerlauscher ist **noch nicht** gebaut (siehe Offen)
- Endpunkt: `GET /api/discover?timeout=5`, Knopf im Reiter *Drucker* über dem CSV-Bereich

Der Zugangscode steht nicht in der SSDP-Antwort und muss weiterhin von Hand eingetragen werden. Die Übernahme füllt Modell, Name, IP und Seriennummer vor und springt in das Code-Feld.

**Nicht gegen echte Drucker getestet** — mein Testrechner hat keine. Geprüft sind Parsen, Mehrfachversand, Deduplizierung und Filterung, jeweils über echtes UDP gegen einen Stand-in.

### Daten in %APPDATA%, Update über GitHub

Beides umgesetzt. Die Arbeitsdaten liegen in `%APPDATA%\Druckerfarm`, eine vorhandene Einrichtung wird beim ersten Start einmalig übernommen. Das Update holt sich das neueste GitHub-Release, prüft die Datei und ersetzt sich selbst über den Umbenennungs-Weg; die Vorgängerversion bleibt als `druckerfarm.exe.old` liegen, bis der nächste Start geglückt ist.

**Noch einzutragen:** das Repository unter *Einstellungen → Programmversion* (Format `besitzer/name`). Solange dort nichts steht, meldet die Prüfung genau das und tut sonst nichts.

Die Version kommt aus der Datei `VERSION` und wird beim Bauen eingestempelt. Für ein Release also: `VERSION` hochzählen, `build.ps1` laufen lassen, die Exe als Anhang an ein GitHub-Release hängen. Der Anhang muss auf `.exe` enden, sonst findet ihn die App nicht.

### Dateisuche und Löschen

Suche und Löschen laufen als verfolgbare Vorgänge (`/api/sync/search/start`, `/api/sync/delete/start`, Zwischenstand über `/api/job/status`, Abbruch über `/api/job/stop`). Treffer stehen aufgeklappt unter der jeweiligen Druckerzeile, Auswahl über Kästchen, Löschen über alle Drucker hinweg nach Rückfrage.

Erreichbarkeit angehoben: sechs gleichzeitige FTP-Sitzungen statt vier, 40 Sekunden je Drucker statt 20. **An der echten Farm nachprüfen** — beim letzten Lauf waren 13 von 42 Druckern nicht erreichbar.

### Builds ohne lokale Pfade

`build.ps1` und `build.sh` bauen mit `-trimpath`. Ohne das landen die Pfade der Quelldateien in der Exe — und damit der Name des Ordners, in dem gebaut wurde.

### Gebündeltes Löschen

Erledigt. `deletemany` nimmt mehrere Pfade über die Standardeingabe, eine FTP-Sitzung je Drucker.

**Die Speicher-Übersicht ist in 1.5.0 wieder entfallen** — auf deinen Wunsch, vollständig samt Endpunkten, `scan`-Aktion und CSS.

### Kammerbeleuchtung bei Fehler

`ledctrl` mit `led_mode: "flashing"`, je 1000 ms, **alle 5 s** aufgefrischt und geprüft. Modelle in `blink_models` (Vorgabe X1E, X2D), abschaltbar über `error_blink`.

**Achtung beim Ändern:** der X1 schickt seine vollständige HMS-Liste bei *jeder* Statusmeldung mit, auch während eines laufenden Drucks und auch dann, wenn die Meldung am Gerät nur nicht quittiert wurde. Deshalb merkt sich `PrinterStatus.AckedHms`, was beim Anlaufen des Drucks schon anstand; nur unbekannte Meldungen lösen Blinken aus. Wer das vereinfacht, bekommt entweder endloses Blinken oder verpasste Störungen — beides ist schon passiert.

### Pausierte Streams

Snapshots: Zeitfenster 8 s, danach wächst die Pause je Stream (15 s → 2 min). Während der Pause wird go2rtc nicht gefragt, die Oberfläche erfährt die Restzeit über `X-Snapshot-Retry-In`.

### Komponenten aktualisieren

go2rtc über das neueste GitHub-Release, ffmpeg über den Vergleich der Archiv-Prüfsumme (`component_sha` in der Konfiguration). Erstinstallation weiterhin mit fest hinterlegtem Hash, Aktualisierung nur mit Start- und Versionsprüfung.

### Logo

`assets/logo.svg` per `go:embed`, ausgeliefert unter `/logo.svg`, im dunklen Schema per CSS invertiert.

### Gestaltung ohne Inline-CSS

Rund 90 benannte Klassen, Sichtbarkeit über `setShown()` und die `hidden`-Klasse. Übrig sind 16 Attribute, die nur CSS-Variablen setzen (`--w`, `--mc`, `--tray`, `--cols`, `--grad`, `--o`), und 4 Zuweisungen für die Position der aufklappbaren Menüs.

### Blinken

Beim Wechsel auf RUNNING werden alte HMS-Meldungen und die Fehlernummer verworfen. Wer blinkt, steht in `blinking_ips` und übersteht damit einen Neustart. Schalter in den Einstellungen, Abschalten setzt das Licht sofort zurück.

### Zugriffe auf noch nicht gezeichnete Elemente

Teile der Oberfläche entstehen erst zur Laufzeit. Zugriffe darauf laufen über `byId`, `setText`, `setVal`, `setCls`, `setDis` — diese tun nichts, wenn das Element fehlt, statt die aufrufende Funktion abzubrechen. **Beim Erweitern bitte beibehalten**, sonst kommt die Sorte Fehler zurück.

### Snapshot-Diagnose

go2rtc antwortet bei fehlendem Bild mit 200 und leerem Rumpf und nennt den Grund nur im eigenen Log. `diagnoseStream()` prüft deshalb selbst: Stream in der Konfiguration (`/api/streams?src=` → 404), ffmpeg vorhanden, Kamera auf Port 322 erreichbar, MQTT-Zustand. **Nie wieder pauschal ffmpeg beschuldigen** — es ist auf der Farm installiert.

### go2rtc-Überwachung

`superviseGo2rtc` unterscheidet Absturz von gewolltem Stopp (`go2rtcWanted`, `go2rtcGen`) und startet nur im ersten Fall neu, mit wachsender Wartezeit bis 1 min. `go2rtcHealthLoop` prüft alle 30 s die Erreichbarkeit auf dem Port und fängt hängende Prozesse ab. **Wichtig:** jede Stelle, die go2rtc zum Ersetzen der Datei beendet, muss es danach in jedem Fall wieder starten — auch im Fehlerfall.

### go2rtc nachvollziehbar machen

`go2rtc.log` im Datenordner enthält die Ausgabe von go2rtc — **nie wieder auf `nil` setzen**, sonst ist jede Fehlersuche Raten. Nach dem Start vergleicht `verifyStreamsLoaded()` die Zahl der Drucker mit den geladenen Streams und meldet Abweichungen. Der Diagnose-Knopf unter *Einstellungen → Komponenten* liefert alles auf einmal.

**Doppelte Streamnamen sind tödlich:** go2rtc verwirft dann die gesamte streams-Sektion und kennt gar keinen Stream, läuft aber weiter. `namedStreams()` macht die Namen eindeutig — nicht entfernen.

### Snapshot-Rate

Jede Bildanfrage startet in go2rtc einen eigenen ffmpeg-Vorgang. `effectiveSnapInterval()` deckelt die Gesamtrate auf `snapMaxRatePerSec` (3/s) — bei 42 Druckern also mindestens 14 s je Drucker, bei wiederholten Abstürzen mehr. Die Oberfläche übernimmt den Wert aus `X-Snapshot-Interval`. **Nicht die Parallelität allein begrenzen** — der Stau wächst dann vor der Sperre.

### Die Seite wirklich laden, nicht nur prüfen

`node --check` prüft Syntax und sonst nichts. Eine Auslieferung ist mir schon einmal komplett durchgerutscht, weil ich das für einen Test gehalten habe. Es gibt jetzt `/tmp/domtest/load.js`: lädt die Seite mit `jsdom`, fängt Laufzeitfehler ab und liest die Beschriftungen so aus, wie sie im Fenster stehen. **Vor jeder Auslieferung laufen lassen.**

### Blinken: häufig prüfen, selten senden

Die Schleife prüft alle 5 Sekunden, sendet aber höchstens einmal pro Minute und Drucker. In 1.5.0 stand beides auf 5 Sekunden — das vervierfachte die MQTT-Befehle an die Farm ohne Nutzen, denn das Auffrischen dient nur dazu, einen Neustart des Druckers zu überstehen.

### Sichtbarkeit — Achtung beim Ändern

`setShown()` setzt **beide** Klassen: `hidden` zum Ausblenden, `shown` zum Einblenden. Elemente, die im Stylesheet grundsätzlich auf `display:none` stehen (`.view`, `.sd-files`, `.ir`), brauchen zwingend eine `.shown`-Regel — sonst bleiben sie für immer unsichtbar. Genau das hat in 1.3.0 sieben von acht Ansichten lahmgelegt.

Der Prüfstand `visible.js` liest die echten CSS-Regeln und rechnet für jede Ansicht den `display`-Wert nach dem Umschalten aus. **Nach jeder Änderung an Sichtbarkeit oder CSS laufen lassen.**

## Offen

- **Kamera-Port 322 bei 18 von 42 Druckern zu.** Mit dem Log vom 29.07. sauber eingegrenzt: alle 14 `woobly-h*` plus Nr. 10, 26, 27, 28. Ab der ersten Anfrage, anderthalb Stunden lang, 260 Versuche, kein Erfolg — also weder Last noch Wettlauf. `.51` weist aktiv ab (Gerät lebt, Dienst fehlt), die übrigen 17 antworten gar nicht.
  **Nächster Schritt am Gerät, nicht im Code:** *Einstellungen → 📡 Drucker erreichbar?* laufen lassen. Zeigt es „Kamera ist abgeschaltet", ist es der Schalter *LAN Mode Liveview*; zeigt es „Gerät aus oder nicht im Netz", ist es Strom, Netz oder VLAN.
  Zu klären bleibt, welches Modell die `h`-Serie ist: H2D/H2C/H2S können RTSPS auf 322 — P1P/P1S/A1 dagegen **gar nicht**, die liefern ihr Kamerabild über ein eigenes Protokoll auf Port 6000, mit dem go2rtc nichts anfangen kann. Im zweiten Fall wäre das kein Fehler, sondern ein Neubau.
- **Warum stürzt go2rtc unter Last ab?** Statt Anfragen abzulehnen, stirbt der Prozess. Der Grund steht seit 1.4.0 in `go2rtc.log` — sobald eine Datei von der echten Farm vorliegt, lässt sich das klären.
- **Nur öffentliche GitHub-Repositories.** Das Token-Feld ist auf Wunsch entfallen. Für ein privates Repository müsste es zurück.
- **Logo im dunklen Schema.** Wird derzeit invertiert, weil die weiße Fläche sonst als Kasten steht. Eine eigene helle Fassung des SVG wäre sauberer.
- **Welche Ordner darf der Löschknopf anfassen?** Weiter offen. Die Speicher-Übersicht, die das beantworten sollte, ist auf deinen Wunsch entfallen — die Ordnernamen müssen also anders herkommen, bevor der Löschknopf für Videos und Thumbnails gebaut werden kann.
- **Umbenennen auf dem Gerät — abschließend geprüft, geht nicht.** Die dokumentierte lokale Befehlsliste umfasst `info`, `pushing`, `upgrade`, `print`, `system`, `camera`, `xcam`; unter `system` stehen nur `ledctrl`, `get_access_code` und `set_accessories`. Kein Namensbefehl, auch nicht in den offenen Reverse-Engineering-Sammlungen. Der Name gehört zur Kontobindung, nicht zum Gerät — im LAN-Only-Betrieb gibt es deshalb keinen Weg dorthin; als Wunsch seit 2023 gemeldet und offen.
  **Der Umweg, der funktioniert:** LAN-Only kurz abschalten, in der Hersteller-App über „Geräte in der Nähe koppeln" umbenennen, LAN-Only wieder an. Etwa eine Minute je Gerät, der Name bleibt danach auch im LAN stehen. Diese App macht den Rückweg: sie liest den Gerätenamen per Netzwerksuche aus und bietet an, ihn zu übernehmen.
  Falls eine künftige Firmware das öffnet: alle Befehle laufen über dieselbe Stelle in `mqtt.go`.
- **Signatur für Updates.** Die App führt beim Update Code aus, den sie aus dem Netz geladen hat. Geprüft wird derzeit nur, ob es überhaupt eine Windows-Programmdatei ist. Gegen einen übernommenen GitHub-Zugang hilft das nicht — dafür bräuchte es eine Signatur mit einem Schlüssel, dessen öffentlicher Teil in der Exe steckt.
- **Passiver Discovery-Lauscher.** Dauerhaft auf Port 2021 auf `NOTIFY`/`ssdp:alive` lauschen und die Druckerliste im Hintergrund aktuell halten — dann fällt ein IP-Wechsel per DHCP von selbst auf. Der Parser kann es schon, es fehlt der Dauerlauscher plus die Frage, was bei einer geänderten IP automatisch passieren soll.
- **IP-Abgleich.** Wenn ein bekannter Drucker (gleiche Seriennummer) unter neuer IP auftaucht, aktuell nur als „bereits angelegt" markiert. Sinnvoll wäre ein Angebot, die IP zu aktualisieren.
- **Zugangscodes im Klartext.** Sie stehen weiterhin unverschlüsselt in `config.json` und `go2rtc.yaml` — jetzt immerhin in `%APPDATA%` statt im synchronisierten Ordner. go2rtc braucht sie im Klartext, eine Verschlüsselung müsste also im Moment des Schreibens der YAML wieder aufgelöst werden und brächte wenig.
- **Druck starten.** Aktuell nur Pause, Fortsetzen und Abbrechen. Eine Datei von der SD-Karte anzuwerfen wäre der nächste Schritt — die SD-Browse-Funktion ist vorhanden, die Startsequenz mit Platten- und AMS-Zuordnung ist der aufwendige Teil.
- **Staffelungsintervalle nachjustieren.** 80 ms zwischen Streamstarts, Snapshot-Abfragen über das Poll-Intervall verteilt, höchstens 4 gleichzeitige Transcodes. Begründete Schätzungen, an der echten Farm zu prüfen.

### Auswahl statt Filter im File Sync

`syncFilteredPrinters()` liefert seit 1.5.0 die **angehakten** Drucker, nicht mehr den Modellfilter. Suche und „Inhalte laden" hängen beide daran. Wer das zurückdreht, lässt die Suche wieder über alle 42 Geräte laufen.

### Kachel und Zeile: ein Bauweg

Seit 1.6.0 gibt es **eine** Datenquelle (`druckerFelder`), **einen** Feldgenerator (`feldHTML`) und **eine** Aktualisierung (`aktualisiere`). Die Anordnung macht ausschliesslich das Stylesheet: `.df-item.zeile` gegen `.df-item.kachel`.

Wer eine Spalte ergaenzen will: Eintrag in `DF_FELDER`, Fall in `feldHTML`, Breite im Raster von `.df-item.zeile,.list-thead`. **Nicht** einen zweiten Bauweg anfangen — genau das war die Ursache der Fehlerserie im Juli.

Kopfzeile und Datenzeile teilen sich eine einzige Rasterdefinition. Werden sie getrennt, laufen sie garantiert auseinander; das ist schon passiert (9 Spalten gegen 7).

Das Bild bleibt bewusst ausserhalb der gemeinsamen Felder — es gehoert nur in die Kachel und ins Vollbild.

### Steuerbefehle brauchen Developer Mode

Temperatur (`M104`/`M140` via gcode_line), Filament (`ams_filament_setting`), Pause/Weiter/Stop, Druckstart — alle über `SendRaw` mit Quittung. Firmware ab 01.08.05 weist sie ohne Developer Mode mit "mqtt message verify failed" ab; die Meldung verweist darauf.

**Filament-Hersteller** ist kein freies Feld: er steckt in `tray_info_idx` (GFA00 = PLA Basic usw.). Die Anzeigetexte dürfen keinen Herstellernamen enthalten — `strings druckerfarm.exe | grep -ci bambu` muss nach jedem Bau **0** sein.

### Netzwerkpause

`netzPausiert()` ist der Schalter. Wer eine neue Hintergrundschleife oder einen neuen Abruf baut, muss ihn abfragen — sonst ist die Pause loechrig.

Besonders zu beachten: **`superviseGo2rtc` darf waehrend der Pause nicht neu starten.** Es deutet sonst das gewollte Beenden als Absturz und hebelt die Pause aus. Ebenso `go2rtcHealthLoop`.

Der Zustand wird **nicht** gespeichert — nach einem Neustart laeuft das Programm wieder.

### go2rtc-Neustarts werden gesammelt

`restartGo2rtcAsync` setzt eine Frist von 800 ms neu, statt sofort zu starten. Ohne das loest jeder angelegte Drucker eine eigene Stopp-Start-Folge aus; bei zehn Geraeten hintereinander kommt go2rtc kaum zum Laufen. Die Pruefung auf vorhandenes go2rtc geschieht **sofort**, nicht in der Goroutine.

### Die Detailzeile nie unter laufender Bedienung neu bauen

`aktualisiere()` zieht die offene Detailzeile mit — aber nur, wenn `detailFrei(ip)` das erlaubt. Gesperrt wird bei offenem Formular (`data-halten="1"`), bei Eingabezeiger in der Zeile und bei aufgeklappter Auswahlliste.

Wer diese Sperre entfernt, macht jede Bedienung unmoeglich, die laenger als fuenf Sekunden dauert — genau das war der Zustand bis 1.11.0.

### Erst Daten, dann filtern

Beim Start gilt der Filter "Online". Wird gezeichnet, **bevor** der erste Status da ist, ist niemand online und das Grid bleibt leer; `applyOnlineFilter` rettet das nicht, weil es nur vorhandene Kacheln umschaltet. Deshalb: `await loadStatus()` vor dem ersten Zeichnen, und bei geaenderter sichtbarer Menge **neu zeichnen** statt umzuschalten (`ZULETZT_SICHTBAR`).

Langsames wie die Komponentenpruefung gehoert hinter das Zeichnen, nicht davor.

### Das laufende Programm pruefen, nicht nur den Quelltext

`CGO_ENABLED=0 go build -o /tmp/dflinux .` erzeugt aus demselben Quelltext einen Linux-Bau. Den starten und ueber `curl` bedienen — das prueft, was das Programm **tut**. Genau so ist aufgefallen, dass der Server Drucker ohne Zugangscode mit `400` abwies, waehrend die Oberflaeche sie anbot.

### Modellkennungen

Die Netzwerkantwort liefert Kennungen, keine Modellnamen: C11=P1P, C12=P1S, **C13=X1E**, N1=A1 mini, N2S=A1, O1D=H2D, BL-P001=X1C, BL-P002=X1 (`modelle.go`).

**H2C, H2S, X2D und P2S sind nicht belegt** — diese Kennungen bleiben absichtlich unuebersetzt stehen. Wer sie ergaenzt, sollte eine Quelle haben; ein falscher Name ist schlimmer als eine Kennung, die man nachschlagen kann.

### Erst das Naheliegende ausschliessen

Die H-Reihe lieferte wochenlang kein Bild. Ich habe nacheinander das Adressschema, den Liveview-Schalter und einen Kamera-Test gebaut. Die Ursache war: **die Geraete hatten neue IP-Adressen.**

Bei "Geraet antwortet nicht" deshalb in dieser Reihenfolge:
1. *Einstellungen → Drucker erreichbar?* — antwortet unter der eingetragenen Adresse ueberhaupt etwas?
2. Netzwerksuche — meldet sich dasselbe Geraet unter einer anderen Adresse? Umzuege werden gelb markiert.
3. Erst danach Schema, Firmware, Liveview.

### Netzwerksuche: hoeren, nicht nur fragen

M-SEARCH allein reicht nicht — die Geraete antworten darauf nicht verlaesslich. Sie melden sich aber von selbst per NOTIFY an `239.255.255.250:2021`. Dafuer muss auf **Port 2021** gelauscht und der Multicast-Gruppe beigetreten werden (`startDiscoveryListener`). Ein Empfaenger auf `0.0.0.0:2021` plus je Netzkarte ein Multicast-Empfaenger; scheitert einer, laufen die anderen weiter.

Der Suchbericht nennt die offenen Empfangswege und die Zahl der gehoerten Pakete. **Diese Auskunft nicht entfernen** — ohne sie ist "findet nichts" nicht von "hoert nichts" zu unterscheiden.

### Das Symbol der Programmdatei

`rsrc_windows_amd64.syso` traegt das Symbol; erzeugt wird es von `/tmp/mksyso.py` aus einer ICO-Datei. Der Go-Linker bindet jede `*.syso` automatisch ein.

**Achtung:** ein fehlerhaftes COFF wird stillschweigend verworfen — der Bau meldet hoechstens beilaeufig `fail to read string table` und laeuft durch. Nach dem Bauen deshalb pruefen, ob die `.rsrc`-Sektion wirklich in der Exe steckt. Symbol- und Zusatzeintrag sind **je genau 18 Byte**.

### Kamera-Adressen: gemessen schlaegt geraten

Die Reihenfolge in `kameraQuellen()` ist nur eine Vermutung aus dem Modellnamen. Der Knopf *Kamera-Adresse testen* misst am Geraet und legt das Ergebnis in `state.KameraSchema` ab; das hat Vorrang.

**`kameraQuellen` und `buildYaml` duerfen die Sperre NICHT selbst nehmen** — sie laufen unter `writeGo2rtcYaml`, das sie schon haelt. Die Karte wird hereingereicht. Ein Rueckfall haengt das Programm beim Start des Videodienstes auf; `TestBuildYamlSperrtNichtDoppelt` faengt das ab.

### Kamera-Adressen sind modellabhaengig

`rtspx://` ist eine Sonderform, die bei der X1-Reihe funktioniert. **H2/H2C/H2S und P2S antworten darauf mit einer Umleitung und liefern kein Bild** — sie brauchen `rtsps://`. In der Konfiguration stehen deshalb beide Adressen unter demselben Streamnamen, die modellrichtige zuerst; go2rtc probiert der Reihe nach durch.

Wer das auf ein Schema zusammenstreicht, sperrt eine ganze Geraeteklasse aus. Die Reihenfolge steht in `kameraQuellen()` in `main.go`.

### Betriebsstunden kommen nicht vom Geraet

Weder die Statusmeldung noch `info.get_version` enthalten einen Betriebsstundenzaehler. Die angezeigte Laufzeit zaehlt dieses Programm selbst mit (`laufzeitLoop`, Minutentakt, nur waehrend RUNNING) und ist in der Oberflaeche als *"gemessen seit Eintrag"* gekennzeichnet. **Diese Kennzeichnung nicht entfernen** — sonst sieht es aus wie ein Geraetezaehler.

Firmware und AMS-Staende dagegen sind echt und kommen aus `info.get_version`.

### Befehle gelten erst mit Quittung

Ein MQTT-Publish, das durchgeht, heisst nur: die Nachricht ist beim Broker. Ob der Drucker sie ausfuehrt, steht in seiner Antwort auf `device/<serial>/report` — gleicher Befehl, gleiche `sequence_id`, dazu `result` und `reason`. Darauf wartet `SendPrintCommand` und `SendRaw` bis zu vier Sekunden.

**Schweigt der Drucker**, ist er mit hoher Wahrscheinlichkeit an der Herstellercloud: dann nimmt er Statusabfragen an, aber keine Steuerbefehle, und antwortet auf sie gar nicht. Wer diese Wartezeit entfernt, bekommt wieder stille Fehlschlaege.

### Doppelte Funktionsnamen

Bisher **acht** Mal passiert: `loadAllSDFiles`, `setView` (zweimal), `updateListRow`, `editPrinter`, `savePrinter`, `fmtSize`, `setAllResolution`. In JavaScript gewinnt stillschweigend die letzte — kein Fehler, keine Warnung, nur falsches Verhalten. Die verlierende `editPrinter` hat einen Knopf still wirkungslos gemacht.

Der Prüfstand `/tmp/domtest/doppelt.js` findet das jetzt automatisch und muss **vor jeder Auslieferung** laufen. Das war die Ursache dafür, dass nur ein Drucker seine Dateien lud. Vor dem Anlegen einer Funktion nachsehen, ob es sie schon gibt:

```
grep -c "function NAME" ui.go
```

## Hinweise, die beim Bauen wichtig bleiben

- **Herstellername.** Kommt in Quellcode, Buildskripten und Oberfläche nicht mehr vor. Was bleiben **muss**, weil die Drucker es so verlangen: der RTSP-Benutzername `bblp`, das Schema `rtspx://` und die MQTT-Topics `device/<seriennummer>/request` bzw. `/report`. Das sind Protokollwerte, keine Bezeichnungen — ohne sie spricht die App nicht mehr mit den Geräten.
- **MQTT-Mitschrift.** Vollständige Payloads werden nur mit gesetzter Umgebungsvariable `DRUCKERFARM_MQTT_DEBUG` protokolliert. Ohne die Bremse wuchs das Logfile bei 42 Druckern auf über 700 MB.
- **go2rtc-Version.** Der SHA256 in `components.go` hängt an einem festen Release. Scheitert der Download an der Prüfsumme, ist eine neuere Fassung erschienen — `go2rtcVersion` und `go2rtcSHA256` anpassen.
