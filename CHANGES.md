# Druckerfarm — behobene Fehler

Arbeitsbericht zum Projekt. Der Herstellername kommt in Quellcode, Buildskripten, Oberfläche und Programmdatei nicht mehr vor — was bleiben muss, steht in `CHECKLISTE.md`.

Neue Dateien: `components.go`, `discovery.go`, `update.go`, `jobs.go`, `assets/logo.svg`, `VERSION`, `CHECKLISTE.md`, `testdata/`, 99 Tests.
Stand: **Version 0.9.75**.
Entfernt: `embed_binaries.go`, `binaries/`. Gebaut wird mit `build.ps1` → `druckerfarm.exe`, Version aus der Datei `VERSION`.
Arbeitsdaten liegen ab sofort in `%APPDATA%\Druckerfarm`, nicht mehr neben der Exe.

---

### Runde 4 (0.9.75)

- Neues Logo (`logo-druckerfarm2.svg`) und App-Icon (aus dem Hochkant-SVG erzeugt, mehrere Größen).
- Pillen (Online/Druckzustand/Modell): normale Schriftstärke und einheitliche Größe.
- Druckfehler-Sensorik: `xcam`-Felder (KI-/Kamera-Überwachung, Erstschicht-/Spaghetti-/Fremdkörper-Erkennung) werden gelesen und im Modal angezeigt („KI-Überwachung an/aus" + Abbruch-Empfindlichkeit); HMS-Codes wie bisher. Die rohe LIDAR-Punktwolke gibt die lokale Schnittstelle nicht her.
- Chinesisch (简体) und Spanisch als Sprachen (Schlüssel + Overlay), umschaltbar in der Sprachauswahl und den Einstellungen.
- File-Sync automatisch: Ordner je Druckermodell + Abweichungs-Check (Dateien auf SD, die nicht im Ordner liegen).
- Job-/Brett-Planer (Einstellungen): Zielmenge + Figuren pro Brett → Anzahl Bretter, Verteilung auf freie Drucker, Zeit-/Fertig-Prognose.
- KI-Assistent (Einstellungen): eigener API-Key für OpenAI (ChatGPT) oder Anthropic (Claude). Fehler-Erklärer und Kamera-Fehlererkennung (frisches Vollbild → Vision-Modell) direkt im „mehr"-Modal. Keys bleiben lokal (config.json, gitignored), werden nie geloggt oder nach außen gegeben.
- Aufräumen: deutsche Dateinamen auf Englisch (`networkpause.go`, `presence.go`, `models.go`, `control.go`, `CHANGES.md`, `CHECKLIST.md`, `LICENSE-REVIEW.md` …).
- Kleinere Fixes: Kachel-Flackern (Mengen- statt Reihenfolge-Vergleich, go2rtc-Debounce), Progress-Sort schiebt „bereit"-Drucker ans Ende, go2rtc-Neustart nur in den Einstellungen (mit PID-Anzeige).


## Version 0.9.75 — Streaming & Sichtbarkeit nach außen

Fünf neue Funktionen, alle mit `go test`, jsdom-Rendertest und Windows-Build geprüft; Herstellername in der Exe weiterhin 0×.

**1. „+Drucker"-Dialog erweitert.** Beim Anlegen eines Druckers (manuelle Eingabe) lassen sich jetzt drei Zustände direkt setzen: „Kamera aus", „Blinken aus", „In Reparatur". Backend: `POST /api/printers` nimmt die Flags entgegen und setzt `KameraAus`/`BlinkAus`/`Reparatur` gleich mit (`main.go`), Oberfläche in `ui.go` (`addM`).

**2. Einzelvollbild im Gesamt-Vollbild.** Im Vollbild aller Drucker hat jede Kachel unten rechts einen ⛶-Knopf, der genau dieses Video in den echten Browser-Vollbild schaltet (`cellVollbild` in `ui.go`); `stopPropagation` verhindert, dass der Klick das Overlay schließt.

**3. Sendeseite `/broadcast`.** Eine entkoppelte Nur-Ansicht (eigener Browser-Tab, eigener Zustand), unabhängig von der Bediennavigation. Festes Layout, keine Steuer-Buttons. Einstellbar in den Einstellungen: Ansicht (Alle/Online/Offline), Spalten, Klicken erlauben ja/nein. Neu: `broadcast.go`, Route `/broadcast`, `/api/broadcast/config`.

**4. Veröffentlichung per Cloudflare-Tunnel.** `cloudflared` wird wie go2rtc als Helfer-Binary geladen und gestartet; das Tool fängt die öffentliche `…trycloudflare.com`-Adresse ab. Wichtig für die Sicherheit: der Tunnel zeigt NICHT auf die Bedienoberfläche, sondern auf einen eigenen, streng lesenden Public-Server (Port 8766) — nur die Sendeseite und bereinigte Lesedaten, **keine** Zugangscodes der Drucker, **keine** Kamera-Zugangsdaten, keine Steuer-Endpunkte. Wählbar: öffentlich vs. geheim (Token), Klicken ja/nein. Neu: `cloudflare.go`, `/api/tunnel/*`.

**5. YouTube-Livestream direkt aus dem Tool.** Ffmpeg (bereits vorhanden) baut aus den Kamera-Streams ein Kachel-Mosaik und schiebt es per RTMP an YouTube. Die Eingänge sind die lokalen go2rtc-RTSP-Streams — dadurch stehen **keine** Drucker-Zugangscodes in der ffmpeg-Kommandozeile, und jeder Stream wird nur einmal vom Drucker geholt. Standardwerte (1920×1080, 4500 kbit/s, 15 fps) sind gesetzt und lassen sich überschreiben. Neu: `youtube.go`, `/api/youtube/*`.

> Hinweis: Für die Punkte 4 und 5 werden beim ersten Start die Helfer geladen (cloudflared) bzw. genutzt (ffmpeg). Der eigentliche Live-Push/Tunnel wurde in der Sandbox nicht gegen die echten Dienste getestet — die Argumentbildung und das URL-Abfangen sind per Unit-Test abgesichert (`streaming_test.go`).


## 1. Videospeicher läuft beim Start voll

Vier Ursachen, alle in `startWebRTC`/`stopWebRTC` in `ui.go`:

**Der Fehlerpfad hat die Verbindung nie geschlossen.** Im `catch`-Block stand `delete PCS[ip]`, aber kein `pc.close()`. Die Referenz war weg, die PeerConnection samt Video-Decoder lief weiter. Beim Start ist go2rtc noch nicht bereit, jeder SDP-Request scheitert — und wurde alle 3 s wiederholt. Das ist der Hauptgrund.

**Reconnect-Storm.** `startWebRTC` ruft als Erstes `stopWebRTC` auf. Das schließt die alte Verbindung, wodurch deren `onconnectionstatechange` mit `'closed'` feuert — und das hat einen weiteren `startWebRTC` in 2 s eingeplant. Jeder gewollte Neustart erzeugte also eine zusätzliche parallele Reconnect-Kette. Behoben durch `if (PCS[ip] !== pc) return;` am Anfang des Handlers plus `RETRY_TIMERS`, das maximal einen ausstehenden Reconnect pro Drucker zulässt.

**Keepalive-Intervall.** Das 15-s-`setInterval` wurde nur in `dc.onclose` aufgeräumt. Beim Schließen der PeerConnection feuert dieses Event nicht — das Intervall lief endlos weiter.

**ICE-Gathering.** Listener und Timeout wurden nicht entfernt, wenn der jeweils andere zuerst auflöste.

Alle Abbruchpfade laufen jetzt über eine gemeinsame `killPC()`-Funktion (Handler ab, Timer weg, Tracks gestoppt, `close()`).

### Nachgemessen

Node-Harness, das `startWebRTC`/`stopWebRTC` gegen eine simulierte Browser-Umgebung fährt: 8 Tile-Starts bei nicht erreichbarem go2rtc (= Startsituation), **pro Drucker**:

| | erzeugte PCs | offen geblieben | verwaiste Intervalle |
|---|---|---|---|
| vorher | 25 | **22** | **16** |
| nachher | 10 | 0 | 0 |

Bei 36 Druckern also rund 800 offene PeerConnections statt keiner.

Zusätzlich starten Streams jetzt gestaffelt (80 ms Abstand) statt 36 gleichzeitiger SDP-Handshakes — in `renderGrid`, `renderFSV`, `renderFSC`, der Fullscreen-Overlay und der go2rtc-Recovery.

### Weitere Änderungen (0.9.75, zweite Runde)

**Projektstruktur aufgeräumt.** Der gesamte Go-Code samt `vendor/`, `lang/`, `assets/` und `THIRD_PARTY_LICENSES.md` liegt jetzt im Unterordner `app/`. Das Hauptverzeichnis enthält nur noch Doku, Buildskripte und `VERSION`. Gebaut wird weiterhin über `build.ps1`/`build.sh` (rufen intern `app/` auf); die fertige `druckerfarm.exe` landet wie gehabt im Hauptverzeichnis.

**File-Sync: Meldung bei falschem Dateiformat.** Beim Hineinziehen/Auswählen wird jede Datei geprüft; alles außer `.gcode`/`.3mf` wird abgelehnt und klar gemeldet (Oberfläche und zusätzlich serverseitig beim Upload).

**Filter-Anzeige synchron.** Nach Wechsel FileSync → Übersicht → zurück stimmt die hervorgehobene Pille jetzt mit dem tatsächlich angewandten Filter überein (vorher zeigte sie „Alle", filterte aber noch nach dem zuletzt gewählten Modell).

**Fullscreen-Video-Only:** unter dem Druckernamen steht jetzt dezent die Restzeit.

**Kacheln:** Karten ohne Status/Fortschritt/Druckdatei behalten die gleiche Höhe wie die mit Infos (Platz wird reserviert). Fortschrittsbalken und Prozentzahl stehen in der Kachel auf einer Zeile. Der „mdot"-Punkt in der Übersicht wurde entfernt.

**Pillen:** Status-/Druckzustand-/Modell-Pille sitzen auf einer Linie; die „Online"-Pille hat dieselbe leichte Eckenrundung wie die anderen.

**„mehr"-Modal:** Blöcke stehen bei genug Platz nebeneinander und sind dann gleich hoch; bei wenig Platz stapeln sie.

### Feinschliff-Runde (0.9.75)

- Steuerungs-Buttons im Modal heißen jetzt „Druck pausieren / Weiter drucken / Druck abbrechen" (klare Abgrenzung zum Video).
- Die Geräte-Aktionen (Reparatur, Firmware, Kamera aus, Blinken aus) sitzen jetzt im Status-Block; in der Mitte bleiben nur Druck-/Temperatursteuerung.
- Sortierung „Fortschritt" ordnet jetzt nach Restzeit; der separate „Restzeit"-Sortierknopf ist entfallen.
- Pillen-Zeile (Online/Druckzustand/Modell) auf einheitlicher Höhe/Grundlinie.
- Dropdown-Pfeil bei „mehr" nur noch in der Liste, nicht im Grid.
- „Saved Printers"-Liste: Spalten neu (Modell, Status, Name …), Reparatur/Kamera/Blinken als Wort-Buttons, Löschen mit Rückfrage, „Alle löschen" mit Rückfrage.
- Filter- und Sortierleiste sitzen ohne Zwischenabstand direkt aufeinander; die Sortierleiste sieht aus wie die Filterleiste.
- go2rtc-Neustart ist jetzt direkt in der Stream-Leiste erreichbar (🔄 go2rtc).
- Kacheln flackern nicht mehr weg: ein einzelner go2rtc-Aussetzer (Timeout unter Snapshot-Last) wird erst nach dem zweiten Fehlversuch als Ausfall gewertet — dadurch wird das Grid nicht mehr bei jedem kurzen Haenger komplett neu gezeichnet.
- Schriftart auf Arial umgestellt; technische Werte (IP, Temperaturen, %) bleiben in Monospace.
- Hinweis zur Snapshot-Rate: die „2 s" werden bei vielen Druckern bewusst gedrosselt (Schutz vor go2rtc-Absturz) — der tatsaechliche Takt haengt an der Druckerzahl.

---

## 2. Snap-Modus

**Falscher Streamname.** Das Frontend baute ihn an vier Stellen als `'woobly-' + (Index+1)`, das Backend erzeugt ihn in `streamName()` aus dem Druckernamen. Solange die Drucker nicht exakt „Woobly 1…36" heißen, ging jede Anfrage ins Leere. Betroffen waren neben `fetchSnap` auch die go2rtc-Recovery und beide Vollbild-Ansichten — dort brachen also ebenfalls Streams weg.

**Fehler waren unsichtbar.** go2rtc antwortet auf `/api/frame.jpeg` mit **HTTP 200 und leerem Body**, wenn es kein Bild liefern kann. Der Code prüfte nur `r.ok` — jeder Fehlschlag sah aus wie ein Erfolg und endete in einer leeren Kachel. Verifiziert gegen die mitgelieferte go2rtc-Binary.

**36 parallele Snapshots überlasten go2rtc.** Jeder Snapshot lässt go2rtc einen ffmpeg-Transcode starten. Bei 12 parallelen Anfragen im Test war der go2rtc-Prozess anschließend weg. Bei 36 Druckern alle 2 s wäre das der Dauerzustand gewesen.

**Lösung:** neuer Endpoint `GET /api/snapshot/{ip}?max_age=N` in `main.go`. Er ermittelt den Streamnamen serverseitig über `streamName()` (keine zweite Kopie der Namensregeln mehr), cacht Frames pro Stream, lässt maximal 4 Transcodes gleichzeitig laufen, fasst gleichzeitige Anfragen für denselben Drucker zusammen und liefert bei Fehlern 503 mit lesbarem Grund statt eines leeren Bildes.

Im Frontend: gestaffeltes Polling über das Intervall verteilt, kein Nachladen solange eine Anfrage läuft, saubere Freigabe der Blob-URLs, Zeitstempel bzw. Fehlertext direkt in der Kachel.

---

## Wichtig für den Betrieb: ffmpeg

Die Druckerkameras liefern H264. Für JPEG-Snapshots muss go2rtc transcodieren und startet dafür **ffmpeg** — eine Binary, die die App nicht mitbringt. Ohne `ffmpeg.exe` im PATH (oder neben `go2rtc.exe`) bleibt der Snap-Modus leer; ab jetzt immerhin mit der Meldung „ffmpeg fehlt im PATH" statt schweigend.

Das war die offene Entscheidung — sie ist inzwischen beantwortet, siehe Punkt 3.

---

## 3. Komponenten werden nachgeladen statt eingebettet

`embed_binaries.go` und der Ordner `binaries/` sind weg. go2rtc und ffmpeg werden bei Bedarf heruntergeladen, was zwei Dinge bewirkt: die Exe schrumpft von **33,5 MB auf 7,1 MB**, und ffmpeg ist nicht mehr die stille Voraussetzung, an der der Snap-Modus scheitert.

**Ablauf.** Beim Start prüft die App, was im Programmordner liegt. Fehlt etwas und wurde der Dialog nicht dauerhaft weggeklickt, erscheint ein Popup mit „Jetzt herunterladen" / „Später" / „Nicht mehr fragen" und einem Fortschrittsbalken. Bis dahin ist die App voll benutzbar — Drucker anlegen, MQTT-Status, Datei-Sync laufen ohne beide Komponenten. Nur Video und Snapshots fehlen. „Nicht mehr fragen" landet in `config.json`; nachholen geht jederzeit über den neuen Abschnitt **Komponenten** im Setup-Reiter.

**Quellen und Integrität.**

| | Quelle | Größe | Prüfung |
|---|---|---|---|
| go2rtc v1.9.14 | github.com/AlexxIT/go2rtc | ~7 MB | fest hinterlegter SHA256 |
| ffmpeg | github.com/BtbN/FFmpeg-Builds | ~85 MB | SHA256 aus `checksums.sha256` desselben Release |

go2rtc hängt an einem festen Release-Tag, deshalb ist der Hash im Code eingetragen. BtbN baut unter dem rollenden Tag `latest` ständig neu — ein fester Hash würde dort nach jedem Rebuild brechen, also wird die mitveröffentlichte Prüfsummenliste geladen und daraus verifiziert. Beide Downloads werden zusätzlich durch einen Aufruf von `-version` als lauffähig bestätigt, bevor sie an ihren Platz wandern.

Aus dem ffmpeg-Archiv wird nur `bin/ffmpeg.exe` samt DLLs entpackt; ffplay, ffprobe, Header und Import-Libs bleiben draußen. Beim Entpacken wird nie der Name aus dem Archiv als Zielpfad übernommen, sondern immer ein selbst bestimmter Dateiname — damit ist Zip-Slip ausgeschlossen (eigener Test).

**PATH.** ffmpeg landet in `<Programmordner>\ffmpeg\`, und go2rtc bekommt dieses Verzeichnis beim Start vorne in seinen PATH gesetzt. Du musst am System-PATH nichts ändern.

Ein bereits systemweit installiertes ffmpeg wird erkannt und dann gar nicht erst zum Download angeboten.

---

## 4. ERR_CONNECTION_REFUSED beim ersten Start

Das war kein Zufall und auch keine Folge der Umbauten — der Fehler steckte von Anfang an im Startcode.

`waitForWindow` wartet mit `cmd.Wait()` auf den gestarteten `msedge.exe` und fährt danach die App herunter. Der Prozess, den man da startet, ist unter Windows aber oft nur ein Starter: bei kaltem Profil initialisiert Edge sich und startet sich selbst neu, und wenn schon ein Edge dasselbe Profil hält, reicht er die Kommandozeile weiter. In beiden Fällen endet der gestartete Prozess nach Sekundenbruchteilen — `cmd.Wait()` kehrt zurück, `main()` läuft aus, **der HTTP-Server ist weg**. Das Fenster, das gerade erst hochkommt, findet auf Port 8765 dann nichts mehr.

Verschärft hat das `appUserDataDir()`: es löschte das Edge-Profil bei **jedem** Start. Damit war jeder Start ein Kaltstart — genau die Situation, die das frühe Prozessende auslöst. Beim zweiten Klick ist Edge selbst warm, der Starter bleibt lange genug am Leben, und es funktioniert. Daher „immer beim ersten Mal".

**Behoben an drei Stellen:**

Das Ende des Browser-Prozesses zählt jetzt nur noch als Hinweis. Das verlässliche Signal ist das UI selbst: das Dashboard pollt alle 5 s, also wird erst heruntergefahren, wenn 15 s lang kein Request mehr kam. Kommt gar keiner, gibt die App nach 90 s auf.

Der Listener wird synchron geöffnet. Vorher lief `http.ListenAndServe` in einer Goroutine und der Rückgabewert wurde verworfen — bei belegtem Port startete eine zweite Instanz halb und riss beim Beenden die erste mit. Jetzt heißt belegter Port: es läuft schon eine Instanz, Fenster darauf öffnen, fertig.

Das Edge-Profil wird nicht mehr bei jedem Start gelöscht. Zum Cache-Busting war das ohnehin überflüssig, `serveUI` liefert die Seite mit `no-store` aus.

**Dazu ein Logfile.** Durch `-H windowsgui` gibt es keine Konsole, jede `log.Printf`-Meldung ging ins Leere — deshalb war so ein Fehler bisher gar nicht diagnostizierbar. Ab jetzt schreibt die App nach `druckerfarm.log` im Programmordner (bei über 1 MB wird rotiert).

**Nachgestellt.** Mit einem Fake-Browser, der sich sofort beendet — also genau dem Windows-Verhalten:

| | erreichbar über 10 s | Prozess |
|---|---|---|
| vorher | **0 von 5 Abfragen** | beendet |
| nachher | 5 von 5 Abfragen | läuft |

Und das Herunterfahren funktioniert weiterhin: sobald das UI still ist, endet die App nach 15 s.

---

## 5. Gemeldete Fehler und neue Funktionen

### „Alle stoppen" hatte keine Wirkung

`stopAll()` setzte nur `ST[ip]='idle'` und rendert neu — `stopWebRTC` wurde nie aufgerufen. Das Neurendern warf bloß die `<video>`-Elemente weg, die Verbindungen liefen unsichtbar weiter. Gleiches bei „Nicht-Favoriten stoppen".

Jetzt läuft **jeder** Zustandswechsel über eine einzige Funktion `setTileState(ip, zustand)`, die beim Verlassen eines Zustands dessen Ressourcen abräumt. Damit kann „stoppen" das Schließen nicht mehr vergessen — auch nicht beim Wechsel auf 2s/5s/10s.

### Fullscreen öffnete zweite Streams

Die Vollbildansichten bauten eigene `RTCPeerConnection`s auf, getrennt von der zentralen Verwaltung, und ignorierten den gewählten Modus. Neu: laufende Streams landen in `STREAMS[ip]` und werden beim Ansichtswechsel nur **neu angehängt**. Vor dem Aufbau der Vollbildansicht wird das Grid geleert, sonst gäbe es die Element-IDs doppelt. Fullscreen übernimmt jetzt den aktiven Modus (2s/5s/10s bzw. Video).

### Neue Drucker blieben ohne Daten

`connectAllMQTT()` lief genau einmal beim Start — beim Anlegen eines Druckers wurde nur die go2rtc-Config geschrieben, keine MQTT-Verbindung aufgebaut. Der Drucker blieb bis zum Neustart datenlos. Jetzt wird beim Anlegen, Ändern und Löschen abgeglichen, und alle 45 s fordert die App den vollen Status neu an, damit eine verlorene Erstantwort nicht dauerhaft schadet.

Dazu ein zweiter Fehler im Parser: alle Felder standen unter `if v > 0`. Ein Drucker bei 0 % bekam nie einen Fortschritt gesetzt und zeigte stattdessen den Wert des vorigen Auftrags weiter. Der Drucker schickt Teil-Updates, deshalb wird jetzt nach *Anwesenheit* des Schlüssels entschieden: fehlt er, bleibt der alte Wert; steht er auf 0, ist es wirklich 0.

Der Fortschrittsbalken wird außerdem immer gerendert, auch ohne Daten — mit „warte auf Daten …" statt einer leeren Kachel.

### „mehr" war unsichtbar

Kein z-index-Problem, sondern Clipping: `.tile-name-wrap` und `.tile-header-row1` haben `overflow:hidden`, das Menü stand auf `top:100%` und wurde schlicht abgeschnitten. Es hängt jetzt als Popover am `body` und wird an der Kachel ausgerichtet — damit kann es nichts mehr wegschneiden.

### Restzeit im Grid

Unter dem Fortschrittsbalken steht jetzt Prozent, Restzeit und Schicht — dieselben Angaben wie in der Listenansicht.

### 2s als Vorgabe

Startet in der 2s-Snapshot-Ansicht statt 720p. Fehlt ffmpeg, bleibt der Modus aktiv und die Kacheln zeigen den Grund an (so gewählt).

### AMS, Farben, Drucksteuerung

`PrinterStatus` kennt jetzt AMS-Einheiten mit Fächern, Filamenttyp, Farbe und Restmenge sowie die externe Spule. Im „mehr"-Menü erscheinen sie als Farbpunkte mit Beschriftung, das aktive Fach ist hervorgehoben; mehrfarbiges Filament wird als Verlauf dargestellt.

Pause, Fortsetzen und Abbrechen gehen als MQTT-Kommando raus. Abbrechen fragt mit Druckernamen nach, Pause und Fortsetzen laufen direkt. Der Endpunkt verweigert Aufrufe ohne ausdrückliche Druckerliste, damit ein Fehlaufruf nicht die ganze Farm trifft.

**Gegen echte Daten geprüft.** Die AMS-Felder habe ich nicht geraten: in deinem `druckerfarm.log` standen vollständige MQTT-Payloads. Daraus ist ein Testdatensatz geworden (`testdata/ams_real.json`), gegen den ein Test läuft — er liest 3 belegte Fächer mit den Farben `161616FF`, `D4B1DDFF`, `FFFFFFFF` und erkennt das leere Fach als leer.

### Logflut

`mqtt.go` schrieb **jede** empfangene Nachricht in voller Länge mit. Dein Logfile war dadurch **647 MB** groß — und lag in OneDrive. Das Payload-Protokoll hängt jetzt an der Umgebungsvariable `DRUCKERFARM_MQTT_DEBUG`; im Normalbetrieb bleiben ein paar Zeilen übrig.

### Ohne Edge oder Chrome

Der Rückfallweg öffnete zwar den Standardbrowser, aber `waitForWindow` kehrte bei fehlendem Prozess zurück, ohne das Abschaltsignal zu geben — die App lief endlos weiter, auch nach dem Schließen des Tabs. Jetzt übernimmt die Leerlauf-Erkennung, und im Setup-Reiter steht ein Hinweis samt Adresse.

---

## 6. Umbenennung auf „Druckerfarm" und Netzwerksuche

### Umbenennung

Der Herstellername ist aus Quellcode, Buildskripten, Modulnamen und Oberfläche verschwunden — geprüft mit einer Volltextsuche über alle Dateien außer `vendor/`, Ergebnis null Treffer. Betroffen waren unter anderem das Go-Modul, die FTP-Hilfstypen in `sync.go`, die Umgebungsvariable für das MQTT-Protokoll, die Namen von Exe, Logdatei, CSV-Export und Edge-Profil sowie zwei Schlüssel im lokalen Browserspeicher.

**Was bleiben muss**, weil die Drucker es so verlangen: der RTSP-Benutzername `bblp`, das Schema `rtspx://` und die MQTT-Topics `device/<seriennummer>/request` bzw. `/report`. Das sind Protokollwerte, keine Bezeichnungen — ohne sie spricht die App nicht mehr mit den Geräten. Wenn dir auch das zu nah ist, müsste man das Protokoll wechseln, nicht die Zeichenkette.

Zwei Nebenwirkungen: die im Browser gespeicherte Spaltenzahl und Sprachwahl setzen sich einmalig zurück (ich hatte eine Übernahme eingebaut, sie aber wieder entfernt — sie hätte den alten Namen im Code behalten). Und den Diff gegen die Originalquellen habe ich aus der Auslieferung genommen, weil sein Vorher-Teil den alten Code zwangsläufig enthält.

### Drucker im Netzwerk suchen

Neu in `discovery.go`, Knopf im Reiter *Drucker* über dem CSV-Bereich, Endpunkt `GET /api/discover?timeout=5`.

M-SEARCH geht an **Port 2021**, nicht an den üblichen 1900. Gesendet wird **drei Mal mit 400 ms Pause** — bei 36 Geräten geht sonst verlässlich ein Paket verloren. Zusätzlich zur Multicast-Adresse wird an die Broadcast-Adresse jedes aktiven IPv4-Netzes gesendet, falls Switches Multicast nicht weiterreichen.

Als Suchtyp steht `ssdp:all` statt eines herstellerspezifischen Typs: unabhängig von der Firmware-Variante, und gefiltert wird über die Kopfzeilen, die nur Drucker mitschicken. Gelesen werden `DevModel`, `DevName`, `DevConnect`, `DevBind`, `USN` und `DevVersion`, jeweils auch in der Schreibweise mit Suffix. Periodische `NOTIFY`-Broadcasts versteht der Parser ebenfalls.

Der Zugangscode steht nicht in der Antwort. „Übernehmen" füllt Modell, Name, IP und Seriennummer vor und springt ins Code-Feld. Bereits angelegte Drucker werden erkannt und nicht doppelt angeboten.

**Nicht gegen echte Drucker getestet** — die habe ich nicht. Geprüft sind Parsen, Filterung, Mehrfachversand und Deduplizierung, davon zwei Tests über echtes UDP gegen einen Stand-in: einer antwortet erst auf das dritte Paket (Paketverlust), einer doppelt (Deduplizierung).

---

## 7. Daten in %APPDATA%, Update über GitHub, Oberfläche

### Alle Daten in %APPDATA%

`config.json`, `go2rtc.yaml`, `druckerfarm.log`, `go2rtc.exe` und der `ffmpeg\`-Ordner liegen jetzt unter `%APPDATA%\Druckerfarm`. Damit sind die Zugangscodes aus dem synchronisierten Projektordner heraus, und — wichtiger für den nächsten Punkt — die Exe ist austauschbar geworden.

Eine vorhandene Einrichtung wird beim ersten Start **einmalig übernommen**, damit niemand 42 Drucker neu einträgt. Ein zweiter Lauf rührt eine inzwischen geänderte Konfiguration nicht mehr an (eigener Test).

### Update über GitHub

Die Version wird beim Bauen aus der Datei `VERSION` eingestempelt. Repository und optionaler Token stehen in den Einstellungen — du kannst sie später eintragen, ohne dass etwas neu gebaut werden muss.

Ablauf: Release abfragen, Version vergleichen, Datei laden, prüfen, ersetzen, neu starten. Das Ersetzen nutzt den beschriebenen Weg — Windows verbietet das Überschreiben einer laufenden Datei, erlaubt aber das Umbenennen. Die App legt sich als `druckerfarm.exe.old` beiseite, schreibt die neue Fassung an den frei gewordenen Pfad, startet sie und beendet sich. Beim nächsten Start verschwindet die alte Datei — bis dahin ist sie der Rückweg.

Der Versionsvergleich rechnet zahlenweise, sonst wäre 1.12 kleiner als 1.9. Vor dem Ersetzen wird geprüft, ob die geladene Datei überhaupt ein Windows-Programm ist — eine HTML-Fehlerseite von passender Größe darf nicht an den Platz der laufenden App wandern.

**Getestet mit echten Dateien:** Ersetzen samt Start der neuen Fassung, Rückweg-Kopie, und der Fehlerfall — schlägt das Ablegen fehl, kommt die alte Fassung zurück, statt dass der Rechner ohne lauffähiges Programm dasteht.

Die Sicherheitsabwägung von neulich gilt unverändert: die App führt damit Code aus, den sie aus dem Netz geladen hat. HTTPS und die Formatprüfung fangen Übertragungsfehler ab, nicht aber einen übernommenen GitHub-Zugang.

### Zwei Farbschemata

Reines Schwarz und reines Weiß, umschaltbar in den Einstellungen. Die kontrastarmen Grautöne sind raus — `--muted` stand auf `#4a5968`, das ergab auf dem alten Grund **2,8:1** und war kaum zu entziffern. Nachgerechnet nach WCAG:

| | Fließtext auf Grund | schwächster Wert |
|---|---|---|
| vorher | 2,8:1 | unter jeder Norm |
| dunkel | 21,0:1 | 10,8:1 |
| hell | 21,0:1 | 8,9:1 |

WCAG AA verlangt 4,5:1. Der Scanline-Effekt läuft nur noch im dunklen Schema, im hellen machte er die Fläche schmutzig.

### Einstellungen und Header

Sprache und Anzahl der Videos nebeneinander sind aus dem Header in die Einstellungen gewandert. Beides liegt jetzt **serverseitig** in `config.json` statt im Browserspeicher — der war weg, sobald das Edge-Profil neu angelegt wurde. Fenstertitel und Kopfzeile heißen „Druckerfarm".

Die Knöpfe „Favoriten starten" und „Nicht-Favoriten stoppen" sind entfernt.

### Streams starten von selbst

Beim Öffnen prüft die App, ob go2rtc **und** ffmpeg vorhanden sind, und schaltet dann automatisch in die 2s-Ansicht. Fehlt eines von beidem, bleibt es beim Hinweis — sonst stünden 42 Kacheln mit derselben Fehlermeldung da.

### Drucker bearbeiten

In der Druckerliste lassen sich Name, Zugangscode und Seriennummer direkt ändern, ohne den Drucker zu löschen und neu anzulegen. Enter speichert, Escape bricht ab.

### Dateisuche über mehrere Drucker

Im Datei-Reiter über dem Modellfilter. Sucht ab drei Zeichen mit kurzer Verzögerung, wahlweise über die gefilterten oder alle Drucker, höchstens vier FTP-Sitzungen gleichzeitig. Treffer werden nach Drucker gruppiert — so sieht man, wo dieselbe Datei mehrfach liegt. Nicht erreichbare Drucker werden benannt statt verschwiegen.

---

## 8. Dateisuche und Löschen

### Suche ist jetzt ein Vorgang, kein einzelner Aufruf

Vorher wurde einmal gefragt und einmal geantwortet — bei 42 Druckern hieß das: warten und dann alles auf einmal. Jetzt wird ein Vorgang gestartet, dessen Zwischenstand die Oberfläche abfragt. Treffer erscheinen, sobald der jeweilige Drucker geantwortet hat, der Fortschrittsbalken zeigt „18 von 42", und der Knopf wird zu **„Suche stoppen"**.

Der Zwischenstand liefert nur, was seit dem letzten Abruf dazugekommen ist. Sonst müsste die Oberfläche die Liste zweimal pro Sekunde neu bauen.

### Warum 13 von 42 nicht erreichbar waren

Vermutlich zu knapp bemessen: vier gleichzeitige FTP-Sitzungen bei 20 Sekunden Zeitfenster. Da die Treffer jetzt einzeln eintröpfeln, kostet ein größeres Fenster nichts an gefühlter Geschwindigkeit — **sechs gleichzeitig, 40 Sekunden je Drucker**. Nicht erreichbare Geräte stehen zudem direkt in ihrer eigenen Zeile mit dem Grund als Kurzinfo, statt gesammelt am Ende.

Ob das reicht, zeigt der nächste Lauf an der echten Farm. Falls weiter Geräte ausfallen, wäre der nächste Schritt, die Fehlermeldung des FTP-Aufrufs mitzuprotokollieren.

### Treffer stehen beim Drucker

Die Ergebnisse liegen nicht mehr im Block über der Suchleiste, sondern aufgeklappt unter der jeweiligen Druckerzeile. Jede Zeile zeigt die Trefferzahl, „keine Treffer" oder „nicht erreichbar". Die Auswahl „alle Drucker statt nur gefilterte" ist weg — gesucht wird im aktuellen Filter, und „alle" ist selbst ein Filter.

### Löschen

Auswahlkästchen an jedem Treffer, dazu „alle auswählen" und „Auswahl aufheben". Ein Knopf löscht die gesamte Auswahl über alle Drucker hinweg, nach Rückfrage mit Anzahl und Druckerzahl.

Das Löschen läuft über dieselbe Vorgangsmechanik: Fortschritt „37 von 126 gelöscht", danach ein Bericht. Erfolgreich gelöschte Dateien verschwinden aus der Liste, fehlgeschlagene bleiben stehen — mit den Namen der betroffenen Drucker.

Dateinamen mit Pfadtrennern werden abgewiesen. Ohne diese Prüfung wäre ein Name wie `../../etc/…` ein Weg aus dem SD-Verzeichnis heraus (eigener Test).

### Footer

„Made by Reinhard from druckerfarm.ch" ist entfernt.

### Nebenbefund: der Herstellername steckte in der Exe

Beim Prüfen des neuen Ordners fiel auf, dass die Programmdatei den Namen doch noch enthielt — nicht im Quellcode, sondern als **eingebetteter Dateipfad**. Go schreibt die Pfade der Quelldateien für Absturzmeldungen mit hinein, und mein Arbeitsverzeichnis lag unterhalb eines entsprechend benannten Ordners.

Beide Buildskripte bauen jetzt mit `-trimpath`. Nachkontrolle: null Treffer in der Exe. Das ist auch sonst die bessere Einstellung — es hält lokale Pfade aus dem Ergebnis heraus.

Der Ordner **„Bambu Drucker überwachung"** heißt allerdings weiterhin so; solange dort gebaut wird, bleibt `-trimpath` die einzige Absicherung.

---

## 9. Version 1.1.0

### Löschen gebündelt

Das Python-Hilfsskript nimmt jetzt mehrere Pfade über die Standardeingabe entgegen (`deletemany`). Damit werden alle Dateien eines Druckers in **einer** FTP-Sitzung erledigt statt in einer pro Datei. Bei 126 Dateien auf 42 Druckern sind das 42 Anmeldungen statt 126.

Ein Fehlschlag bei einer Datei bricht die übrigen nicht ab, jede bekommt ihr eigenes Ergebnis. Geprüft mit einer FTP-Attrappe: vier Pfade, einer verweigert — die anderen drei wurden gelöscht, der vierte kam mit dem Grund zurück.

### Speicher durchsehen

Neuer Knopf im Datei-Reiter. Er listet je Ordner auf, wie viele Dateien darin liegen und wie viel Platz sie belegen, zusammengefasst über alle Drucker im aktuellen Filter — mit Fortschritt und abbrechbar wie die Suche.

**Bewusst noch ohne Löschknopf.** Ich kenne die Ablage auf deinen Geräten nicht, und ein falscher Pfad löscht im Zweifel Druckdateien statt Videos. Sag mir nach dem ersten Durchlauf, welche Ordner der Löschknopf anfassen darf.

Auch die Ordnersuche ist gegen eine Attrappe geprüft: Unterordner werden erfasst, die Tiefenbegrenzung wird eingehalten.

### Kammerbeleuchtung blinkt bei Fehler

X1E und X2D haben keine eigene Signalleuchte. Bei diesen Modellen wird die Kammerbeleuchtung zum Blinken gebracht, sobald ein Fehler ansteht — `print_error`, eine HMS-Meldung oder der Zustand `FAILED`.

Der Drucker kann das selbst: `ledctrl` mit `led_mode: "flashing"` und je 1000 ms an und aus. Deshalb wird nicht im Sekundentakt gefunkt, sondern alle 20 Sekunden aufgefrischt — so überlebt das Blinken auch einen Neustart des Druckers. Ist der Fehler weg, wird die Beleuchtung genau einmal auf „an" zurückgesetzt.

Welche Modelle blinken, steht in der Konfiguration (`blink_models`, Vorgabe X1E und X2D); H2D und H2C bleiben in Ruhe, weil sie eigene Signale haben. Abschaltbar über `error_blink`.

### Kleinere Änderungen

Das helle Schema ist die Vorgabe. Der Abschnitt „Farben" in den Einstellungen ist ersatzlos weg — er arbeitete ohnehin gegen die beiden Schemata. Bei go2rtc.yaml steht jetzt der Speicherort mit einem Knopf, der den Ordner im Explorer öffnet; ein `file://`-Link würde von der Seite aus blockiert.

Das große Einfügefeld beim CSV-Import ist raus. Der Datei-Import läuft dafür jetzt direkt: Datei auswählen, fertig — vorher landete der Inhalt erst im Textfeld und musste mit einem zweiten Klick bestätigt werden.

Beim Verlassen einer Vollbildansicht laufen wieder alle sichtbaren Kacheln an. Vorher blieben Kacheln dunkel, die vorher auf „idle" standen.

Beim Umbenennen eines Druckers schlägt der neue Name überall durch — Kachel, Liste, Dateisuche. Weil sich der go2rtc-Streamname aus dem Druckernamen ableitet, wird eine laufende Videoverbindung dabei neu aufgebaut; sonst zeigte sie auf einen Namen, den go2rtc nicht mehr kennt.

Die Schichtangabe ist aus der Kachelzeile verschwunden.

### Was ich nicht bauen konnte: Umbenennen auf dem Gerät

Du wolltest, dass der Name auch auf den Drucker geschrieben wird. Dafür habe ich keinen Weg gefunden. Der Gerätename wird über Bambu Studio beziehungsweise die Handy-App gesetzt und läuft über die Cloud; im lokalen MQTT-Protokoll ist kein Befehl dafür dokumentiert. Intern identifiziert sich der Drucker ohnehin über die Seriennummer, der Name ist reine Anzeige.

Ich habe deshalb bewusst nichts gebaut, das still nichts tut. Wenn dir ein Befehl dafür bekannt ist, baue ich ihn gern ein.

---

## 10. Version 1.2.0

### Dateisuche zeigte Treffer mehrfach

`pollSearch` fragt im 500-ms-Takt den Zwischenstand ab und schickt mit, wie viele Treffer es kennt. Die Funktion ist asynchron und hatte **keine Sperre gegen Überlappung**: dauerte eine Abfrage länger als der Takt, startete die nächste mit demselben Zählerstand. Der Server lieferte pflichtgemäß dieselben Treffer erneut, und beide Antworten wurden angehängt.

Behoben mit einer Sperre plus einer Ablage nach Schlüssel aus IP und Dateiname statt einer Liste — derselbe Treffer kann jetzt gar nicht mehr doppelt auftauchen.

**Nachgestellt** mit acht gleichzeitigen Abfragen bei künstlich verzögerter Antwort:

| | Abfragen | angezeigte Treffer | tatsächlich verschieden |
|---|---|---|---|
| vorher | 8 | **24** | 3 |
| nachher | 1 | 3 | 3 |

### Hängende Drucker bremsen die gesunden nicht mehr

Es dürfen nur vier Transcodes gleichzeitig laufen. Ein Drucker, der nicht antwortet, belegte einen dieser Plätze bis zum Ablauf der Zeitgrenze von 20 Sekunden — bei mehreren solchen Geräten kamen die gesunden praktisch nicht mehr dran.

Zwei Änderungen: das Zeitfenster ist auf 8 Sekunden verkürzt, und jeder Stream bekommt nach einem Fehlschlag eine Pause, die mit jedem weiteren wächst (15 s, 30 s, 60 s, höchstens 2 min). Während der Pause wird go2rtc **gar nicht erst gefragt**, der Platz bleibt frei. Die Oberfläche erfährt über eine Kopfzeile, wie lange pausiert wird, und fragt so lange auch nicht nach. Ein erfolgreicher Abruf hebt die Pause sofort auf.

**Nachgemessen:** acht kaputte Drucker in der Pause, danach ein gesunder — der wurde nach **0,16 ms** bedient statt hinter den anderen zu warten.

### Logo fest im Programm

Dein SVG steckt per `go:embed` in der Exe, ersetzt das Symbol in der Kopfzeile und dient als Favicon. Der Inkscape-Ballast ist raus, die Datei von 13,5 auf 11,4 KB geschrumpft.

Das Logo ist schwarz auf weißer Fläche. Im dunklen Schema würde daraus ein weißer Kasten, deshalb wird es dort invertiert. Ob dir das so gefällt, siehst du am besten selbst — sonst wäre eine zweite Fassung des SVG die sauberere Lösung.

### go2rtc unter den Komponenten

Die beiden getrennten Karten sind zusammengelegt: Laufzustand, Port, „Prüfen" und „Neu starten" stehen jetzt als eingerückter Unterpunkt in der Komponenten-Karte.

Die Zeile **„go2rtc ist eingebettet — wird automatisch gestartet"** ist weg. Sie stimmte seit dem Umbau auf Nachladen nicht mehr, ich hatte sie übersehen. Stattdessen steht dort jetzt, was tatsächlich passiert.

### Update für die Komponenten

Neuer Knopf „Auf Updates prüfen", daneben „Komponenten aktualisieren", sobald etwas verfügbar ist. Beim Start wird still geprüft.

Für go2rtc wird das neueste GitHub-Release abgefragt und mit der Version verglichen, die das installierte Programm selbst meldet. Für ffmpeg geht das nicht — BtbN baut unter einem rollenden Tag ohne Versionsnummer. Dort wird beim Installieren die Prüfsumme des Archivs gemerkt und später mit der veröffentlichten verglichen.

**Zur Absicherung:** die Erstinstallation von go2rtc hängt weiterhin an einer fest hinterlegten Prüfsumme. Beim Aktualisieren auf die jeweils neueste Fassung gibt es die nicht — dort wird stattdessen geprüft, dass die Datei startet und die erwartete Versionsnummer meldet. Das fängt einen kaputten Download ab, keinen übernommenen GitHub-Zugang.

### Farbanpassung restlos entfernt

Die Karte war schon in 1.1.0 weg, die Funktionen aber nicht: `applyCustomColor`, `resetColors`, `loadSavedColors` und `adjustColor` standen noch im Code, und `loadSavedColors()` lief bei jedem Start. Früher gespeicherte Farben hätten das helle Schema also weiterhin überschrieben. Alles entfernt, gespeicherte Werte werden beim Start einmalig weggeräumt.

### Zur Versionsanzeige

Die war korrekt. Nachgesehen in beiden Programmdateien: die im Ordner `Druckerfarm` trägt 1.0.0, die in `Druckerfarm-1.1.0` trägt 1.1.0. Du hattest die ältere gestartet — was auch erklärt, warum dort noch die alte Einbettungs-Zeile stand.

---

## 11. Version 1.3.0

### Warum das Blinken weiterlief

Zwei Ursachen, beide passen zu „bei einigen, nicht bei allen".

**Alte Störungsmeldungen blieben hängen.** Der Drucker schickt das `hms`-Feld nicht bei jeder Nachricht mit. Meine Auswertung ersetzt die Liste nur, wenn das Feld vorhanden ist — sonst hätte jede beliebige Statusmeldung eine bestehende Störung gelöscht. Folge: quittiert man die Störung am Gerät und setzt den Druck fort, schickt der Drucker fortan Nachrichten **ohne** `hms`, und die alte Meldung stand für immer im Speicher. Für die App blieb der Drucker gestört.

Jetzt gilt: wechselt der Zustand von Pause oder Fehler zurück auf **RUNNING**, werden Störungsmeldungen und Fehlernummer verworfen — die Störung ist damit quittiert. Tritt ein Fehler während eines laufenden Drucks auf, bleibt er erhalten (eigener Test für beide Fälle).

**Die App vergaß nach einem Neustart, wer blinkt.** Der Rücksetz-Weg merkte sich nur, welche Drucker *diese* Programminstanz zum Blinken gebracht hatte. Nach einem Neustart wusste sie nichts davon und hätte nie wieder zurückgeschaltet. Die Liste steht jetzt in der Konfiguration (`blinking_ips`).

Neu ist außerdem eine eigene Karte **„Sonderbeleuchtung bei Fehler"** in den Einstellungen. Beim Ausschalten wird das Licht bei allen betroffenen Geräten sofort auf normal zurückgesetzt.

### Inline-CSS

Von **286 Stellen auf 20**:

| | vorher | nachher |
|---|---|---|
| `style`-Attribute im Markup | 111 | 0 |
| `style`-Attribute in erzeugtem Markup | 76 | 16 |
| Stil-Zuweisungen im Skript | 96 | 4 |
| `cssText` | 3 | 0 |

Die verbliebenen 20 sind keine Gestaltung, sondern Laufzeitwerte: die 16 Attribute setzen ausschließlich **CSS-Variablen** (`--w` für Fortschrittsbalken, `--mc` für Modellfarben, `--tray` für Filamentfarben, `--cols` für die Rasterbreite), die eigentliche Gestaltung steht im Stylesheet. Die 4 Zuweisungen positionieren die beiden aufklappbaren Menüs, was zwangsläufig gerechnet werden muss.

Dabei sind rund 90 benannte Klassen entstanden. Für Sichtbarkeit gibt es jetzt durchgängig `setShown()` mit einer `hidden`-Klasse statt `style.display`.

### Die übrigen Punkte

Farbschema: `.tile` bekommt im hellen Schema `--seg-bg`, `--green` steht auf `#00c750`, `.top-bar` und `.stream-bar` liegen auf `--seg-bg` ohne Rahmen und Außenabstand, `.sbtn.s-play` und `.hbtn.primary` schreiben in Weiß.

Aufbau: `top-bar` steht jetzt unter `stream-bar`, die Unterleiste hat die id `sub-nav` und liegt auf `--seg-bg`, nach `lbl-filter` folgt ein `stream-divider`. Neben dem Logo steht **Switzerland** — bewusst außerhalb der Sprachtabelle, sonst würde es beim Umschalten auf Englisch mit übersetzt.

Setup-Reiter: der go2rtc-Laufzustand hängt jetzt direkt am go2rtc-Eintrag der Komponentenliste, samt klickbarem Verweis auf die `go2rtc.yaml`. Die beiden Karten „go2rtc herunterladen" und „go2rtc.yaml" sind ersatzlos weg. Bei der Programmversion lässt sich die vollständige GitHub-Adresse eintragen, das Token-Feld ist entfallen — **damit funktionieren nur öffentliche Repositories**. „Daten liegen in" heißt jetzt „Lokaler Speicherort". Am Ende steht die Copyright-Karte.

Drucker: im Bearbeiten-Modus ist auch das Modell änderbar (die Modell-Kennzeichen werden danach neu gezeichnet), und IP-Adresse, Access Code und Seriennummer tragen ein Pflicht-Sternchen. Die Seriennummer wird jetzt auch tatsächlich verlangt — vorher stand sie im Formular, wurde aber nicht geprüft.

---

## 12. Version 1.3.1 — Absturz beim Start

Von mir verursacht, in 1.3.0. Beim Verschieben des go2rtc-Zustands in die Komponentenliste habe ich übersehen, dass diese Liste erst **zur Laufzeit** gezeichnet wird. `chkG2()` läuft aber schon beim Start und danach alle 15 Sekunden — und griff ungeprüft auf `sind`, `stitle` und `sdesc` zu. Beim ersten Durchlauf gibt es die noch nicht:

```
Uncaught (in promise) TypeError: Cannot set properties of null (setting 'className')
    at chkG2
```

Weil die Ausnahme mitten in der Funktion auftrat, brach auch alles Nachfolgende ab — unter anderem das Wiederanlaufen der Streams nach einer go2rtc-Unterbrechung.

**Beim Suchen nach der Ursache fielen weitere Blindgänger auf.** Eine systematische Durchsicht aller Zugriffe der Form `getElementById(…).eigenschaft = …` ergab **34 Stellen**, deren Element beim Start fehlen kann. Ein Teil davon war schon vorher da und nicht von mir: `g2dot` etwa gibt es im Markup überhaupt nicht mehr, und ein knappes Dutzend Elemente der Sync-Ansicht ebenso wenig. Letztere fielen nie auf, weil sie in `try/catch` liefen und still nichts taten — `updateSyncUI` allerdings nicht, dort hätte es ebenfalls abbrechen können.

Alle 34 laufen jetzt über kleine Helfer (`setText`, `setVal`, `setCls`, `setDis`, `byId`), die nichts tun, wenn es das Element nicht gibt. Nachkontrolle: **0 ungeschützte Zugriffe**.

**Nachgestellt** mit einem Prüfstand, der nur die Elemente bereitstellt, die beim Start tatsächlich im Markup stehen, und dann `chkG2()` aufruft — einmal mit erreichbarem go2rtc, einmal ohne:

| | go2rtc erreichbar | go2rtc weg |
|---|---|---|
| 1.3.0 | **Absturz** | **Absturz** |
| 1.3.1 | läuft | läuft |

Der zweite Fall ist der wichtigere: dort wäre der Absturz auch dann gekommen, wenn die Komponentenliste längst gezeichnet war — `g2dot` fehlt ja dauerhaft.

---

## 13. Version 1.3.2 — die Snapshot-Meldung war geraten

Die Kachel meldete „Stream offline oder **ffmpeg fehlt im PATH**", obwohl ffmpeg auf deinem Rechner nachweislich installiert ist. Der Text war eine Vermutung, keine Feststellung — und hat damit in die falsche Richtung geschickt.

**Warum das so war.** go2rtc antwortet auf `/api/frame.jpeg` mit **HTTP 200 und leerem Rumpf**, wenn es kein Bild erzeugen kann. Den Grund behält es für sich; er steht nur in seinem eigenen Protokoll. Nachgestellt mit einem Stream auf eine nicht erreichbare Adresse:

```
Antwort:  http=200  size=0  zeit=0,0004 s
go2rtc-Log: error="streams: dial tcp 10.255.255.10:322: connect: network is unreachable"
```

Die Antwort kommt also **sofort** — es ist keine Zeitüberschreitung, go2rtc weiß bereits, dass es nicht geht.

**Jetzt wird nachgesehen statt geraten.** Bei leerem Rumpf prüft die App der Reihe nach:

1. Kennt go2rtc den Stream überhaupt? `/api/streams?src=…` antwortet mit 404, wenn er nicht in der Konfiguration steht — dann ist die Konfiguration veraltet und go2rtc muss neu starten.
2. Ist ffmpeg tatsächlich da? Nur dann wird es überhaupt erwähnt.
3. Antwortet die Kamera des Druckers auf Port 322? Dabei wird der MQTT-Zustand mit einbezogen — „antwortet per MQTT, die Kamera aber nicht" ist eine andere Lage als „Gerät ganz aus".
4. Steht die Verbindung und kommt trotzdem kein Bild, ist meist der Zugangscode falsch.

Statt einer pauschalen Vermutung steht jetzt zum Beispiel:

> `woobly 10 antwortet per MQTT, die Kamera auf Port 322 aber nicht — Kamera eingeschaltet?`

> `Stream "woobly-10" steht nicht in der go2rtc-Konfiguration — go2rtc neu starten`

**Zu deinem konkreten Fall.** Welche der vier Ursachen bei `woobly-10` zutrifft, sagt dir die neue Meldung beim nächsten Versuch. Mein Verdacht liegt auf Punkt 1 oder 3 — ffmpeg scheidet aus, das läuft bei den anderen Druckern ja.

Zwei ältere Tests mussten nachgezogen werden: einer zählte alle Anfragen an go2rtc, jetzt zählt er nur die Bildabrufe (die Diagnose fragt zusätzlich den Streamzustand ab); der andere prüfte auf den alten Wortlaut und stellt nun sicher, dass die Pauschalmeldung **nicht** zurückkommt.

---

## 14. Version 1.3.3 — go2rtc war weg und kam nicht wieder

`ERR_CONNECTION_REFUSED` auf Port 1984 heißt: go2rtc läuft nicht. Zwei Lücken erklären, warum es weg war **und warum es wegblieb**.

### Nichts hat go2rtc je wieder gestartet

Der Prozess wurde mit `go cmd.Wait()` nur abgeräumt — endete er, passierte schlicht nichts mehr. Kein Neustart, kein Hinweis, nur eine Oberfläche ohne Bilder, bis jemand von Hand „Neu starten" drückt oder die App neu startet.

Das ist keine graue Theorie: go2rtc ist mir in den eigenen Belastungstests reproduzierbar weggestorben, als zu viele Snapshot-Anfragen gleichzeitig liefen.

Jetzt wacht eine Überwachung darüber. Sie unterscheidet einen Absturz von einem gewollten Stopp und startet nur im ersten Fall neu — mit wachsender Wartezeit (1 s, 2 s, 4 s … höchstens 1 min), damit ein dauerhaft kaputtes go2rtc nicht im Sekundentakt gestartet wird. Lief es länger als eine Minute, gilt der Zähler als zurückgesetzt. Dazu alle 30 Sekunden eine Erreichbarkeitsprüfung auf dem Port, die auch einen hängenden Prozess erwischt.

### Eine gescheiterte Komponenten-Aktualisierung ließ go2rtc gestoppt zurück

Das war mein Fehler aus 1.2.0. Zum Ersetzen der Datei muss go2rtc beendet werden — das passiert **vor** dem Herunterladen und Prüfen. Schlug danach irgendetwas fehl, kehrte die Funktion mit einer Fehlermeldung zurück und übersprang den Neustart. go2rtc blieb für immer aus, obwohl die alte Datei unversehrt an ihrem Platz lag.

Der Neustart läuft jetzt in **jedem** Fall, auch nach einem Fehlschlag.

### Nachgewiesen

Mit einem Programm, das sich als go2rtc ausgibt, seine Starts mitzählt und sich nach kurzer Zeit selbst beendet:

| Fall | Ergebnis |
|---|---|
| Absturz | wird neu gestartet (2 Starts in 3 s) |
| gewollter Stopp | bleibt aus (1 Start) |
| Aktualisierung schlägt fehl | läuft weiter (1 → 2 Starts), Fehler wird trotzdem gemeldet |

### Was du jetzt tun kannst

Die neue Fassung startet go2rtc beim Programmstart und hält es am Leben. Falls es sofort wieder abstürzt, steht der Grund in `druckerfarm.log` im Datenordner — dort taucht dann alle paar Sekunden „go2rtc unerwartet beendet" auf. Schick mir in dem Fall die Zeilen, dann sehen wir weiter.

---

## 15. Version 1.4.0 — Schluss mit dem Raten

Ich habe bei go2rtc mehrfach auf Verdacht repariert. Das hört hier auf: **die Ausgabe von go2rtc wurde bisher weggeworfen** (`cmd.Stdout = nil`). Endete oder verweigerte es den Dienst, stand nirgends warum — weder für dich noch für mich.

### go2rtc schreibt jetzt mit

Alles, was go2rtc ausgibt, landet in `go2rtc.log` im Datenordner. Nachgewiesen mit einem Programm, das sich beim Beenden erklärt: die Zeile steht anschließend in der Datei.

### Ein doppelter Streamname legt alles lahm

Beim Nachmessen am echten go2rtc gefunden — und es ist schlimmer als erwartet:

| go2rtc.yaml | Ergebnis |
|---|---|
| ein Schlüssel doppelt | **`{}` — kein einziger Stream**, go2rtc läuft aber weiter |
| sauber | alle Streams geladen |

go2rtc verwirft bei einem doppelten Schlüssel die **gesamte** streams-Sektion, meldet das nur als Warnung in seinem Log — und läuft weiter, als wäre nichts. Von außen sieht das aus wie „go2rtc läuft, aber nichts funktioniert".

`buildYaml` hatte dagegen keinen Schutz. Zwei Drucker, deren Namen sich nur in Schreibweise oder Satzzeichen unterscheiden — „woobly 1" und „Woobly-1" — ergeben denselben Streamnamen. **Eine einzige Umbenennung hätte gereicht, um alle 42 Kameras auf einen Schlag stillzulegen.**

Namen werden jetzt eindeutig gemacht (`woobly-1`, `woobly-1-2`, …) und die Kollision protokolliert. Gegen das echte go2rtc geprüft: dieselben vier Drucker, die vorher `{}` ergaben, liefern jetzt vier Streams.

**Bei dir liegt es nicht daran** — ich habe deine 42 Druckernamen durchgerechnet, 42 verschiedene Streamnamen, keine Kollision. Der Schutz bleibt trotzdem drin.

### Gegenprüfung nach dem Start

Nach jedem Start vergleicht die App, wie viele Drucker eingetragen sind und wie viele Streams go2rtc daraus gemacht hat. Weicht das ab, steht es im Log:

```
❌ go2rtc kennt KEINEN Stream, obwohl 42 Drucker eingetragen sind — go2rtc.yaml fehlerhaft
```

### Diagnose-Knopf

Neu in den Einstellungen unter Komponenten. Ein Klick zeigt auf einen Blick: ist go2rtc da und in welcher Fassung, soll es laufen, läuft es, unter welcher PID, wie oft wurde es schon neu gestartet, ist der Port erreichbar, wie viele Streams kennt go2rtc gegenüber wie vielen Druckern, gibt es Namenskollisionen — dazu die letzten 40 Zeilen aus `go2rtc.log` und 25 aus `druckerfarm.log`. Mit „In die Zwischenablage" schickst du mir alles in einem Stück.

**Zugangscodes stehen nicht drin** — eigener Test, der genau das prüft.

---

## 16. Version 1.4.1 — die Snapshot-Last war zu hoch

Deine Meldung `wsarecv: An existing connection was forcibly closed` sagt etwas Genaueres als die vorherigen: go2rtc **nimmt die Anfrage an und stirbt mitten in der Bearbeitung**. Nicht „nicht erreichbar", sondern Absturz während der Bilderzeugung.

Nachgerechnet, was meine Einstellungen bei dir erzeugen:

```
42 Drucker, alle 2 s ein Bild   →  21 Anfragen pro Sekunde
jede startet in go2rtc einen eigenen ffmpeg-Vorgang
gedeckelt auf 4 gleichzeitig, das schafft etwa 4 Anfragen/s
                                →  dauerhafter Überhang von 17 Anfragen/s
```

Das ist keine Randlage, das ist der Dauerzustand seit dem Umbau auf 2 s als Vorgabe. Ich hatte die Parallelität begrenzt, aber nicht die **Rate** — der Stau wächst einfach vor der Sperre.

**Das Intervall richtet sich jetzt nach der Druckerzahl.** Höchstens 3 Bildanfragen pro Sekunde insgesamt; daraus ergibt sich der kleinstmögliche Abstand je Drucker:

| Drucker | tatsächliches Intervall bei Wunsch „2 s" |
|---|---|
| 1 | 2 s |
| 6 | 2 s |
| **42** | **14 s** |

Stürzt go2rtc trotzdem ab, wird weiter zurückgefahren — nach drei Abstürzen sind es 56 s statt 14 s, bis es sich beruhigt hat.

Die Oberfläche zeigt neben den Modusknöpfen an, was wirklich gilt („tatsächlich alle 14 s"), und passt ihren Abfragetakt daran an. Sonst würden 42 Kacheln weiter im 2-Sekunden-Takt gegen einen Server anfragen, der ohnehin nur alle 14 s ein neues Bild hat.

**Was das für dich heißt:** 2 s sind bei 42 Druckern über diesen Weg nicht zu haben — go2rtc erzeugt jedes Bild einzeln über ffmpeg. Willst du kürzere Abstände, wäre der Videomodus der richtige Weg; der hält eine Verbindung offen, statt für jedes Bild einen Vorgang zu starten.

**Was ich noch nicht weiß:** *warum* go2rtc unter der Last abstürzt statt Anfragen abzulehnen. Das steht jetzt in `go2rtc.log` — schick mir die Datei oder die Diagnose, dann ist auch das geklärt. Mein Testrechner ist zu knapp bemessen, um das nachzustellen; dort stirbt go2rtc schon beim Aufbau von zwölf Teststreams.

**Nebenbei bestätigt:** die Überwachung aus 1.3.3 arbeitet — go2rtc kommt zurück, auch wenn du es im Taskmanager beendest. Wenn du es bewusst anhalten willst, geht das über den Schalter in den Einstellungen, nicht über den Taskmanager.

---

## 17. Version 1.4.2 — mein Fehler: sieben von acht Ansichten waren unsichtbar

Du hattest recht, und ich habe an der falschen Stelle gesucht. Der Grund, warum „gar nichts mehr lädt", liegt nicht bei go2rtc, sondern in **meiner Umstellung von heute früh**.

Beim Überführen des Inline-CSS habe ich 48 Stellen von `element.style.display = …` auf eine `hidden`-Klasse umgestellt. Was ich übersehen habe: `.view` ist im Stylesheet grundsätzlich auf `display:none` gesetzt. Vorher schrieb `sv()` beim Anzeigen `style.display='block'` direkt ans Element — ein Inline-Stil, der die Regel übersteuert. Danach entfernte `setShown()` nur noch die `hidden`-Klasse. **Die Ansicht blieb damit auf `display:none` und wurde nie sichtbar.**

Dass du überhaupt noch etwas gesehen hast, war Zufall: `view-uebersicht` trug zusätzlich die Klasse `col-grow`, und die setzt `display:flex`. Alle anderen — Einstellungen, Drucker, Dateien, das Raster selbst, beide Vollbildansichten — blieben leer.

**Nachgerechnet** mit einem Prüfstand, der die echten CSS-Regeln aus der Datei liest und für jede Ansicht ausrechnet, welchen `display`-Wert sie nach dem Umschalten hat:

| | sichtbar | unsichtbar |
|---|---|---|
| 1.4.1 | 1 | **7** |
| 1.4.2 | 8 | 0 |

Behoben, indem `setShown()` zusätzlich eine `shown`-Klasse setzt und es dafür Regeln gibt (`.view.shown`, `.sd-files.shown`, `.ir.shown`). Damit ist die Sichtbarkeit vollständig im Stylesheet beschrieben, ohne Inline-Stile. Die widersprüchliche `col-grow`-Klasse an `view-uebersicht` ist weg.

**Der Prüfstand bleibt.** Er hätte den Fehler sofort gefunden — ich hatte nach der Umstellung nur die Syntax geprüft, nicht die Wirkung.

### Ladeanzeige und Gerüst

Zwei Dinge, die du zu Recht vermisst hast:

Ein **Kreisel in der Streamleiste**, solange im Hintergrund etwas läuft — Statusabfragen, Bildabrufe, Dateisuche. Und **je Kachel ein Kreisel**, solange noch kein Bild angekommen ist. Bei 42 Druckern mit 14 Sekunden Abstand vergeht sonst viel Zeit, in der die Oberfläche tot wirkt, obwohl sie arbeitet.

Leere Bereiche zeigen jetzt ein **Gerüst** statt gar nichts — mit Kreisel, solange noch keine Drucker geladen sind, und einem Hinweistext, wenn ein Filter nichts übrig lässt.

### Zu den Zeitüberschreitungen im Log

Die 14 Einträge aus deinem Auszug sind ausnahmslos `dial tcp <IP>:322` — dreizehnmal ohne jede Antwort, einmal aktiv abgewiesen. **Kein Absturz.** Das sind 14 verschiedene Drucker, deren Kamera-Port nicht antwortet, und es passt zu den 13 von 42 nicht erreichbaren Geräten bei der Dateisuche. Ob die Geräte aus sind, die Kamera abgeschaltet ist oder das Netz dazwischenfunkt, sagt das Log nicht — aber es ist ein Zustand der Farm, kein Fehler der App.

---

## 18. Version 1.5.0 — deine Liste mit 24 Punkten

### Beschriftungen

| vorher | jetzt |
|---|---|
| `Video Bridge` (englische Oberfläche) | `Settings` — die deutsche hieß schon immer „Einstellungen" |
| Alle starten / Alle stoppen | **Video starten / Video anhalten** (englisch: Start Video / Pause Video) |
| Ansicht | **Profile** |
| `&#9745;` als roher Text im Knopf | ☑ — der Knopf wurde per `textContent` gesetzt, dabei bleibt eine HTML-Entität stehen, statt sich in ein Häkchen zu verwandeln |
| Alle laden | **Inhalte laden** |
| ⛶ Minimal | **⛶ Fullscreen Video Only** |

### Der File-Sync-Reiter wurde nie markiert

Der Reiter trug die Kennung `tab-sync2`, gesucht wurde beim Umschalten aber nach `tab-sync`. Also fand `sv()` nichts und setzte die Markierung nirgends. Umbenannt.

### Nur ein Drucker lud seine Dateien — zwei Fehler übereinander

`loadAllSDFiles` gab es **zweimal** im Quelltext. Die zweite, später stehende Fassung gewinnt in JavaScript — und die feuerte alle Abfragen gleichzeitig los. `loadSDFiles2` arbeitet aber als Ziehharmonika: es klappt zuerst *jede* offene Liste zu und dann die eigene auf. Alle Aufrufe schlossen sich also gegenseitig, am Ende blieb genau einer stehen.

Behoben: die doppelte Fassung ist weg, und `loadSDFiles2` bekommt ein Argument `keepOthers`. Der Einzelklick bleibt eine Ziehharmonika, das Sammelladen lässt alle offen.

Dazu, wie gewünscht: **„Inhalte laden" arbeitet nur noch die angehakten Drucker ab.** Ist keiner angehakt, kommt „Drucker wählen" statt eines stillen Nichtstuns. Dasselbe gilt für die Dateisuche — die lief vorher über den Modellfilter, also über alle 42 Geräte, obwohl vielleicht nur drei ausgewählt waren.

### Der Python-Fehlertext

Bei einer Zeitüberschreitung stand der vollständige Traceback in der Oberfläche — sechs Zeilen Technik, aus denen nicht hervorging, dass schlicht der Drucker nicht antwortet. Es gibt jetzt eine Übersetzung:

| was das Skript meldet | was du liest |
|---|---|
| `TimeoutError: [WinError 10060] …` | Zeitüberschreitung — Drucker antwortet nicht auf Port 990 |
| `ConnectionRefusedError: [WinError 10061] …` | Verbindung abgelehnt — FTP am Drucker aus oder falscher Port |
| `OSError: [Errno 113] No route to host` | Drucker nicht erreichbar — im Netz nicht auffindbar |
| `ftplib.error_perm: 530 Login incorrect` | Zugangscode wird nicht angenommen |

Unbekanntes bleibt erhalten, aber nur die letzte Zeile und auf 160 Zeichen gekürzt.

### Der Upload läuft weiter — und lässt sich anhalten

Der Upload lief bisher als reine Schleife im Klick-Handler. Verließ man die Ansicht und kam zurück, wurde die Druckerliste neu aufgebaut: **Auswahl weg, Fortschritt weg.** Anhalten oder abbrechen ging gar nicht.

Jetzt steht der Zustand in einem Objekt, und die Anzeige wird daraus gezeichnet. Ein Ansichtswechsel malt sie einfach neu. Dazu **⏸ Pause** (hält zwischen zwei Dateien an, ▶ setzt fort) und **✕ Abbrechen**. Die Auswahl der Drucker übersteht den Neuaufbau ebenfalls.

Nachgemessen mit einem Prüfstand ohne Browser:

| Fall | Ergebnis |
|---|---|
| Ansichtswechsel mitten im Lauf | Fortschrittszeilen wieder da, 4 von 4 Dateien hochgeladen |
| Abbrechen nach der zweiten Datei | Lauf endet bei 2 von 4, Vermerk „abgebrochen" |
| Pause, 60 ms warten, fortsetzen | Zähler bleibt bei 2 stehen, läuft danach auf 4 |

### Streams nur dort, wo man sie sieht

Beim Wechsel in File Sync, Drucker oder Einstellungen werden jetzt **alle Verbindungen und Bildabfragen beendet**; beim Zurückkehren in die Übersicht laufen sie wieder an. Vorher liefen 42 Verbindungen im Hintergrund weiter, während man Dateien hochlud — genau dann, wenn die Bandbreite anderswo gebraucht wird.

Die Listenansicht zählt dazu: dort gibt es keine Bilder, also auch keine laufenden Streams und keine Streamleiste.

### Neuer Filter „Pausiert"

Steht nach „Offline" und zeigt, was pausiert, abgebrochen oder mit Druckfehler dasteht. Der Zähler wird zusätzlich alle vier Sekunden nachgerechnet — er hängt am Zustand der Maschine, nicht an einer Bedienhandlung, und wäre sonst bis zur nächsten Neuzeichnung blind.

Der **Favoritenzähler** wird jetzt beim Umschalten sofort mitgeführt; vorher fehlte der Aufruf.

### Warum das Blinken nicht aufhörte

Beim Fortsetzen eines Drucks wurden die alten Störungsmeldungen verworfen — das stand schon drin. Es half nur nichts, denn der **X1 schickt bei jeder Statusmeldung die vollständige HMS-Liste mit**, auch dann noch, wenn der Druck längst weiterläuft und die Meldung am Gerät nur nicht quittiert wurde. Kaum geleert, war die Liste wieder voll.

Zugleich hattest du früher verlangt, dass eine Störung **während** eines laufenden Drucks weiter blinkt. Beides pauschal geht nicht — also merkt sich die App jetzt, welche Meldungen im Moment des Anlaufens bereits anstanden. Kommen genau die per Statusmeldung wieder herein, lösen sie nichts aus. Etwas Neues während des Drucks sehr wohl. Das Gedächtnis wird geleert, sobald der Druck endet.

Der Prüftakt ging von 20 auf **5 Sekunden** herunter; nach dem Fortsetzen blinkte die Kammer sonst noch fast eine halbe Minute weiter.

### Darstellung

* Die Streamleiste hat keine abgerundeten Ecken mehr und dafür eine Trennlinie nach unten.
* Der Hinweis „tatsächlich alle 14 s" und der Ladekreisel standen als Text **in** der Leiste — bei jeder Änderung sprang sie um eine Zeile. Beide sind raus. Stattdessen läuft ein Rahmen um den aktiven Modusknopf, solange etwas arbeitet, und der Hinweis steht als Kurzinfo am Knopf.
* Die **Liste hat jetzt dieselbe Filterleiste wie das Grid**. Dafür sind Filterleiste und Streamleiste aus der Grid-Ansicht herausgezogen und eine Ebene höher gehängt; die Streamleiste blendet sich in der Liste aus.
* `list-header` ist ersatzlos weg — er stand doppelt neben `list-thead`.
* Die Spalten **Status** und **Aktionen** sind aus der Liste verschwunden, ebenso der Aktionsblock in jeder Zeile. **Bearbeiten geht per Doppelklick auf die Zeile** — sonst wäre der Weg dorthin ganz verloren gewesen.
* Zeilenhintergrund auf `var(--seg-bg)`.
* Die aktive Pille nutzt jetzt durchgehend Akzentfarbe mit gedämpftem Hintergrund und fettem Text.

Nachgerechnet: Rasterspalten im CSS **7**, Kopfzellen **7**, Zellen je Zeile **7** — vorher standen 9 Spalten im CSS gegen wechselnde Zellenzahlen, weil die LIVE-Marke nur manchmal gerendert wurde.

### „Speicher durchsehen" ist raus

Vollständig: Knopf, Anzeige, JavaScript, HTTP-Endpunkte, `runStorageScan`, `handleStorageScan`, `handleStorageResult`, `scanStorage` und das zugehörige CSS.

### Netzwerksuche: mehrere auf einmal übernehmen

Die Suche gab es schon, übernehmen ließ sich aber nur ein Gerät, und das auch nur in die manuelle Eingabezeile. Jetzt:

* **Auswahlkästchen je Fund**, dazu „alle neuen auswählen"
* **Zugangscode direkt in der Zeile** — den gibt das Gerät nicht heraus, er muss von Hand kommen. Der Knopf sagt an, bei wie vielen er noch fehlt.
* **Keine Doppelten**: bekannt ist, wessen IP *oder* Seriennummer schon in der Liste steht. Solche Zeilen lassen sich gar nicht erst anhaken, und beim Anlegen wird noch einmal geprüft.
* **Rückfrage zum Namen**: ein Schalter „Gerätenamen aus dem Netz übernehmen" (standardmäßig an) und eine Bestätigung, die die Namen vorher auflistet. Ohne den Schalter vergibt die App Ersatznamen aus Modell und letzter IP-Stelle.

### Zum Gerätenamen — diesmal gründlich nachgesehen

Du hattest widersprochen, als ich das für unmöglich erklärt habe, und ich habe zugesagt, tiefer zu graben. Das Ergebnis bleibt leider dasselbe, aber jetzt mit Belegen statt einer Behauptung.

Das lokale MQTT-Protokoll ist offen dokumentiert. Die vollständige Befehlsliste umfasst `info`, `pushing`, `upgrade`, `print`, `system`, `camera` und `xcam`. Unter `system` stehen genau drei Befehle: `ledctrl`, `get_access_code` und `set_accessories`. **Ein Befehl zum Setzen des Gerätenamens existiert nicht** — auch nicht undokumentiert in den einschlägigen Sammlungen.

Der Name gehört nicht zum Gerät, sondern zur Kontobindung. Deshalb geht Umbenennen ausschließlich über die Hersteller-App oder den Slicer bei bestehender Cloud-Verbindung. Im LAN-Only-Betrieb gibt es keinen Weg dorthin; das ist seit 2023 als Wunsch gemeldet und bis heute offen. Andere Betreiber mit derselben Farmgröße beschreiben genau dein Problem: mehrere Geräte kommen mit identischem Werksnamen.

**Was tatsächlich funktioniert**, sind zwei Umwege — beide erfordern, den LAN-Only-Modus vorübergehend abzuschalten:

1. LAN-Only aus → in der Hersteller-App „Gerät binden" / „Geräte in der Nähe koppeln" → umbenennen → LAN-Only wieder an. Dauert pro Gerät etwa eine Minute; der Name bleibt danach auch im LAN-Betrieb stehen.
2. Derselbe Weg über den Farm-Manager des Herstellers, ohne die Geräte an ein Konto zu binden.

Beides ist Handarbeit am Gerät und lässt sich von hier aus nicht fernsteuern. Was diese App kann, ist der Umkehrweg: sie **liest** den im Gerät hinterlegten Namen per Netzwerksuche aus und bietet an, ihn zu übernehmen — das ist Punkt 23.

Sollte ein Firmware-Stand das einmal öffnen, ist die Stelle vorbereitet: die Befehle gehen alle über dieselbe Funktion in `mqtt.go`.

---

## 19. Version 1.5.1 — kein Bild bei 18 Druckern, und was das Log wirklich sagt

### Der Befund

Aus deinem `go2rtc.log` und der `go2rtc.yaml`:

| Gruppe | gesamt | ohne Bild |
|---|---|---|
| `woobly-h1` … `h15` | 14 | **14** |
| `woobly-1` … `28` | 28 | 4 (Nr. 10, 26, 27, 28) |

Die Konfiguration ist in Ordnung: richtiger Pfad, alle 42 Streams vorhanden, keine doppelten Namen, jede IP genau einmal. Daran liegt es nicht.

Entscheidend ist der Zeitverlauf. Die 18 betroffenen Adressen scheitern **ab der allerersten Anfrage** — 15:24:38 bis 16:57:00, anderthalb Stunden, 260 Versuche, kein einziger Erfolg. Wäre es Last, Verbindungsdruck oder ein Wettlauf im Programm, käme zwischendurch ein Bild durch. Es kommt keines.

Der Fehler steht auf TCP-Ebene: `dial tcp <IP>:322: i/o timeout`. Das heißt, das erste Paket bleibt unbeantwortet — noch bevor irgendein Zugangscode, Streamname oder Programmpfad eine Rolle spielt. Bei `192.168.189.51` steht stattdessen *actively refused*: dort antwortet das Gerät, der Port ist nur zu. **Kein Programmcode kann einen geschlossenen Port öffnen.**

### Was ich verworfen habe

Meine erste Vermutung war, dass jeder Snapshot eine neue Verbindung zur Kamera aufmacht (`/api/frame.jpeg` tut das tatsächlich, anders als der WebRTC-Weg des ersten Builds) und die Geräte das nicht mitmachen. Ich habe angefangen, den Snapshot-Weg auf eine dauerhafte Verbindung umzubauen — und dann den Zeitverlauf ausgewertet. Der widerlegt die Vermutung: bei Verbindungsdruck gäbe es Teilerfolge. Umbau verworfen, bevor er Schaden anrichtet.

### Was jetzt drin ist: die Frage lässt sich selbst beantworten

Neuer Knopf unter *Einstellungen* → **„📡 Drucker erreichbar?"**. Er klopft bei jedem Drucker an drei Ports einzeln an und sagt im Klartext, was Sache ist:

| Port | wofür |
|---|---|
| 322 | Kamera — genau der, an dem go2rtc scheitert |
| 990 | Dateizugriff |
| 8883 | Status und Steuerung |

Daraus wird ein Urteil je Drucker:

* **alle drei zu** → Gerät aus oder nicht im Netz
* **322 abgelehnt, 990/8883 offen** → Drucker läuft, die Kamera ist abgeschaltet → *LAN Mode Liveview* am Gerät einschalten
* **322 stumm, 990/8883 offen** → Drucker läuft, Kamera antwortet nicht → Liveview prüfen, danach Drucker neu starten
* **322 offen** → an der Erreichbarkeit liegt es nicht

Der Unterschied zwischen *abgelehnt* und *keine Antwort* ist dabei die wichtigste Einzelinformation: abgelehnt heißt, das Gerät lebt und nur der Dienst fehlt.

### Was ich in 1.5.0 selbst verschlimmert hatte

Ich hatte den Prüftakt für das Kammerlicht von 20 auf 5 Sekunden gesenkt, damit das Blinken nach dem Fortsetzen schneller aufhört. Dabei habe ich übersehen, dass in derselben Schleife auch **gesendet** wird — das hat die MQTT-Befehle an die Drucker vervierfacht, ohne jeden Nutzen. Das Auffrischen dient nur dazu, einen Neustart des Druckers zu überstehen.

Jetzt gilt: **häufig prüfen, selten senden.** Der Zustand wird weiter alle 5 Sekunden geprüft, ein Auffrischbefehl geht höchstens einmal pro Minute und Drucker raus. Neue Störungen werden unverändert binnen fünf Sekunden bemerkt.

### Startbild

`Rastergrafik2.jpg` steckt fest in der Programmdatei, wird beim Start bildschirmfüllend gezeigt und blendet nach zwei Sekunden weich aus — mit demselben Ladekreisel wie beim Video. Zwei Sicherungen: das Bild wird vorgeladen, damit nicht erst eine schwarze Fläche steht, und nach spätestens vier Sekunden verschwindet das Startbild in jedem Fall. Ein Startbild, das hängen bleibt, wäre schlimmer als gar keines.

Im echten DOM nachgemessen: bei 1601 ms noch sichtbar, ab 2002 ms ausgeblendet, nach 4 s nicht mehr im Dokument, keine Laufzeitfehler.

---

## 20. Version 1.5.2 — Schönheitsfehler und ein schneller Start

### Der zuckende Rahmen

Der laufende Rahmen um den aktiven Modusknopf startete bei **jeder** Abfrage neu. Grund: `paintBusy()` nahm die Klasse erst allen Knöpfen ab und hängte sie sofort wieder an — damit beginnt eine CSS-Animation von vorn. Bei 42 Druckern passiert das im Sekundentakt, und genau das sieht man als Zucken.

Jetzt wird nur noch angefasst, was sich tatsächlich ändert. Nachgemessen im echten DOM: **25 Abfragen hintereinander → 2 Klassenwechsel** (einmal an, einmal aus), der Rahmen läuft ununterbrochen durch. Vorher wären es 50 gewesen. Beim Moduswechsel wird der Merker mitgeführt, sonst bliebe der Rahmen am alten Knopf hängen.

Die Animation läuft außerdem etwas ruhiger: 1,6 s mit weichem Verlauf statt 1 s linear.

### Weitere Punkte

* **„📤 Export"** heißt jetzt **„📤 Druckerliste herunterladen"** (englisch: *Download printer list*).
* **Seriennummer** steht bei *Gespeicherte Drucker* als eigene Spalte vor der IP. Fehlt sie, steht dort ein Gedankenstrich statt einer leeren Lücke. Nachgeprüft: Reihenfolge `Name · Seriennummer · IP`.
* **Erreichbarkeitsliste**: die Portmarken hatten eine feste Spalte von 110 px, brauchten aber mehr — und schoben sich über das Urteil. Jetzt bekommen sie genau so viel Platz, wie sie brauchen (`max-content`), das Urteil den Rest und darf umbrechen; jede Zelle beschneidet sauber statt überzulaufen.
* **Das Zahnrad** hat weder Farbe noch Schein, solange alles läuft — es meldet sich erst, wenn etwas fehlt, und dann in Rot. Ein dauerhaft grün leuchtendes Symbol trägt keine Information.

### Der Start ist jetzt sofort bebildert

Bisher: schwarzes Fenster mit fremdem Symbol, dann irgendwann das Bild. Zwei Ursachen, beide behoben.

**Das Bild kommt mit dem ersten Byte.** Eine 337 Byte kleine, absichtlich unscharfe Vorschau steckt direkt in der Seite — die ist da, sobald überhaupt etwas da ist. Das scharfe Bild liegt als eigene Ebene darüber und wird eingeblendet, sobald es geladen ist. Aus „schwarz, dann Bild" wird also „Bild, dann scharf". Kommt das große Bild gar nicht, bleibt die Vorschau stehen; einen schwarzen Bildschirm gibt es in keinem Fall.

**Das Bild ist kleiner.** Die Vorlage war 3855 × 2168 und 662 KB — mehr als jeder Bildschirm braucht. Jetzt 1920 × 1080 und **187 KB**, also 72 % weniger.

**Das fremde Symbol ist weg.** Die Seite gibt ein Favicon an, das auf das bereits eingebettete Logo zeigt.

Was bleibt: das leere Browserfenster, bevor das erste Byte da ist. Das gehört dem Browser, nicht diesem Programm — es lässt sich verkleinern, nicht wegzaubern.

---

## 21. Version 1.5.3 — das Flackern, die Drucksteuerung, die Detailzeile

### Das Flackern — und warum mein Test es übersehen hat

Du hattest recht, es war nicht behoben. Der Fehler saß eine Ebene tiefer als gedacht, und mein Prüfstand hat mich in Sicherheit gewiegt.

`busyStart` und `busyEnd` bilden **je Bildabfrage ein eigenes Paar**. Zwischen zwei Abfragen fällt der Zähler kurz auf null — und genau dann nahm `paintBusy()` die Klasse ab, um sie beim nächsten Bild sofort wieder anzuhängen. Bei 42 Druckern im 2-Sekunden-Takt passiert das dutzendfach.

Mein Test in 1.5.2 hatte die Aufrufe **ineinander geschachtelt** (25× start, dann 25× ende). Dabei bleibt der Zähler durchgehend über null, nichts wechselt, und der Test meldete brav 2 Klassenwechsel. Im Betrieb laufen sie aber **nacheinander**. Mit diesem Muster nachgemessen:

| | Klassenwechsel bei 30 Abfragen |
|---|---|
| vorher | **60** |
| jetzt | **2** |

Behoben mit einer Nachlaufzeit: der Rahmen geht erst aus, wenn eine halbe Sekunde lang wirklich nichts mehr läuft. Während einer Serie bleibt er durchgehend an.

### Pause, Weiter und Abbrechen taten nichts

Im Befehl fehlte das Feld `param`. Laut Protokoll gehört es zwingend dazu, auch wenn es immer leer bleibt:

```
vorher:  {"print":{"sequence_id":"0","command":"pause"}}
jetzt:   {"print":{"sequence_id":"0","command":"pause","param":""}}
```

Das erklärt das gemeldete Verhalten genau: kein Fehler, keine Rückmeldung, keine Wirkung — der Drucker nimmt die Nachricht entgegen und verwirft sie. Der Test prüft jetzt auf das Pflichtfeld und darauf, dass die Nachricht gültiges JSON ist.

### Detailzeile in der Liste

Kein schwebendes Menü mehr, sondern eine Zeile **direkt unter** der jeweiligen Druckerzeile. Standardmäßig zu. Ein Klick auf die Zeile klappt sie auf, ein Knopf **„▸ Alle ausklappen"** öffnet alle — und zwar genau die, die der Filter gerade übrig lässt.

Zwei Dinge, die dabei wichtig waren:

* Die Liste zeigte bisher **stur alle Drucker**, unabhängig vom Filter (`const listPrinters = P; // always show all`). Jetzt gilt derselbe Filter wie im Grid — das war die Voraussetzung dafür, dass „alle ausklappen" überhaupt eine sinnvolle Bedeutung hat.
* Die Liste wird alle paar Sekunden neu gezeichnet. Ohne Gedächtnis wäre jede geöffnete Zeile sofort wieder zugefallen; der Zustand liegt deshalb in `LIST_AUF`.

Am gerenderten Dokument geprüft: Standard 0 offen, Klick öffnet genau eine, Klick auf den Stern klappt nichts auf, Neuzeichnen lässt Offenes offen, mit Filter „online" zeigt die Liste 2 von 3 und „alle ausklappen" öffnet genau diese 2.

Bearbeiten sitzt jetzt als Knopf in der Detailzeile — der Doppelklick auf die Zeile ist entfallen, weil er sich mit dem Aufklappen ins Gehege kam.

### Am Prüfstand selbst

Zwei Lehren aus dem Flacker-Fehler:

* `liste.js` zählte Spalten per Textmuster im Quelltext. Das ging kaputt, sobald sich die Struktur änderte, und meldete 12 Zellen bei 7 Spalten. Die Prüfung sitzt jetzt in `liste2.js` **am echten Dokument** — Raster, Kopf und Zeile werden verglichen, alle drei stehen auf 7.
* Ein Prüfstand muss das Muster des Betriebs nachbilden, nicht ein bequemes. Verschachtelt statt nacheinander war der Unterschied zwischen „sieht gut aus" und „flackert sichtbar".

---

## 22. Version 1.6.0 — ein Bauweg für Kachel und Zeile

Deine Frage war richtig: Grid und Liste tragen dieselben Daten, also braucht es keine zweite Seite. Umgesetzt — mit einer Einschränkung, die bleibt.

### Was vorher schiefging

Für dieselben acht Größen gab es **vier** Stellen: `buildTile` las den Status direkt, `renderList` noch einmal, `updateStatusBar` ein drittes und `updateListRow` ein viertes Mal — jede mit eigener Fehlererkennung, eigener Prozentrechnung, eigener Zustandsliste.

Genau dort entstanden die Fehler dieser Woche: die LIVE-Marke, die auf ein längst entferntes Feld zeigte; der Filter, der in der Liste nie ankam (`// always show all printers`); die Zeile, die bei einem Fehler etwas anderes anzeigte als die Kachel.

### Was jetzt gilt

**Eine Datenquelle.** `druckerFelder(p)` ist die einzige Stelle, an der aus Rohdaten Anzeigewerte werden.

**Ein Feldgenerator.** `feldHTML(f, welches)` erzeugt ein Feld. Kachel und Zeile rufen dieselbe Funktion — deshalb *kann* dort nichts mehr auseinanderlaufen.

**Eine Aktualisierung.** `aktualisiere(ip)` zieht beide Ansichten gleichzeitig nach. `updateStatusBar` und `updateListRow` bleiben nur als Durchreiche bestehen, weil sie an vielen Stellen gerufen werden.

**Die Anordnung macht das Stylesheet.** `.df-item.zeile` ordnet die Felder in einer Rasterzeile, `.df-item.kachel` stapelt sie und macht Platz für das Bild. Eine neue Spalte braucht ab jetzt einen Eintrag in `DF_FELDER` und eine Breite im Raster — nicht mehr zwei Bauwege.

Kopfzeile und Datenzeile teilen sich **eine** Rasterdefinition. Vorher standen sie doppelt im Stylesheet und liefen prompt auseinander (9 Spalten im Kopf gegen 7 in der Zeile).

### Die Einschränkung

Das Bild bleibt der einzige echte Unterschied. 42 Videoelemente dauerhaft im Dokument zu halten wäre teuer, und Bilder sollen ohnehin nur in der Übersicht laufen. Es steht deshalb als Zusatz in der Kachel und nicht in den gemeinsamen Feldern. Nachgeprüft: in keiner Zeile steckt ein Bildelement.

### Der Beweis

Sechs Drucker, die jeden Zustand abdecken — laufend, pausiert, mit Fehler, fertig, offline, ohne Seriennummer. Für jeden wird Feld für Feld verglichen, was die Kachel zeigt und was die Zeile zeigt:

| Prüfung | Ergebnis |
|---|---|
| Unterschiede über 6 Drucker × 7 Felder | **0** |
| Zustandsklassen (Favorit, Fehler, pausiert) | in beiden gleich |
| nach `aktualisiere()` mit neuem Fortschritt | beide auf 77 %, identisch |
| Bild in einer Zeile | keins |
| Spalten in Raster / Kopf / Zeile | 8 / 8 / 8 |

### Aufgeräumt

30 tote CSS-Regeln entfernt, deren Klassen es im Markup nicht mehr gibt. Dabei sind mir zwei Regressionen unterlaufen — die Kopfzeile verlor ihr Raster und der Aufklapp-Pfeil verschwand — beide vom Prüfstand gefunden und behoben, bevor die Fassung das Haus verließ.

Und wieder eine doppelte Funktion gefunden: `updateListRow` gab es zweimal, die spätere gewann stillschweigend. Das ist das dritte Mal. Der Hinweis steht in `CHECKLISTE.md`.

---

## 23. Version 1.7.0 — sieben Punkte, davon zwei mit tieferer Ursache

### Warum Pause, Weiter und Abbrechen wirklich nicht wirkten

In 1.5.3 hatte ich das fehlende Pflichtfeld `param` ergänzt und die Sache für erledigt gehalten. War sie nicht. Diesmal habe ich den ganzen Weg gemessen — Knopf im Dokument, Klick, abgehende Anfrage:

| Prüfung | Ergebnis |
|---|---|
| Knöpfe in der Detailzeile | vorhanden, Sperrzustände korrekt |
| bei laufendem Druck | Pause frei, Weiter gesperrt |
| bei pausiertem Druck | Pause gesperrt, Weiter frei |
| ohne Seriennummer | gar keine Steuerung |
| gesendete Anfrage | `{"ips":["10.0.0.1"],"command":"pause"}` |
| nach Statusaktualisierung | Knöpfe da, Klick geht durch |

**Die Oberfläche war also in Ordnung.** Der Fehler saß dahinter: ein Befehl galt als erfolgreich, sobald der MQTT-Broker ihn angenommen hatte. Das heißt nur, dass die Nachricht raus ist — nicht, dass der Drucker sie ausführt. Die Oberfläche meldete „pausiert", während nichts geschah.

Das Protokoll sieht eine Antwort vor: derselbe Befehl mit derselben `sequence_id` zurück, dazu `result` und `reason`. Darauf wird jetzt bis zu vier Sekunden gewartet, und die Antwort entscheidet.

Besonders wichtig ist der Fall, in dem der Drucker **schweigt**. Solange ein Gerät mit der Herstellercloud verbunden ist, nimmt es über die lokale Verbindung Statusabfragen an, aber **keine Steuerbefehle** — und antwortet auf sie gar nicht. Vorher war das von einem Erfolg nicht zu unterscheiden. Jetzt steht dort:

> keine Rückmeldung vom Drucker — steht er auf LAN-Modus? Mit der Herstellercloud verbunden nimmt er nur Statusabfragen an, keine Steuerbefehle

Das ist meine begründete Erklärung für dein Problem, aber ich kann sie ohne deine Geräte nicht beweisen. Die neue Meldung sagt dir beim nächsten Versuch, ob sie stimmt.

### Druck starten mit Dateiwahl

Neuer Knopf **„🖨 Datei drucken …"** in der Detailzeile. Er liest die Dateiliste vom Gerät (getippte Namen gibt es bewusst nicht), lässt Datei und Filamentfach wählen und startet nach Rückfrage.

Zur Farbzuordnung, ehrlich: `ams_mapping` ordnet den Farben der Datei die Fächer zu und wird **von hinten** gefüllt — bei einer Farbe also `[-1,-1,-1,-1,fach]`. Wie viele Farben eine Datei braucht, steht aber in der Datei selbst, und die liegt auf dem Drucker. Deshalb gibt es hier die einfarbige Zuordnung; alles Weitere wäre geraten und der Drucker würde schlicht nicht anfangen.

Ein laufender Auftrag wird nie überschrieben — läuft etwas, lehnt der Server ab.

### Bilder nach dem Ansichtswechsel

Mein eigener Fehler aus 1.5.0: beim Wechsel in die Liste schaltet `stopAllMedia()` die Bildabfrage ab, beim Zurückwechseln hat sie niemand wieder eingeschaltet. `renderGrid()` lief, blieb aber wirkungslos. Nachgemessen: Grid `an` → Liste `aus` → Grid wieder `an`.

### „Alle einklappen"

Die Richtung stand fest im Klickbefehl, geändert wurde nur die Beschriftung. Nach einmal Ausklappen rief der Knopf ewig weiter „ausklappen". Er entscheidet jetzt beim Klick selbst. Gemessen: 3 offen → Beschriftung „▾ Alle einklappen" → 0 offen.

### Vier weitere doppelte Funktionen

Nach dem dritten Vorfall habe ich einen Wächter gebaut (`doppelt.js`). Er fand beim ersten Lauf sofort **vier** weitere: `editPrinter`, `savePrinter`, `fmtSize`, `setAllResolution` — dazu ein viertes `setView`.

Besonders ärgerlich: die gewinnende `editPrinter` bearbeitet Zeilen im Reiter *Drucker*, die verlierende die Listenzeilen. Mein Bearbeiten-Knopf in der Detailzeile war deshalb **wirkungslos**. Er führt jetzt in den Druckerreiter und öffnet dort den Editor. Alle toten Fassungen sind entfernt, der Wächter läuft ab jetzt bei jeder Auslieferung mit.

### Upload-Historie

Liegt in der Konfiguration in `%APPDATA%` und übersteht Neustarts. Festgehalten werden Zeit, Drucker, Datei, Größe, Dauer und Ergebnis — **auch die Fehlschläge**, gerade die sind später interessant. Höchstens 2000 Einträge, die ältesten fallen heraus. Dazu ein Knopf zum Leeren mit Rückfrage.

### Kachel neu angeordnet

Reihenfolge jetzt: Kopf → Pillen → Video → **Fortschrittsbalken** → Datei, Prozent, Restzeit. Am Dokument nachgeprüft.

### Startbild — warum der Gorilla nicht kam

Zwei Ursachen, beide gefunden:

1. **Die eingebettete Vorschau war da, aber praktisch schwarz.** 48 × 26 Pixel eines dunklen Gewitterbildes ergeben eine schwarze Fläche. Ich habe sie mir angesehen, statt mich auf „ist eingebettet" zu verlassen. Jetzt 240 Pixel breit, aufgehellt — der Gorilla ist erkennbar. Kostet 5,8 KB statt 452 Byte.
2. **Sie stand 33 KB tief im Dokument**, hinter dem gesamten Stylesheet. Bis dahin blieb das Fenster leer. Der Startbild-Stil steht jetzt als **erster** Block im Dokument, bei Zeichen 379 statt 33 430.

Das Favicon ist ebenfalls direkt eingebettet — eine Anfrage weniger, kein Aufblitzen des fremden Symbols.

Was bleibt: das leere Browserfenster, bevor das erste Byte da ist. Das gehört Edge.

### Über-Fenster

Ein Klick auf **Druckerfarm Switzerland** zeigt das Bild noch einmal, mittig, mit Logo, Copyright und Version. Das Copyright wird aus den Einstellungen übernommen statt im Code wiederholt — nachgeprüft, dass beide Texte übereinstimmen.

---

## 24. Version 1.8.0 — warum die H-Serie kein Bild lieferte

### Der Fund

Die Adresse in der go2rtc-Konfiguration beginnt mit `rtspx://`. Das ist eine Sonderform, die bei der X1-Reihe zuverlässig funktioniert — und genau das war der Grund, warum sie dort seit Monaten läuft.

**Die H2- und P2-Reihe antwortet darauf mit einer Umleitung auf `rtsps://` und liefert kein Bild.** Das deckt sich exakt mit dem Befund aus dem Log: alle 14 `woobly-h*` ohne Ausnahme, dazu vier Geräte aus derselben IP-Reihe.

Jetzt stehen **beide Adressen** unter demselben Streamnamen, in modellabhängiger Reihenfolge:

| Modell | zuerst | danach |
|---|---|---|
| H2D, H2C, H2S, P2S | `rtsps://` | `rtspx://` |
| X1C, X1E, P1S, Rest | `rtspx://` | `rtsps://` |

go2rtc probiert die Einträge der Reihe nach durch. Ein Gerät, das sich anders verhält als sein Modellname vermuten lässt, bekommt dadurch trotzdem ein Bild.

**Wichtig:** ich konnte das nicht an deinen Geräten prüfen. Die alte Konfiguration bleibt gültig, bis die neue Fassung sie überschreibt — nach dem Start also einmal in *Einstellungen → 📡 Drucker erreichbar?* nachsehen. Bleibt Port 322 zu, ist es weiterhin der Liveview-Schalter am Gerät und nicht das Schema.

### Das Symbol beim Start

Das eingebettete SVG-Favicon hat Edge offenbar nicht angenommen. Jetzt ist es ein **PNG mit dem Gorilla**, 64 × 64, direkt als Datenzeile in der Seite — keine zusätzliche Anfrage, kein fremdes Symbol.

### Bild in der aufgeklappten Zeile

Die Detailzeile hat jetzt zwei Spalten: **links das Bild, rechts die Druckerdaten.** Das Bild wird nur geholt, solange die Zeile offen ist; beim Zuklappen wird der Takt beendet und der Inhalt geleert. Nachgemessen: nach dem Zuklappen kommen keine weiteren Bildanfragen.

Das Bild bekommt eine eigene Kennung (`ldsnap-<ip>`), nicht dieselbe wie die Kachel im Grid. Sonst schriebe die Abfrage in das unsichtbare Element der ausgeblendeten Grid-Ansicht.

### Laufzeit, Firmware, AMS-Versionen

Unter der IP stehen jetzt Betriebszeit und Firmware, bei den AMS-Einheiten deren Version.

**Zur Betriebszeit muss ich ehrlich sein:** der Drucker gibt seinen eigenen Zähler über die örtliche Schnittstelle nicht heraus. Weder die Statusmeldung noch `get_version` enthalten ihn — ich habe beides geprüft. Was angezeigt wird, zählt dieses Programm selbst mit, seit der Drucker eingetragen wurde, und ist mit *„gemessen seit Eintrag"* gekennzeichnet. Es wäre unredlich, das als Gerätezähler auszugeben.

Die Firmware dagegen ist echt: `info.get_version` liefert die Baugruppenliste, daraus kommen die Druckerfirmware (Baugruppe `ota`) und die AMS-Stände. Abgefragt wird einmal nach dem Verbinden und danach etwa stündlich — die Liste ändert sich fast nie.

`ams/0` heißt in der Anzeige **AMS 1**: das Gerät zählt ab null, der Anwender ab eins.

### Kleinigkeiten

* `.header-right` ohne Innenabstand und Zwischenraum.
* In der Kachel stehen **Datei und Restzeit untereinander**, die Prozentzahl rechts daneben.
* Der Favoritenstern wächst nicht mehr beim Anschalten. Ursache war das Zeichen selbst: ⭐ ist ein Emoji und wird größer gesetzt als ☆. Jetzt ★ und ☆ aus derselben Zeichenfamilie, feste Breite, Farbe statt Größe.
* Die Steuerknöpfe sind nur noch so breit wie ihr Text (`flex:0 0 auto` statt `flex:1`).

### Am Prüfstand

Vier Prüfungen meldeten zunächst „falsch", obwohl die Regeln korrekt in der Datei standen. Ursache: seit das Startbild vorgezogen wurde, gibt es **zwei** Stilblöcke, und die Prüfstände lasen mit `split('</style>')[0]` nur den ersten. Sechs Prüfstände angepasst, sie lesen jetzt alle Blöcke.

Das ist innerhalb weniger Tage der dritte Fall, in dem ein Prüfstand stumpf geworden ist, ohne es zu melden. Deshalb prüfe ich bei einem `false` inzwischen zuerst, ob die Messung noch misst — nicht, ob das Programm falsch ist.

---

## 25. Version 1.9.0 — das Symbol, die Kamera-Adressen, vier Kästen

### Das Symbol beim Start — jetzt an der richtigen Stelle

Ich hatte zweimal am Favicon gearbeitet. Das war die falsche Ebene: **was Windows beim Start der Programmdatei zeigt, ist deren eigene Ressource** — kein Favicon, kein Browser, nichts, was die Seite beeinflussen könnte.

Die Programmdatei hatte gar kein Symbol. Jetzt trägt sie eins, in sechs Größen von 16 bis 256 Pixel, mit dem Gorilla.

Das war Handarbeit: der übliche Weg (`rsrc`) braucht einen Modul-Download, und der ist von hier aus gesperrt. Also erzeugt ein eigenes Skript die Ressource direkt — Symbolverzeichnis, COFF-Objekt, Relokationen.

**Und beinahe wäre es stillschweigend schiefgegangen.** Der erste Anlauf hatte einen Zusatzeintrag von 20 statt 18 Byte. Der Linker meldete beiläufig `fail to read string table` und **warf die Ressource weg — der Bau lief durch, „Done" stand da, das Symbol fehlte.** Aufgefallen ist es nur, weil ich hinterher in der fertigen Datei nachgesehen habe, ob die `.rsrc`-Sektion wirklich drin ist. Diese Prüfung bleibt im Ablauf.

### Kamera-Adressen: messen statt raten

In 1.8.0 hatte ich die Adressform aus dem Modellnamen abgeleitet. Das war schon besser als vorher, aber immer noch geraten — und du hast recht: es half nicht.

Jetzt gibt es im Video-Kasten den Knopf **„🔍 Kamera-Adresse testen"**. Er probiert am Gerät selbst durch:

| | |
|---|---|
| Schemata | `rtsps://`, `rtspx://` |
| Pfade | `/streaming/live/1`, `/streaming/live/0` |

Die erste Form, die ein Bild liefert, wird gemerkt und beim nächsten Schreiben der Konfiguration zuerst eingetragen — dauerhaft, je Drucker. Vorgeschaltet ist die Portprüfung: ist 322 zu, sagt der Test das direkt und probiert gar nicht erst vier Adressen durch.

Entscheidend dabei: **go2rtc antwortet auch mit 200, wenn es kein Bild bekommen hat.** Der Test prüft deshalb den Inhalt, nicht den Statuscode — genau der Fehler, der uns hier schon einmal wochenlang Erfolg vorgegaukelt hat. Vier Fälle sind abgesichert: echtes Bild, leerer Rumpf, Textmüll, Fehlerstatus.

### Ein Deadlock, gefunden vom Test

Die erste Fassung ließ `buildYaml` selbst auf die gemerkten Schemata zugreifen — während `writeGo2rtcYaml` die Sperre bereits hielt. Das Programm wäre beim Start des Videodienstes **hängen geblieben**. Der Test lief in die Zeitgrenze und zeigte den Stapel. Behoben, indem die Karte hineingereicht wird; ein eigener Test mit Frist hält das fest.

### Die vier Kästen

Die aufgeklappte Zeile ist jetzt in **Video · Druck · Drucker · AMS** geteilt, in einem Raster, das sich der Breite anpasst.

* **Druck**: Datei, Fortschritt mit Balken, Verbleibend, Schicht, darunter Pause/Weiter/Abbrechen
* **Drucker**: Name (dazu der Name im Gerät, falls abweichend), Modell, IP, Seriennummer, Dienstzeit, Firmware, Favoritenstern
* **AMS**: je Einheit Modell, Seriennummer, Firmware und die vier Fächer mit Filamenttyp und Farbstreifen — bis zu vier Einheiten

Dabei aufgefallen: meldet ein Gerät mehr AMS-Baugruppen als Fächerdaten (angeschlossen, aber noch keine Rolle gelesen), fiel die Einheit vorher aus der Anzeige. Jetzt zählt beides.

### Favoriten ohne Wartezeit

Der Klick wartete auf den Server — und der schreibt beim Speichern die **gesamte Konfiguration samt Upload-Historie** auf die Platte. Bei gefüllter Historie waren das spürbare Verzögerungen für einen Klick, der nur einen Stern färbt.

Jetzt wird sofort angezeigt und im Hintergrund gesichert; schlägt das Sichern fehl, dreht die Anzeige zurück und sagt es. Außerdem wird nicht mehr die ganze Kachel neu gebaut — das riss den laufenden Stream ab.

### Startfilter „Online"

Nach dem Startbild ist **Online** gewählt statt Gesamt. Bei 42 Geräten interessiert zuerst, was läuft.

---

## 26. Version 1.10.0 — die IP war es

Damit ist die Ursache geklärt, und sie war eine andere als alle meine Vermutungen: **die Drucker hatten neue Adressen bekommen.** Weder das Adressschema noch der Liveview-Schalter, sondern schlicht eine neue DHCP-Zuteilung.

Ich habe in zwei Versionen an Symptomen gearbeitet — erst am `rtspx`/`rtsps`-Schema, dann am Kamera-Test. Beides ist nicht falsch und bleibt drin. Aber die naheliegendste Erklärung habe ich nicht geprüft, obwohl das Werkzeug dafür seit 1.5.1 in der App steckt: die Erreichbarkeitsprüfung hätte für die alte Adresse „Gerät aus oder nicht im Netz" gemeldet, während unter der neuen Adresse ein Gerät antwortet.

### Die IP ließ sich gar nicht ändern

Der eigentliche Skandal. Im Editor standen Modell, Name, Zugangscode und Seriennummer als Eingabefelder — **die IP als fester Text.** Ein Drucker, der umzog, war damit unrettbar: löschen und neu anlegen, mit Verlust von Favorit und Laufzeit.

Jetzt ist sie ein Eingabefeld, mit Prüfung auf Format, gültigen Zahlenbereich und Dopplung. Beim Ändern werden Zustand und laufender Stream der alten Adresse sauber abgeräumt.

### Warum die Netzwerksuche nie etwas fand

Die Suche verschickte M-SEARCH und wartete auf Antworten **an ihrem eigenen, zufällig gewählten Port**. Das setzt voraus, dass die Geräte auf M-SEARCH antworten — und das tun sie offenbar nicht verlässlich.

Drucker melden sich aber von selbst: in regelmäßigen Abständen schicken sie eine NOTIFY-Nachricht an `239.255.255.250:2021`. Wer die hören will, muss **auf Port 2021 lauschen und der Multicast-Gruppe beitreten**. Genau das fehlte.

Jetzt läuft ein Dauerlauscher im Hintergrund:

* ein Empfänger auf `0.0.0.0:2021` für Broadcast und Unicast
* dazu je Netzkarte ein Multicast-Empfänger

Scheitert einer, laufen die anderen weiter — mit mehreren Netzkarten oder strenger Firewall ist das der Normalfall. Die Suche zeigt jetzt außerdem an, **welche Empfangswege offen sind und wie viele Pakete überhaupt ankommen.** Ohne diese Auskunft ist „findet nichts" nicht von „hört nichts" zu unterscheiden, und genau daran bin ich bisher gescheitert.

Nachgewiesen mit echtem UDP: ein NOTIFY-Paket wird empfangen und als Drucker erkannt, ein Router-Paket daneben nicht; beide werden gezählt.

### Umzüge werden von selbst erkannt

Taucht eine **bekannte Seriennummer unter neuer Adresse** auf, gilt das nicht mehr als neuer Drucker. Die Zeile wird gelb markiert, zeigt die bisherige Adresse und bietet **„↻ Adresse übernehmen"** — ein Klick, der Rest bleibt erhalten.

Weil der Lauscher dauerhaft mithört, fällt ein Umzug künftig auf, ohne dass jemand suchen muss.

### Was ich daraus mitnehme

Ich habe zweimal geliefert, ohne die einfachste Erklärung geprüft zu haben — und dabei jedes Mal etwas gebaut, das für sich genommen richtig war. Die Reihenfolge stimmte nicht: erst das Naheliegende ausschließen, dann das Interessante bauen.

---

## 27. Version 1.11.0 — die Suchansicht, wie du sie brauchst

### C13 war ein X1E

In der Netzwerkantwort steht nicht der Modellname, sondern eine interne Kennung. Ein X1E meldet sich als **C13** — und genau das stand in der Liste.

Es gibt jetzt eine Übersetzung. Belegt aus den Profildateien des Hersteller-Slicers:

| Kennung | Modell | | Kennung | Modell |
|---|---|---|---|---|
| C11 | P1P | | N1 | A1 mini |
| C12 | P1S | | N2S | A1 |
| **C13** | **X1E** | | O1D | H2D |
| BL-P001 | X1C | | BL-P002 | X1 |

**Für H2C, H2S, X2D und P2S habe ich keine belastbare Quelle gefunden.** Diese Kennungen bleiben deshalb unverändert stehen, statt dass ich sie rate — eine Kennung zum Nachschlagen ist besser als ein falscher Name, dem man glaubt. Das Modell lässt sich ohnehin von Hand berichtigen, und meldet die Firmware direkt einen Namen wie „H2D", bleibt der unangetastet.

### Eine Tabelle statt zweier Welten

Gefundene Geräte, die schon in der Liste stehen, erscheinen jetzt **im selben Aufbau** wie die neuen — mit Seriennummer, in einem eigenen Abschnitt *Bereits angelegt*. Neue Geräte stehen darüber unter *Noch nicht in der Liste*.

Erkannt wird über die **Seriennummer**, nicht über die Adresse. Die Adresse ist ja genau das, was sich ändert.

### Namen: beide, in einem Feld

Angezeigt wird der Name aus dem Programm und dahinter in Klammern der, den das Gerät selbst meldet — `meinName (3DP-H2D-042)`. Steht im Programm keiner, etwa bei einem noch nicht angelegten Gerät, gilt der vom Drucker.

### Ohne Rückfragen, ohne Neustart

* **„Adresse übernehmen" heißt jetzt „IP übernehmen".**
* Kein Bestätigungsfenster mehr. Der Vorgang ist umkehrbar, und man macht ihn im Zweifel für ein Dutzend Geräte hintereinander.
* **Die Suche startet danach nicht mehr neu.** Nur die betroffene Zeile wird neu gezeichnet; die Liste bleibt stehen und man arbeitet sie der Reihe nach ab. Neu suchen kann man weiterhin von Hand.

### Anlegen ohne Zugangscode

Beim Aufbau einer Farm trägt man erst die Geräte ein und die Codes später nach. **Das ging bisher nicht** — und das hätte die Oberfläche allein nicht gelöst: der Server wies solche Anfragen mit `400` ab. Aufgefallen ist das erst, weil ich das Programm laufen lassen und die Anfrage wirklich abgeschickt habe.

Jetzt sind nur noch Name und Adresse Bedingung. Ohne Code gibt es kein Bild und keinen Dateizugriff, aber der Eintrag existiert und lässt sich später ergänzen. Der Knopf sagt dabei an, wie viele der ausgewählten Geräte noch ohne Code sind.

### Bearbeiten an Ort und Stelle

Der Bearbeiten-Knopf in der Detailzeile sprang bisher in den Reiter *Drucker* — über eine Wartezeit von 60 ms und `scrollIntoView`. Man verlor die Stelle, an der man war, und beides kann stillschweigend schiefgehen; im Prüfstand tat es das auch.

Jetzt werden Name, Modell, **IP**, Seriennummer und Zugangscode direkt im Kasten *Drucker* bearbeitet, ohne Reiterwechsel.

### Das Programm wird jetzt wirklich gestartet

Neu im Ablauf: aus demselben Quelltext entsteht ein Linux-Bau, der hier gestartet und über seine Schnittstelle bedient wird. Damit ist zum ersten Mal geprüft, was das *laufende* Programm tut, statt nur was in seinen Bestandteilen steht.

Sofort gefunden: die abgewiesene Anfrage ohne Zugangscode. Diesen Lauf hätte ich mir viel früher einrichten sollen.

---

## 28. Version 1.12.0 — nichts wird mehr unter den Fingern weggezogen

### Der Blocker: die Detailzeile baute sich alle fünf Sekunden neu

Du hattest den Takt richtig erkannt. `aktualisiere()` zog die offene Detailzeile mit — und zwar, indem es sie **komplett neu baute**. Bei jeder Fehlerabfrage, also alle fünf Sekunden. Eine gerade geöffnete Dateiauswahl, ein halb ausgefülltes Formular, eine markierte Zeile: alles weg.

Jetzt gibt es drei Gründe, die einen Neuaufbau verhindern:

* ein Formular ist offen (Dateiauswahl, Bearbeiten)
* der Eingabezeiger steht in einem Feld der Detailzeile
* eine Auswahlliste ist aufgeklappt

Nachgemessen: Auswahl geöffnet, Wert gesetzt, zwei Statustakte ausgelöst — Auswahl noch da, Wert erhalten, Kasten offen.

### Zweimal auf „Online" klicken — reproduziert und behoben

Das war eine Reihenfolge im Start. `setPillFilter('online')` lief, **bevor** der Status da war. Zu dem Zeitpunkt gilt kein Drucker als online, der Filter lässt null übrig, das Grid bleibt leer. Der spätere `applyOnlineFilter()` blendet nur *vorhandene* Kacheln ein und aus — und es gab keine. Erst der zweite Klick zeichnete neu.

Mit verzögerten Antworten nachgestellt, so wie es im Betrieb zugeht:

| | Kacheln nach 300 ms | nach 2,6 s |
|---|---|---|
| 1.11.0 | **0 von 5** | **0 von 5** |
| jetzt | 5 von 5 | 5 von 5 |

Zwei Änderungen: der Start wartet den ersten Status ab, bevor gezeichnet wird, und ändert sich die sichtbare Menge, wird **neu gezeichnet** statt nur umgeschaltet.

Dazu blockiert nichts mehr: die Komponentenprüfung fragt GitHub und kann Sekunden dauern — bisher wartete das Anlaufen der Bilder darauf. Sie läuft jetzt nebenher.

### Die übrigen Punkte

* **Keine Liste „Bereits angelegt"** mehr. Gezeigt werden neue Geräte und, unter *Neue Adresse*, nur die umgezogenen. Ist beides leer, steht dort ein Satz statt einer Wand aus Bekanntem.
* **Der Zugangscode** steht in der Druckerliste als Punkte und wird erst beim Bearbeiten sichtbar.
* **Die Bearbeitungsfelder** stehen in derselben Spaltenfolge wie die Anzeige — Stern, Modell, Name, Code, Seriennummer, IP. Die Zeile ist jetzt ein Raster statt einer umbrechenden Reihe, deshalb rutscht die IP nicht mehr in eine eigene Zeile.
* **Der Favoritenstern** in der Druckerliste ist derselbe wie überall: gleiche Größe, Farbe statt Größenwechsel.
* **Die vier Kästen** sind gleich hoch (`align-items:stretch`, Inhalt füllt den Rest).
* **Die AMS-Fächer** zeigen denselben Farbpunkt wie im Grid — einfarbig, mehrfarbig als Verlauf, leer als Schraffur. Alle vier Fächer werden immer gezeigt; ein leeres bekommt keine Farbe, bleibt aber sichtbar, damit man sieht, welches frei ist.

Nachgemessen: 4 Fächer, 4 Punkte, davon 2 leer und 1 mehrfarbig.

---

## 29. Version 1.13.0 — Netzwerkverkehr anhalten

### Der Schalter

Unter *Einstellungen* → **„⏸ Kompletten Netzwerkverkehr pausieren"**. „Pausieren" ist dabei wörtlich gemeint, nicht nur für die Oberfläche:

* keine Statusabfragen, keine Fehlerabfragen, keine Bilder
* die MQTT-Verbindungen werden **getrennt**
* **go2rtc wird beendet**, damit keine Kameraverbindungen offen bleiben
* alle Hintergrundschleifen setzen aus — Blinklicht, Statusanforderung, Laufzeitzählung

Oben rechts neben dem Zahnrad erscheint ein pulsierendes 🚫, solange nichts läuft, und die Kacheln werden gedämpft. Ohne dieses Zeichen sucht man den Fehler sonst bei den Druckern.

Nachgemessen: nach dem Klick **0 Anfragen**, auch wenn man die Abfragen von Hand auslöst.

**Absichtlich nicht gespeichert.** Nach einem Neustart läuft das Programm wieder. Ein Schalter, der über Nacht stehen bleibt und am nächsten Morgen alles tot aussehen lässt, wäre eine Falle.

### Zwei Fehler, die der Race-Detektor gefunden hat

**Der Aufpasser hätte die Pause ausgehebelt.** `superviseGo2rtc` deutet jedes Ende von go2rtc als Absturz und startet neu — auch das gewollte Beenden beim Pausieren. Die Pause wäre nach Sekunden von selbst vorbei gewesen. Jetzt bleibt go2rtc unten, solange angehalten ist.

**Zehn Drucker anlegen hieß zehn go2rtc-Neustarts.** `restartGo2rtcAsync` startete für jeden Aufruf eine eigene Stopp-Start-Folge. Bei mehreren Druckern hintereinander überlappten die sich, und go2rtc kam kaum zum Laufen. Jetzt setzt jeder Aufruf nur die Frist neu; ausgeführt wird einmal, wenn Ruhe eingekehrt ist. Gemessen: 10 Aufrufe → **1 Start**.

Dazu wird jetzt sofort geprüft, ob go2rtc überhaupt installiert ist — vorher blieb sonst eine Aufgabe liegen, die Sekunden später ins Leere lief.

Beides waren echte Fehler, keine Testartefakte. Aufgefallen sind sie nur, weil der Lauf mit `-race` sie in einen anderen Test hineinlaufen ließ.

### Die zwei kleinen Punkte

* `.ueber-logo` hat jetzt `filter: invert(1) grayscale(2)`.
* **Das leere AMS-Fach ist keine Säule mehr.** Im gestreckten Raster zog die Schraffur des leeren Punktes sich in die Höhe. Der Punkt hat jetzt feste 11 × 11 Pixel, und das Fach wächst nicht mehr mit der Kastenhöhe.

---

## 30. Version 1.13.1 — Zwischenstand

Kleine, in sich abgeschlossene Fassung. Die groesseren Wuensche (Pause-Knopf als eigene stcard, Warnmeldungs-Historie, Upload-Historie erst auf Klick) kommen als naechstes.

### Das leere AMS-Fach — jetzt an der Ursache behoben

In 1.13.0 hatte ich das Symptom kuriert: die Kaestchen auf `align-self:start` gesetzt, damit die Saeule aufhoert. Das machte die leeren Kaestchen kleiner als die gefuellten. Die eigentliche Ursache war eine andere: `.ams-dot` ist ein `<span>` ohne `display`, also `inline` — und auf inline-Elementen wirken `width`/`height` nicht, die Schraffur zog sich in die Hoehe. Der Punkt hat jetzt `display:block` mit fester Groesse, und die Kaestchen sind wieder alle gleich gross (`align-self:stretch`).

### Steuerung: die Meldung nennt jetzt den Developer Mode

Die Firmware ab 01.08.05 lehnt Fremdsteuerung mit „mqtt message verify failed" ab, solange am Geraet kein Developer Mode aktiv ist. Statt dieser kryptischen Meldung steht jetzt:

> Drucker lehnt die Steuerung ab — am Gerät den Developer Mode aktivieren (Einstellungen → WLAN → LAN Mode Only → Developer Mode).

Das ist keine Umgehung der Sperre — die baue ich bewusst nicht ein, sie setzt einen extrahierten Herstellerschluessel voraus. Es ist der Hinweis auf den vom Hersteller vorgesehenen Weg, den du gehst.

---

## 31. Version 0.9.2 — Versionsnummer zurückgesetzt, Dateien je Drucker markieren

### Version steht jetzt unter 1.0

Auf deinen Wunsch (dreifach bestätigt) steht die Versionsnummer ab jetzt auf **0.9.x** und geht **nicht** von selbst auf 1.0 oder höher — nur auf ausdrücklichen Befehl. Die Regel steht ganz oben in `CHECKLISTE.md`, damit sie auch einen Neustart übersteht. Innerhalb von 0.9.x wird weitergezählt.

### Druckdateien je Drucker markieren und löschen

Beim direkten Laden der Dateiliste eines Druckers (der Weg **ohne** die Suche) gibt es jetzt:

* eine Checkbox je Datei
* ein „alle"-Kästchen in der Kopfzeile
* einen Knopf **„Markierte löschen"**, der gesperrt bleibt, solange nichts angehakt ist

Gelöscht wird über denselben Endpunkt wie das bisherige Einzellöschen, der Reihe nach für jede markierte Datei; die Zeilen verschwinden einzeln, und am Ende steht, wie viele gelöscht wurden und ob etwas fehlschlug.

Am gerenderten Dokument geprüft: 3 Dateien → 3 Checkboxen, zwei markiert → Knopf frei und „2 markiert", nach dem Löschen genau diese zwei weg, die dritte bleibt.

---

## 32. Version 0.9.3 — leeres AMS-Fach und Pause-Karte

### Das leere Fach war strukturell zu hoch

Zweimal habe ich am selben Symptom gedreht, ohne die Struktur zu ändern — und es blieb: das leere Fach zog seine Schraffur in die Höhe wie eine Säule. Die Ursache lag im Aufbau selbst. Die Fächer waren aufrechte Kästen (`flex-direction:column`) in einem gestreckten Raster; da genügte eine Kleinigkeit, damit sich der leere Punkt vertikal aufzog.

Jetzt sind die Fächer **flache Chips** mit **fester Höhe von 20 px** — Farbpunkt links, Typ rechts, dieselbe Form wie im Grid. Ein Fach *kann* sich nicht mehr in die Höhe ziehen, egal was drinsteht, weil die Höhe fest steht und nicht mehr vom Inhalt kommt. Das ist der Unterschied zu vorher: nicht den Punkt bändigen, sondern dem Kasten gar keinen Spielraum lassen.

### Pause-Knopf als eigene Karte

„⏸ Kompletten Netzwerkverkehr pausieren" steht jetzt in einer eigenen stcard mit Überschrift *🌐 Netzwerkverkehr* und einem Satz, was der Knopf tut — nicht mehr zwischen den Komponenten-Knöpfen.

---

## 33. Version 0.9.4 — Steuerung im Dropdown

Alle Steuerbefehle laufen über dieselbe MQTT-Quittung; **ohne Developer Mode am Gerät weist die Firmware sie ab**, und die Meldung sagt das dann. Mit aktiviertem Developer Mode funktioniert die ganze Gruppe.

### Das Dropdown wird beim Bedienen nicht mehr zurückgesetzt

Der eigentliche Fehler, den du gemeldet hast: das Grid-Dropdown wurde im Takt der Statusmeldungen komplett neu gebaut — alle paar Sekunden — und riss dabei ein offenes Druckstart-Formular weg, bevor man es ausfüllen konnte. Jetzt gilt dieselbe Sperre wie in der Detailzeile der Liste: `ddFrei()` verhindert den Neuaufbau, solange ein Formular offen ist oder der Zeiger in einem Feld steht. Im DOM geprüft: bei offenem Formular gesperrt.

Damit ist auch dein Einwand gelöst — man muss den Netzwerkverkehr **nicht** abschalten, um in Ruhe einzustellen; das Menü bleibt einfach stehen, während man darin arbeitet.

### Düse und Bett einzeln steuern

Im Dropdown je eine Zeile für Düse und Bett: **− / +** in 5-Grad-Schritten, ein Eingabefeld, **Setzen** und ein **↺**, das auf den vom Drucker gemeldeten Sollwert zurückgeht. Gesendet wird `M104` / `M140`, ohne zu warten. Grenzen im Server: Düse 0–300 °C, Bett 0–120 °C.

### Filament umordnen — Typ, Farbe, Temperatur, Hersteller

Über `ams_filament_setting`: Fach wählen, ein **Profil** auswählen (das füllt Typ und Temperatur vor), Farbe über einen Farbwähler, Temperaturbereich, dann setzen.

Zum Hersteller ehrlich: den nimmt der Drucker **nicht als freien Text** an — er steckt in der Profil-Kennung (`tray_info_idx`, z. B. `GFA00`). Deshalb eine Auswahl gängiger Profile, die Marke und Typ zugleich festlegen. Die Anzeigetexte nennen dabei keinen Herstellernamen (Rechtssicherheit); die Kennungen darunter sind reine Codes.

### Beide AMS im Druckstart, „Kalibrieren"

Die Fachwahl beim Druckstart zeigt jetzt **alle AMS-Einheiten mit allen vier Fächern**, auch leere — bei zwei AMS also 8 Fächer plus externe Rolle. „Bett einmessen" heißt jetzt „Kalibrieren".

### Beim Bauen aufgefallen

Meine ersten Profilnamen hießen „Bambu PLA Basic" usw. — der Herstellername stand damit 5× in der Programmdatei. Die automatische Prüfung nach dem Bauen hat das gefunden, bevor es rausging; die Texte sind jetzt neutral, die Exe wieder sauber (0 Treffer).

---

## 34. Version 0.9.5 — Detailkästen aufgeräumt, Startbild mit Copyright

### Vier Kästen, jeder mit eigener Aufgabe

Die aufgeklappte Detailzeile hatte alles in den Kasten „Druck" gestopft — Pause/Weiter/Stop, Temperatur und Filament zusammen. Jetzt ist es sauber verteilt, und jeder Kasten trägt seine eigene Klasse (`dk-druck`, `dk-drucker`, `dk-ams`, `dk-video`):

* **Druck** — nur der laufende Druck: Datei, Fortschritt, Verbleibend, Schicht, Pause/Weiter/Abbrechen
* **Drucker** — die Gerätedaten *und* die Düsen-/Betttemperatur (die gehört zum Gerät, nicht zum Auftrag)
* **AMS** — die Fächer *und* die AMS-Anordnung (das Filament-Umordnen, umbenannt)

### Das leere Fach — jetzt endgültig

Ich hatte dreimal an der eigenen `dk-slot`-Kachel gedreht. Diesmal ist sie ganz weg: die Fächer nutzen dieselben `.ams-tray`-Chips wie das Grid, die nie zu hoch waren, mit fester Höhe von 20 px. Ein leeres Fach *kann* nicht mehr als Säule erscheinen, weil es dieselbe erprobte Form hat wie ein belegtes. Im DOM geprüft: 8 Fächer bei zwei AMS, 7 als leer markiert, keine `dk-slot`-Klasse mehr.

### Grid-Dropdown zweispaltig

Das Menü wurde mit Temperatur und Filament sehr lang. Jetzt zwei Spalten: links die Angaben und das AMS, rechts die Steuerung samt Temperatur und Filament. Das Menü ist dafür breiter (bis 560 px) und fällt auf schmalen Fenstern wieder auf eine Spalte zurück.

### Startbild zeigt gleich das Copyright

Beim Start erscheinen jetzt sofort Logo, „Switzerland" und die Copyright-Zeile — mittig, über dem Gorilla-Bild, zusammen mit dem Ladehinweis. Damit ist die Info vom Klick aufs Logo schon beim Start da; eine zweite Seite dafür braucht es nicht mehr.

---

## 35. Version 0.9.6 — zehn Änderungen

### Druckauftrag scheiterte mit 0x07FF8012 — behoben

Der Code ist ein Dateisystemfehler. Ursache: eine rohe `.gcode`-Datei wurde mit dem Befehl `project_file` und dem Pfad `Metadata/plate_1.gcode` gestartet — den gibt es aber nur *innerhalb* eines `.3mf`-Archivs. Jetzt wird je nach Endung getrennt: `.gcode` geht über `gcode_file` mit dem Dateinamen, `.3mf` weiter über `project_file` (URL jetzt `file:///sdcard/…`). An echten Geräten konnte ich es nicht prüfen; die Aufteilung folgt dem Protokoll.

### Beim Start starten jetzt alle Streams

Manche Streams liefen erst nach einem zweiten Klick auf „Streams starten". Grund: die Funktion, die die Bildabfragen staffelt, hat bei jedem Aufruf zuerst **alle** Poller gestoppt und neu verteilt. Wird das Grid beim Start mehrfach neu gezeichnet (jede eintreffende Statusmeldung), fing die Staffelung jedes Mal von vorn an und die Kacheln am Ende der Liste kamen nie dran. Jetzt arbeitet sie additiv — nur was noch nicht läuft, wird gestartet. Gemessen: 10 Kacheln, dreifaches Neuzeichnen → 10 Poller; Moduswechsel 2→5 s startet sie sauber neu.

### AMS-Anordnung direkt bei jedem Fach

Kein eigener „Filament umordnen"-Block und keine Fach-Auswahl mehr. Jedes Fach in der AMS-Übersicht hat jetzt inline **Hersteller · Material · Farbe** und einen Setzen-Haken. Der Drucker verlangt Hersteller und Material gekoppelt in einer Profil-Kennung — die wird aus der Auswahl nachgeschlagen (z. B. Generic + PETG → `GFG99`). Wo keine Kennung existiert, genügen Typ und Farbe.

### Die übrigen Punkte

* **Betriebsstunden** statt Einschaltzeit: die Laufzeit zählt jetzt, solange das Gerät eingeschaltet und erreichbar ist — nicht mehr nur die reine Druckzeit. (Weiterhin von diesem Programm gezählt; das Gerät gibt keinen eigenen Gesamtzähler heraus.)
* **Mini-Ladekreis** über dem Snapshot-Bild, winzig oben links, dessen Umlaufzeit sich an den Takt anpasst — bei 2, 5 oder 10 Sekunden entsprechend.
* **Filter ein-/ausblendbar**: unter *Einstellungen → Filter anzeigen* lässt sich Online, Offline, Pausiert und Favoriten je einzeln zeigen oder verbergen. Vorgabe: alle sichtbar.
* **„Gesamt" heißt jetzt „Alle"** (englisch „All").
* **Temperatur in 1-Grad-Schritten** statt 5.
* **Das leere AMS-Fach** war zu hoch, weil ein ungenutzter `.empty`-Stil mit 80 px Padding versehentlich auf die AMS-Punkte durchschlug. Der Stil ist entfernt.
* **Copyright**: „Reinhard von der Waydbrink" → „Reinhard @".

---

## 36. Version 0.9.7 — Druckstart, Dateiliste, Copyright

### Druckstart: der Pfad war falsch (0x07FF8012 → dann 0x05002402)

Zweimal am selben Punkt danebengelegen. Das Protokoll ist eindeutig, ich hatte es nur nicht genau genug umgesetzt:

* **`.gcode`** über `gcode_file` erwartet den **bloßen Dateinamen ohne führenden Schrägstrich**. Mein `/dateiname.gcode` war der „unsupported path".
* **`.3mf`** über `project_file` braucht die URL-Form **`ftp:///dateiname.3mf`** (SD-Wurzel), nicht `file:///sdcard/…`.

Beides ist jetzt so umgesetzt, wie es die Protokolldokumentation angibt. An echten Geräten kann ich es nach wie vor nicht prüfen — sag mir, ob der nächste Versuch durchgeht.

### Filamentzuordnung beim Drucken

Statt eines einzelnen Fachs gibt es jetzt bis zu **vier Farbzuordnungen**: „Farbe 1 … 4 → Fach". Einfarbig genügt Farbe 1; mehrfarbig ordnet man den Farben der Datei der Reihe nach Fächer zu. Daraus wird die `ams_mapping`-Liste gebaut, die der Drucker erwartet. Vor dem Start zeigt die Rückfrage die Zuordnung Zeile für Zeile.

### Dateiliste im File-Sync war leer

Beim „Laden" standen nur noch Checkboxen, die Namen fehlten. Ursache: die Zeile war ein Flexbox-Layout ohne `min-width:0` — ein langer Dateiname (bei den Geräten die Regel) sprengte die Zeile, statt zu kürzen. Die Zeile ist jetzt ein festes Raster (Checkbox · Name · Größe · ⬇ · ✕), der Name bekommt eine eigene Spalte, die immer kürzt statt zu überlaufen. Mit einem absichtlich überlangen Namen geprüft: Name sichtbar und sauber beschnitten.

### Copyright

„© Reinhard @ Brickshouse GmbH · Druckerfarm.ch" — mittiger Punkt vor Druckerfarm.ch, kein Punkt nach dem @.

---

## Verifikation

* `build.ps1`-Flags (`CGO_ENABLED=0 GOOS=windows GOARCH=amd64 -mod=vendor -ldflags="-s -w -H windowsgui"`) — Build ok, 7,3 MB
* Exe auf den Herstellernamen durchsucht: **0 Treffer**; die Protokollwerte (`bblp`) sind erwartungsgemäß noch drin
* `go vet` sauber, `gofmt` sauber für alle neuen Dateien (die vorhandene Abweichung in `main.go` Zeile 257 ist unverändert geblieben)
* JS-Syntax der eingebetteten UI mit `node --check` geprüft
* **173 Go-Tests, auch mit `-race`** (`ok druckerfarm 20,1 s`, keine Race-Meldung), `go vet` sauber:
  * Snapshots (7): Streamname, leerer Body → 503, unbekannter Drucker, Cache, Concurrency-Limit ≤ 4 bei 36 parallelen Anfragen, Request-Coalescing, `max_age`-Untergrenze
  * Komponenten (13): Entpacken nur der gewünschten Einträge, Zip-Slip-Schutz, erfolgreicher Download, verfälschte Prüfsumme → Abbruch ohne Installation, HTTP-Fehler, ffmpeg-Prüfsummenliste, fehlender Listeneintrag, Fortschrittszählung, Statusendpunkt, Persistenz von „nicht mehr fragen", doppelter Download-Start, PATH-Voranstellung
  * Start/Ende (4): kein Abschalten während das UI pollt, Abschalten nach Stille, Aufgeben wenn nie jemand verbindet, Request-Erfassung, belegter Port
  * Blinken (4): alte Störung wird beim Fortsetzen verworfen, Störung während des Drucks bleibt, Liste übersteht einen Neustart, Abschalten leert die Liste
  * Pausierte Streams, Logo, Komponenten-Update (7): Wachstum und Deckelung der Pause, **pausierter Stream wird nicht mehr gefragt**, **gesunder Drucker wartet nicht hinter kaputten**, Pause wird nach Erfolg aufgehoben, eingebettetes Logo wird ausgeliefert, Versionszeile von go2rtc, gemerkte Prüfsumme
  * Bündeln und Blinklicht (5): Gruppierung der Dateien je Drucker, unzulässige Namen ohne FTP-Aufruf abgewiesen, Fehlererkennung über alle Quellen, Modellliste mit Vorgaben, Lichtbefehl ohne Verbindung meldet klar
  * Such- und Löschvorgänge (8): nur neue Einträge im Zwischenstand, unbekannter Vorgang, **Abbruch stoppt laufende Arbeit wirklich**, Mindestlänge, leerer Filter, unbekannte Drucker abgewiesen, **Pfadtrenner im Dateinamen abgewiesen**, Aufräumen beendeter Vorgänge
  * Update und Einstellungen (13): Versionsvergleich inkl. 1.9 gegen 1.12, Auswahl der Programmdatei aus den Anhängen, fehlendes Repository als klare Meldung, Format-Prüfung, **HTML-Seite wird als Programmdatei abgelehnt**, Einstellungen speichern und ungültige Werte abweisen, Datenverzeichnis, einmalige Übernahme der Einrichtung, **Selbstersetzung mit Start der neuen Fassung**, **Rückfall bei Fehlschlag**, Aufräumen der Vorgängerversion
  * Netzwerksuche (9): Kopfzeilen mit und ohne Suffix, fremde Geräte gefiltert, NOTIFY akzeptiert, korrekt geformtes M-SEARCH auf Port 2021, Broadcast-Berechnung, **über echtes UDP: Fund trotz zweier verlorener Pakete**, Deduplizierung, Suchziele, Endpunktform
  * Status und Steuerung (10): Nullwerte werden gespeichert, Fortschritt fällt auf 0 zurück, fehlende Schlüssel überschreiben nichts, AMS-Parsing, AMS überlebt Nachrichten ohne AMS-Teil, AMS im JSON, **AMS gegen echten Drucker-Payload**, unbekanntes Kommando abgelehnt, leere Druckerliste abgelehnt, unerreichbarer Drucker sauber gemeldet
* **Neu in 1.5.0** — zwei weitere Prüfstände ohne Browser:
  * `sync2.js`: Auswahl greift nur auf angehakte Drucker, Upload übersteht einen Neuaufbau der Anzeige (4/4 Dateien), Abbrechen stoppt bei 2/4, Pause hält den Zähler 60 ms lang bei 2 und läuft danach auf 4
  * `liste.js`: Rasterspalten 7 = Kopfzellen 7 = Zellen je Zeile 7, Aktions- und Statusspalte weg, Pausiert-Filter trifft genau Pause, Abbruch und Druckfehler und lässt Offline sowie laufende Drucker aus
* Blinken als Ablauf durchgespielt (`TestBlinkenEndetNachFortsetzen`): Störung in der Pause blinkt, nach dem Fortsetzen nicht mehr, die per pushall wiederholte alte Meldung löst nichts aus, eine neue Meldung während des Drucks schon
* Übersetzung der FTP-Fehler gegen echte Python-Ausgaben geprüft (5 Fälle plus Kürzung)
* **Neu in 1.5.1** — die Seite wird jetzt in einem echten DOM geladen (`jsdom`) statt nur auf Syntax geprüft. Das war die Lücke, durch die zuletzt eine ganze Auslieferung unbemerkt an dir vorbeigehen konnte: `node --check` sieht nur Klammern, nicht Wirkung.
  * `splash.js`: Startbild vorhanden, bei 1601 ms sichtbar, ab 2002 ms ausgeblendet, nach 4 s aus dem Dokument, 0 Laufzeitfehler
  * `reachui.js`: Erreichbarkeitsliste mit der Lage aus deinem Log nachgestellt — 3 Zeilen, Einfärbung grün/gelb/rot richtig zugeordnet, Portmarken passend
  * `schoen.js` (neu in 1.5.2): 25 Abfragen → 2 Klassenwechsel am Modusknopf, Rahmen läuft durch, Moduswechsel lässt keinen Rahmen zurück; Exporttitel, Spaltenreihenfolge mit Seriennummer vor IP, Spaltenmaße der Erreichbarkeitsliste und das entfärbte Zahnrad jeweils am gerenderten Dokument geprüft
* Erreichbarkeitsprüfung gegen **echte Sockets**: offener Port erkannt, geschlossener als „abgelehnt", nicht routende Adresse als „keine Antwort"; die drei Ports werden nachweislich gleichzeitig geprüft (sonst dauerte ein Durchlauf über 42 Drucker über zwei Minuten)
* **Oberfläche mit einem Node-Prüfstand getestet** (kein Browser nötig): nach „Alle starten" 2 offene Verbindungen mit Bild, nach „Alle stoppen" **0 offen**, nach Wechsel in den Snap-Modus **0 offen**, und beim Aufbau der Vollbildansicht **0 neue Verbindungen** bei weiterhin laufendem Bild
* Startverhalten mit einem sofort endenden Fake-Browser nachgestellt — vorher 0/5, nachher 5/5 erreichbar
* Farbkontraste nach WCAG nachgerechnet, schwächster Fließtext-Wert 8,9:1 gegenüber 2,8:1 vorher
* Python-Hilfsskript gegen eine FTP-Attrappe geprüft: gebündeltes Löschen mit Teilfehler, rekursive Ordnersuche, Tiefenbegrenzung
* End-to-End Snapshots gegen einen echten go2rtc-Prozess mit H264-Quelle: Live-Stream → 44 KB JPEG, toter Drucker → 503 mit Klartext-Grund
* End-to-End Download-Pipeline mit einer echten go2rtc-Binary im Archiv: Download → Prüfsumme → Entpacken → `-version` → Prozessstart, alles durchgelaufen
* WebRTC-Leak-Harness nach allen Umbauten erneut: 0 offene Verbindungen, 0 verwaiste Timer
* ZIP entpackt und daraus frisch gebaut: `PE32+ executable (GUI) x86-64`

**Was ich nicht testen konnte.** Mein Testrechner kommt nicht an GitHub, der echte Download ist also nur gegen einen lokalen Server geprüft, nicht gegen die echten URLs. Die URLs und Hashes habe ich am 28.07.2026 direkt von den GitHub-Release-Seiten übernommen. Falls der go2rtc-Download an der Prüfsumme scheitert, ist ein neueres Release erschienen — dann `go2rtcVersion` und `go2rtcSHA256` in `components.go` aktualisieren.

Ebenfalls nicht testbar ohne Hardware: das Verhalten an echten Druckern (RTSP-Verbindungsaufbau, tatsächliche Snapshot-Latenz). Die Staffelungsintervalle (80 ms Stream-Start, Snapshot-Verteilung über das Poll-Intervall) sind begründete Schätzungen und ggf. an der Farm nachzujustieren.

## 37. Version 0.9.8 — Helles Design ist jetzt Standard

Bei einer frischen Installation kam das dunkle Design. Es soll immer hell starten. Drei Stellen legten den Standard fest:

**Server.** In `main.go` wurde ein fehlender gespeicherter Wert auf `"dark"` gesetzt (`if theme == "" { theme = "dark" }`). Jetzt `"light"`. Das ist der eigentliche Auslöser: bei einer Neuinstallation gibt es noch keine gespeicherte Einstellung, also griff dieser Vorgabewert.

**Oberfläche.** In `ui.go` machte `setTheme` alles außer exakt `"light"` zu dunkel (`(name === 'light') ? 'light' : 'dark'`). Umgedreht: nur ein ausdrückliches `"dark"` ergibt dunkel, alles andere hell. So kann ein fehlender Wert nie mehr versehentlich dunkel werden.

**Erstes Bild.** Das `<html>`-Element bekommt gleich `data-theme="light"`, damit vor dem Start des JavaScripts kein dunkler Aufblitzer entsteht.

Wer schon einmal auf Dunkel umgestellt hat, behält seine Wahl — gespeicherte Werte werden weiter geladen. Es geht nur um den Auslieferungszustand.

Geprüft: jsdom-Prüfstand (Wurzel startet hell, `setTheme` gibt bei leerem/fehlendem Wert hell, `"dark"` bleibt dunkel, Server-Vorgabe hell), `go test` und `-race` sauber, Herstellername 0×, `data-theme="light"` in der Exe.

## 38. Version 0.9.9 — Drucker-Firmware-Updates anzeigen

Neue Funktion: das Programm zeigt, welche Drucker eine neue Firmware offen haben. Es löst **nichts** aus — es liest nur mit, was der Drucker selbst weiß.

**Warum der Drucker das im „Nur LAN"-Zustand überhaupt weiß.** „Nur LAN" trennt die Cloud-Bindung — Konto, Fernsteuerung, Handy-App —, nicht das Netzwerk. Der Drucker prüft weiterhin über einen eigenen Kanal beim Hersteller, ob neue Firmware vorliegt, und schreibt das Ergebnis in seinen Statusbericht (`upgrade_state` / `new_ver_list`). Voraussetzung ist nur, dass dein Netz einen Weg ins Internet hat.

**Was neu ist.**

Eine stcard „🔄 Drucker-Software" in der Übersicht mit dem Knopf „Drucker Software Update suchen". Der Knopf schickt an jeden verbundenen Drucker einen `pushall` und liest die frischen Update-Meldungen aus. Er startet kein Update.

Das Ergebnis erscheint als Marke „⬆ Update" hinter dem Druckernamen in „Gespeicherte Drucker" — in Grid und Liste. Der Mauszeiger über der Marke zeigt je Baugruppe die laufende und die neue Nummer (z. B. „Firmware: 01.07.00.00 → 01.08.02.00", „AMS 1: …").

Die Funde werden lokal gespeichert (`fw_updates` im Zustand) und überstehen einen Neustart. Sie bleiben stehen, bis der Drucker die Fassung installiert hat: bei jedem Statusabruf gleicht das Programm die gemerkten Updates gegen die laufende Firmware ab (`info.get_version`) — ist die neue Nummer erreicht oder meldet der Drucker die Baugruppe nicht mehr, verschwindet die Marke von selbst.

**Technisch.** `mqtt.go` liest `upgrade_state` aus dem pushall mit. `fwupdate.go` hält Datenmodell, Parser (`parseUpgradeState`), Abgleich (`nochOffen`) und Scan (`scanMergeFw`) samt Endpunkt `/api/fwupdate/scan`. `/api/status` liefert je Drucker die offenen Updates gleich mit. Die Marke steckt in derselben `.df-name`-Hülle wie der Name, damit der Statustakt sie nicht vervielfacht.

Geprüft: sechs Go-Tests (Fund, Gleichstand, leer, Wegräumen nach Installation, offline unangetastet, ohne Info erhalten), jsdom-Prüfstand (Karte + Knopf da, Marke mit richtigem Tooltip, **eine** Marke auch nach drei Aktualisierungstakten, weg nach Installation, Knopf löst nur die Suche aus). `go test` und `-race` sauber, Herstellername 0×.

**Ehrlich zur Einschränkung.** Ob ein Drucker im „Nur LAN"-Zustand die Firmware auch wirklich prüft, hängt vom Firmware-Stand und davon ab, ob das Netz Internet hat. Findet der Drucker selbst nichts, kann auch dieses Programm nichts anzeigen — es erfindet keine Updates.

## 39. Version 0.9.10 — Update-Suche in den Drucker-Tab, X1E erkannt

Zwei Korrekturen zur Update-Anzeige aus 0.9.9.

**Karte und Marke jetzt im Tab „+ Drucker".** Die stcard „Drucker-Software-Updates" saß in der Monitoring-Übersicht. Sie gehört zur Druckerverwaltung und steht jetzt im Tab „+ Drucker", direkt über „Gespeicherte Drucker". Die Marke „⬆ Update" erscheint nun auch in der Liste „Gespeicherte Drucker" hinter dem Namen (zusätzlich weiter in Grid und Liste der Übersicht).

**X1E wurde nicht erkannt.** Der Grund lag am Meldeformat. Neuere Drucker (P1/A1/H2) schicken eine `new_ver_list` mit laufender und neuer Nummer je Baugruppe. Die X1-Reihe — auch der X1E — kennt diese Liste nicht, sondern legt je Baugruppe ein Einzelfeld an: `ota_new_version_number`, `ams_new_version_number`, `ahb_new_version_number`, dazu `new_version_state` (1 = aktuell, 2 = etwas offen). Der Parser las nur die Liste und übersah deshalb den X1E — bei dem, wie vermutet, offenbar nur das AMS ein Update offen hat.

Jetzt liest der Parser beide Formate. Beim Einzelfeld-Schema steht die laufende Fassung nicht dabei; deshalb fragt die Suche zusätzlich `get_version` ab und vergleicht die neue Nummer gegen die tatsächlich laufende Firmware — so verschwindet die Marke, sobald installiert. Für das AMS ohne Index zählt der niedrigste AMS-Stand: liegt irgendeine AMS-Einheit darunter, bleibt das Update offen. Platzhalter wie `00.00.00.00` gelten nicht als Update, und `new_version_state == 1` unterdrückt die Anzeige ausdrücklich.

**Diagnose eingebaut.** Findet die Suche nichts, blendet die Karte die rohe `upgrade_state`-Meldung jedes Druckers ein (aufklappbar). Damit lässt sich ablesen, was ein Gerät wirklich schickt — falls doch ein weiteres Format auftaucht, kann ich es danach ergänzen.

Geprüft: elf Go-Tests (u. a. X1-Einzelfelder AMS-only, OTA+AMS zusammen, `state==1` unterdrückt, Nullversion, niedrigster AMS-Stand), zwei jsdom-Prüfstände (Karte im Drucker-Tab und nicht mehr in der Übersicht, Marke in „Gespeicherte Drucker" ohne Duplikate, X1E-AMS-Tooltip, Diagnose sichtbar) und der bestehende Übersichts-Prüfstand unverändert grün. `go test` und `-race` sauber, Herstellername 0×.

**Weiterhin ehrlich:** getestet wurde gegen nachgebaute Meldungen, nicht am echten X1E. Klick nach dem Update einmal auf „Drucker Software Update suchen" — kommt die Marke, passt es; kommt sie nicht, klapp die Rohmeldung auf und schick sie mir, dann sehe ich das genaue Format.

## 40. Version 0.9.11 — X1E-Update erkannt, go2rtc-Killswitch, Filter „Fertig"

Drei Punkte.

**X1E-Update wurde übersehen — jetzt behoben.** Aus einer echten Meldung (woobly 7) war zu sehen: `new_version_state` steht auf 1, und `ota_new_version_number` ist trotzdem mit „01.03.00.00" gefüllt — also liegt ein Update vor. Der Zustandswert taugt nicht als Schalter. In 0.9.10 unterdrückte genau dieser Wert die Anzeige. Jetzt zählt allein, ob ein Nummernfeld gefüllt ist (`ota_`, `ams_`, `ahb_`, `ext_new_version_number`) und kein Platzhalter (`00.00.00.00`). Die laufende Fassung wird weiterhin über `get_version` gegengeprüft, damit Installiertes verschwindet. Zum Gegentest ist die echte H2-Meldung (woobly h12, alle Felder leer, kein Update) als Testfall hinterlegt — sie darf nichts anzeigen.

**go2rtc-Killswitch.** Es kam vor, dass go2rtc mehrfach lief (bis zu zehnmal), und ein Programm-Update griff nicht, solange ein alter Prozess noch Dateien und den Port belegte. Neu:

- Beim Start räumt das Programm verwaiste go2rtc-Prozesse auf, bevor es seinen eigenen startet.
- Im go2rtc-Bereich steht die Zahl der laufenden Prozesse (rot, wenn mehr als einer läuft) und ein roter Knopf „🧯 Alle go2rtc beenden". Der beendet alle go2rtc-Prozesse — auch verwaiste, samt Kindprozessen — und startet danach genau einen frischen. Unter Windows über `taskkill /F /IM go2rtc.exe /T`.

**Filter „Fertig".** Neben Online/Offline/Pausiert/Favoriten gibt es jetzt „Fertig". Er zeigt Drucker, die 100 % erreicht haben (Zustand FINISH, ersatzweise Fortschritt ≥ 100 und nicht laufend). Sobald ein neuer Druck startet (RUNNING), fällt die Markierung weg. Wie die anderen Pillen lässt er sich in den Einstellungen aus- und einblenden.

Geprüft: vierzehn Go-Tests (u. a. echte X1E- und H2-Meldung), jsdom-Prüfstände für den Fertig-Filter (FINISH ja, RUNNING bei 100 % nein, IDLE bei 100 % ja, weg nach Neustart; Zähler; nur Fertige sichtbar) und den Killswitch-Knopf; die bestehenden Prüfstände (Übersicht, Update-Pille) unverändert grün. `go test` und `-race` sauber, Herstellername 0×.

**Ehrlich:** Killswitch und Startbereinigung sind Windows-Pfade, die ich hier nicht am echten System ausführen kann — die Logik ist getestet, der echte `taskkill`-Lauf nicht. Sag mir, ob die Prozesszahl danach wirklich auf 1 steht.

## 41. Version 0.9.12 — Version auf dem Startbild, Copyright gekürzt

Zwei kleine Änderungen.

**Version auf der Gorilla-Ladeseite.** Auf dem Startbild steht jetzt unter dem Copyright die Versionsnummer („Version 0.9.12"). Damit sie sofort dasteht — noch bevor die Oberfläche irgendetwas nachlädt — wird sie serverseitig in die Seite eingesetzt: ein Platzhalter `__APP_VERSION__` wird beim Ausliefern einmalig durch die gebaute Nummer ersetzt und danach zwischengespeichert.

**Copyright gekürzt.** Aus „© Reinhard @ Brickshouse GmbH · Druckerfarm.ch" wird „© Brickshouse GmbH · Druckerfarm.ch" — an beiden Stellen (Startbild und Info-Karte; die Info-Karte speist auch das „Über"-Fenster).

Geprüft: Go-Test (Platzhalter ersetzt, „Version …" vorhanden, altes Copyright weg, neues da), jsdom (Startbild zeigt die Version, neues Copyright, kein „Reinhard", kein übrig gebliebener Platzhalter). `go test` und `-race` sauber, Herstellername 0×.

## 42. Version 0.9.13 — Blinken: Pause ja, fertig nein

Bisher blinkte die Kammerbeleuchtung (bei den dafür vorgesehenen Modellen, Vorgabe X1E/X2D) nur bei einer Störung: PrintError, offene HMS-Meldung oder Zustand FAILED. Ein fertiger Druck (FINISH) hat dabei schon nicht geblinkt — das war also bereits richtig. Ein angehaltener Druck (PAUSE) blinkte aber ebenfalls nicht.

Neu: Geblinkt wird jetzt bei Störung ODER Pause. Ein fertiger Druck ist ausdrücklich ausgenommen — FINISH blinkt nie, auch wenn zum Abschluss noch eine Meldung ansteht. Damit signalisiert das Blinken nur noch das, was wirklich Aufmerksamkeit braucht: Fehler und angehaltene Drucke.

Umgesetzt über eine eigene Entscheidung `printerWantsBlink` (FINISH → nein, PAUSE/PAUSED → ja, sonst wie `printerHasError`); die bisherige Fehlererkennung bleibt unverändert und wird weiter für andere Zwecke genutzt.

Geprüft: Go-Test über zehn Fälle (fertig blinkt nicht — auch nicht mit Restmeldung, Pause blinkt, FAILED/PrintError/HMS blinken, laufend/idle/offline blinken nicht). `go test` und `-race` sauber, Herstellername 0×.

## 43. Version 0.9.14 — Sauberes Beenden, Lampen zurück, Lizenz-Prüfung

Drei Dinge.

**go2rtc wird beim Schließen wirklich beendet.** Bisher beendete das Programm beim Schließen nur den einen von ihm verwalteten go2rtc-Prozess — verwaiste Prozesse und Kindprozesse (ffmpeg) konnten weiterlaufen. Jetzt räumt eine eigene Beenden-Routine restlos auf: der eigene Prozess wird beendet und zusätzlich alle übrigen go2rtc-Prozesse samt Kindern (unter Windows `taskkill /F /IM go2rtc.exe /T`). Zusammen mit der Startbereinigung aus 0.9.11 kann sich so nichts mehr anhäufen.

**Signalleuchten gehen beim Schließen zurück auf Normal.** Wenn das Programm geschlossen wird, setzt es die Kammerbeleuchtung aller gerade blinkenden Drucker (X1E/X2D) von „blinkend" zurück auf „an". So bleibt im Druckerraum nach dem Beenden kein Lämpchen im Alarmblinken stehen. Das Zurücksetzen ist zeitlich begrenzt (max. ein paar Sekunden), falls ein Drucker nicht antwortet.

**Lizenz- und Rechte-Prüfung.** Auf Wunsch geprüft, ob das Programm Urheberrechte, Lizenzen oder Marken verletzt — besonders mit Blick auf Bambu Lab und Reverse-Engineering. Kurzfassung:

- Kein Herstellermaterial eingebettet: keine Zertifikate, Schlüssel, Firmware, kein Slicer-Code, Herstellername 0×.
- Das lokale Protokoll (MQTT/SSDP/FTP/RTSP) ist Interoperabilität mit eigener Hardware; Protokolle/Feldnamen sind nicht urheberrechtlich geschützt. Es wird **kein Kopierschutz umgangen** — Steuerbefehle laufen über den am Gerät aktivierten Developer Mode, nicht über einen extrahierten Herstellerschlüssel.
- Alle Open-Source-Abhängigkeiten sind freizügig oder dateiweise schwach-copyleft; keine verlangt die Offenlegung des eigenen Quellcodes. Die einzige Pflicht — die Nennung — wird jetzt mit `THIRD_PARTY_LICENSES.md` erfüllt und zusätzlich in der „Über"-Karte angezeigt.
- go2rtc und ffmpeg werden nur zur Laufzeit nachgeladen und nicht mitverbreitet — das sollte so bleiben (sonst greift die GPL von ffmpeg).
- Offen bleibt allein die Herkunft der eigenen Bildmittel (Startbild/Logo) — bitte sicherstellen, dass sie selbst erstellt oder lizenziert sind.

Der ausführliche Befund steht in `LIZENZ-PRUEFUNG.md`, die Nennung in `THIRD_PARTY_LICENSES.md`. Beide liegen im Ordner. (Keine Rechtsberatung — im Zweifel eine IP-Kanzlei hinzuziehen.)

Geprüft: `go test` und `-race` sauber, jsdom-Smoke-Test (Startbild mit Version und neuem Copyright, OSS-Hinweis vorhanden, kein „Reinhard"), Herstellername 0×.

## 44. Version 0.9.15 — Sprachwahl beim ersten Start, Ladeseite, Lizenzlink

Drei Punkte.

**Sprachwahl beim allerersten Start.** Beim ersten Öffnen fragt Druckerfarm einmalig nach der Sprache — zwei große Schaltflächen „🇩🇪 Deutsch" und „🇬🇧 English" über dem Startbild. Die Wahl wird dauerhaft gespeichert (`sprache_gewaehlt` im Zustand) und kommt danach nicht wieder. Ändern lässt sie sich weiterhin jederzeit in den Einstellungen.

**Ladeseite ohne Icon-Kästchen.** Die erste Ladeseite (noch vor dem eigentlichen Fenster) zeigte ein türkises Kästchen mit einem Kamera-Symbol. Das ist weg — stattdessen liegt jetzt das Gorilla-Startbild als Hintergrund, mit „Druckerfarm" und dem Ladebalken. So ist der Übergang zum App-Startbild nahtlos.

**Lizenzen aus der App heraus lesbar.** Die Datei THIRD_PARTY_LICENSES.md ist jetzt in die Programmdatei eingebettet und über die Adresse `/licenses` abrufbar. In der „Über"-Karte ist der Dateiname als Link angelegt — ein Klick öffnet die Lizenzübersicht. So findet man sie auch ohne den Ordner zu durchsuchen.

Geprüft: Go-Test (Lizenzseite liefert die eingebetteten Einträge, richtiger Content-Type), jsdom (Sprachauswahl erscheint beim ersten Start, verschwindet nach der Wahl, kommt bei bereits gewählter Sprache nicht, speichert das Kennzeichen; Lizenzlink zeigt auf /licenses), Ladeseite ohne Emoji. `go test` und `-race` sauber, Herstellername 0×.

## 45. Version 0.9.16 — Lizenzen im Explorer, schlanke Lizenzdatei, Sprachwahl ohne Symbole

Drei Anpassungen.

**Lizenzen öffnen im Datei-Explorer statt im Browser-Tab.** Der Link „THIRD_PARTY_LICENSES.md" in der „Über"-Karte öffnet die Datei jetzt nicht mehr in einem neuen Tab. Stattdessen schreibt das Programm die Lizenzdatei ins Datenverzeichnis und zeigt sie markiert im Windows-Explorer an (`explorer /select,…`; unter macOS „im Finder zeigen", unter Linux der Ordner).

**Lizenzdatei ohne Erklärungstexte.** THIRD_PARTY_LICENSES.md enthält jetzt nur noch die Nennung der Komponenten (Name, Version, Lizenz, Rechteinhaber) und die eigentlichen Lizenztexte — kein erklärendes Beiwerk mehr.

**Sprachwahl ohne Symbole.** Die Abfrage beim ersten Start zeigt nur noch „Deutsch" und „English" — keine Flaggen, keine Kürzel „DE"/„EN".

Geprüft: Go-Test (Lizenzdatei wird ins Datenverzeichnis geschrieben, Erklärungstexte sind entfernt), jsdom (Sprachbuttons ohne Flaggen/Kürzel, Lizenzlink löst das Explorer-Öffnen aus statt eines Tabs). `go test` und `-race` sauber, Herstellername 0×.

## 46. Version 0.9.17 — go2rtc-Neustart repariert

Der Neustart von go2rtc im Programm griff nicht. Ursache: der Neustart beendete nur den **selbst verwalteten** go2rtc-Prozess (`stopGo2rtc`). Hing daneben noch ein verwaister go2rtc (wie bei den früher beobachteten Mehrfach-Instanzen), belegte der weiterhin Port 1984 — und der frisch gestartete go2rtc konnte nicht binden und verschwand sofort wieder. Von außen sah es aus, als ließe sich go2rtc nicht neu starten.

Neu gibt es einen gründlichen Neustart (`neustartGo2rtc`): erst den eigenen Prozess beenden, dann **alle** noch laufenden go2rtc-Prozesse samt Kindern beseitigen (Windows `taskkill /F /IM go2rtc.exe /T`), kurz warten, bis Port und Logdatei frei sind, und genau einen frischen starten. Dieser gründliche Neustart wird jetzt überall verwendet: beim Neustart-Knopf, beim automatischen Neustart nach Konfigurationsänderungen und in der Selbstheilung, wenn go2rtc nicht mehr antwortet.

Geprüft: `go test` und `-race` sauber, Herstellername 0×. Der eigentliche `taskkill`-Lauf lässt sich nur am echten Windows prüfen — sag mir, ob der Neustart-Knopf jetzt greift.

## 47. Version 0.9.18 — Übersetzungen als JSON, erweiterbar vor dem Bau

Die Übersetzungen lagen bisher fest im Code, und der Großteil des sichtbaren Textes lief gar nicht über das Wörterbuch — deshalb blieb auf Englisch fast alles deutsch. Umgestellt:

**Je Sprache eine JSON-Datei.** Unter `lang/` liegen jetzt `de.json` und `en.json`. Sie werden beim Bauen fest in die Programmdatei eingebettet und beim Ausliefern der Seite eingesetzt. Wer eigene Übersetzungen will, bearbeitet die Datei **vor** dem Erstellen der Exe und baut neu — Anleitung in `lang/README.md`.

**Neuer, generischer Weg zum Übersetzen.** Jedes Element mit dem Attribut `data-i18n="schluessel"` bekommt seinen Text automatisch aus der JSON (dazu `data-i18n-html` für formatierten Text und `data-i18n-ph` für Platzhalter). So lässt sich weiterer Text anbinden, ohne im Code je eine Zeile zu ergänzen: Schlüssel in beide JSONs, `data-i18n` ans Element.

**Mehr Text übersetzt.** Bereits angebunden sind u. a. die Filter-Pillen (Offline, Pausiert, Favoriten, Fertig), die Stream-Leiste, die Vollbild-Reiter und die Firmware-Update-Karte samt Knopf. Die vorhandenen ~100 Schlüssel sind auf Englisch vollständig; neu hinzugekommen sind weitere.

Ehrlich zum Stand: Es ist noch nicht die komplette Oberfläche verschlagwortet — einige dynamisch aufgebaute Bereiche (Detailmenü, go2rtc-Steuerung im Detail) folgen. Die Mechanik steht aber, und weitere Bereiche sind jetzt schnell und gefahrlos nachziehbar. Sag mir, welche Ansicht dir zuerst wichtig ist, dann verschlagworte ich sie.

Geprüft: Go-Test (Sprachdateien werden injiziert, Platzhalter weg, deutsche und englische Einträge in der Seite), jsdom (Umschalten auf Englisch ändert die getaggten Elemente „Pausiert→Paused", „Fertig→Done", Firmware-Knopf; zurück auf Deutsch klappt). `go test` und `-race` sauber, Herstellername 0×.

## 48. Version 0.9.19 — Vollständige Englisch-Übersetzung

In 0.9.18 war die Mechanik da, aber der Großteil des sichtbaren Textes lief nicht darüber. Jetzt wird bei englischer Sprache die **komplette** Oberfläche automatisch übersetzt — Überschriften, Beschreibungen, Bezeichnungen, Buttons, Listen, Meldungen und Menüs, egal wo sie eingesetzt werden.

**Wie es funktioniert.** Deutsch ist die Ausgangssprache im Code. Eine neue Datei `lang/de-en.json` ordnet deutschen Text dem englischen zu (rund 190 Einträge). Bei englischer Sprache übersetzt ein DOM-Übersetzer den gerenderten Text anhand dieser Tabelle. Damit auch später aufgebauter Text erfasst wird (Kacheln, Detailmenüs, Toasts, Suchergebnisse), beobachtet ein MutationObserver die Oberfläche und übersetzt jede Änderung sofort nach. So muss nicht jede einzelne Textstelle im Code angefasst werden — und es bleibt nichts stehen.

**Umschalten.** Ein Sprachwechsel in den Einstellungen lädt die Seite neu: Deutsch ist die Quelle, Englisch wird beim Laden erzeugt. Dadurch ist der Wechsel in beide Richtungen sauber. Die Erstauswahl beim Programmstart übersetzt sofort, ohne Neuladen.

**Selbst erweitern.** Findet sich im englischen Modus doch noch deutscher Text, genügt ein neuer Eintrag in `lang/de-en.json` (deutscher Text → englischer Text) und ein Neubau. Anleitung in `lang/README.md`. Auch Titel (Tooltips) und Eingabe-Platzhalter werden übersetzt.

Geprüft: Go-Test (Übersetzungstabelle wird injiziert, Platzhalter weg, englische Einträge in der Seite), jsdom (auf Englisch werden statische Überschriften „Übersicht→Overview", „🌐 Netzwerkverkehr→Network traffic", die Firmware-Karte und die Pillen übersetzt; ein nachträglich eingefügter deutscher Knoten wird vom Observer übersetzt; Tooltip und Text ebenso). `go test` und `-race` sauber, Herstellername 0×.

Ehrlich: Die Tabelle deckt die heutige Oberfläche ab. Kommt neuer deutscher Text hinzu, muss er (einmalig) in `de-en.json` ergänzt werden — sonst bliebe genau dieser Text englisch-seitig deutsch. Sag mir, falls dir irgendwo noch etwas Deutsches auffällt, dann trage ich es nach.

## 49. Version 0.9.20 — Startzustand sichtbar, ein Gorilla, Aufwachen nach Standby

Vier Punkte.

**Zahnrad zeigt „arbeitet" statt nur rot.** Beim Start war das Einstellungen-Zahnrad rot, bis go2rtc lief. Jetzt pulsiert es in der Startphase (und bei einem Neustart) gelb — „läuft gerade an" — und wird erst rot, wenn go2rtc nach rund 45 Sekunden immer noch nicht antwortet. Ist go2rtc da, ist das Zahnrad normal.

**Video-Kästchen mit Rückmeldung.** Solange go2rtc startet, zeigt jedes Kästchen einen Ladekreis mit „go2rtc startet … bitte warten", statt schwarz zu bleiben — man sieht, dass man warten muss. Ist ein Stream gestoppt und go2rtc bereit, steht „▶ bereit" im Kästchen. Und beim Starten eines Videos liegt ein „verbinde …"-Ladekreis über dem Bild, bis der Stream wirklich läuft — kein schwarzes Videofeld mehr, bei dem man nicht weiß, ob etwas passiert.

**Nur noch ein Gorilla-Bild.** Das allererste, sofort sichtbare Hintergrundbild war ein anderes als das Bild beim Klick aufs Logo. Jetzt ist die Sofortvorschau eine Miniaturversion desselben Gorilla-Bildes (`splash.jpg`) — überall dasselbe Motiv, kein Bildwechsel mehr.

**Selbst-Neuladen nach dem Ruhemodus.** Nach Standby und erneutem Anmelden hing das Fenster oft (der Server lief, aber die Seite war tot). Zwei Wächter fangen das ab: eine große Zeitlücke zwischen zwei Takten (der Rechner hat geschlafen) und das Sichtbar-/Online-Werden des Fensters. Dann wartet das Programm, bis der Server wieder antwortet, zeigt kurz „Verbindung wird wiederhergestellt …" und lädt die Seite frisch. Während einer gewollten Netzwerkpause greift der Wächter nicht.

Übersetzungen: die neuen Texte sind in de/en.json bzw. de-en.json ergänzt und erscheinen im Englisch-Modus auf Englisch.

Geprüft: jsdom (Zahnrad „arbeitet" während des Starts; Warte-Kästchen mit Text; „▶ bereit" nach Stopp; „verbinde …"-Overlay beim Live-Start; Reconnect-Overlay vorhanden; keine Laufzeitfehler). `go test` und `-race` sauber, Herstellername 0×.

Ehrlich: Das Auto-Neuladen greift bei einer Zeitlücke über 60 Sekunden. Backgroundet man das Fenster sehr lange, kann es also auch mal neu laden — der Zustand liegt aber ohnehin serverseitig, es geht nichts verloren. Sag mir, falls es zu oft auslöst, dann drehe ich an der Schwelle.

## 50. Version 0.9.21 — go2rtc-Neustart repariert, Wächter entschärft (Stufe 1)

Drei Punkte aus der Rückmeldung.

**go2rtc-Neustart lädt wieder.** Der Neustart war nicht abgesichert: der frische Prozess wurde nicht geprüft, und zwei gleichzeitige Neustarts (Knopf + automatischer Anlass) konnten sich gegenseitig abwürgen — der eine beendete den frisch gestarteten des anderen. Jetzt läuft immer nur EIN Neustart (Sperre), es wird gewartet, bis Port 1984 wirklich frei ist, dann gestartet und geprüft, ob go2rtc tatsächlich hochkommt; klappt es nicht, räumt das Programm auf und startet ein zweites Mal. Die Oberfläche zeigt während des Neustarts das gelb pulsierende Zahnrad und sieht mehrfach nach (statt nur einmal nach 2 Sekunden), sodass sich der Zustand von selbst erholt.

**Reconnect-Overlay ist abbrechbar.** Das abgedunkelte „Verbindung wird wiederhergestellt …"-Overlay hat jetzt zwei Knöpfe — „Abbrechen" (weiterarbeiten) und „Jetzt neu laden". Man bleibt nicht mehr darin stecken.

**Wächter lädt nicht mehr blind neu.** Der 60-Sekunden-Wächter unterscheidet jetzt: War das Fenster nur minimiert oder der Rechner im Schlaf, wird beim Sichtbarwerden zuerst geprüft, ob der Server erreichbar ist. Ist er es (der Normalfall — ein laufender Upload arbeitet im Backend ohnehin weiter), wird nur die Anzeige aufgefrischt, **ohne Neuladen**. Nur wenn der Server wirklich nicht antwortet, erscheint das abbrechbare Overlay und lädt neu, sobald er wieder da ist. Während einer gewollten Netzwerkpause bleibt der Wächter aus.

Zur Einordnung deiner Sorge: Datei-Uploads laufen im Go-Backend, nicht im Browser. Minimieren drosselt nur die Oberfläche; die Übertragung läuft weiter. Selbst ein Neuladen würde einen serverseitigen Upload nicht abbrechen. Das größere „Aufträge/Worker"-Panel mit Fortschritt und Abbrechen kommt als Stufe 2.

Übersetzungen ergänzt (Abbrechen/Reload). Geprüft: jsdom (Overlay-Knöpfe, Abbrechen schließt, bei erreichbarem Server nur Auffrischen ohne Overlay), `go test` und `-race` sauber, Herstellername 0×.

Ehrlich: Den `taskkill`-Neustart und das echte Standby-Verhalten kann ich nur am Windows-Gerät nicht selbst auslösen — sag mir, ob der go2rtc-Neustart jetzt greift und ob nach dem Aufwachen alles sauber weiterläuft.

## 51. Version 0.9.22 — Aufträge-/Worker-Panel (Stufe 2)

Ein gemeinsames „Aufträge"-Panel macht sichtbar, woran das Programm gerade arbeitet — und lässt es abbrechen.

**Header-Indikator.** Oben rechts steht „⚙ Aufträge" mit einem Zähler, sobald etwas läuft. Ein Klick öffnet ein Panel am rechten Rand.

**Das Panel** listet alle aktiven und kürzlich beendeten Vorgänge mit Fortschrittsbalken, Zähler (erledigt/gesamt), Fehlerzahl und Status (läuft/angehalten/fertig). Enthalten sind:

- die Server-Vorgänge aus der Job-Registry — **Suche** und **Löschen** über viele Drucker (mit begrenzter Gleichzeitigkeit, damit die Drucker nicht überrannt werden),
- der **Datei-Sync** als eigener Eintrag,
- der laufende **Upload** aus dem Browser.

Jeder laufende Auftrag hat ein **Abbrechen**. Server-Jobs werden über den bestehenden Stopp-Weg beendet, der Sync über seine Stopp-Funktion, der Upload über den vorhandenen Abbruch. Neu ist der Sammel-Endpunkt `/api/jobs`, der alles zusammenführt; die Oberfläche fragt ihn im Takt ab und hält Zähler und Panel aktuell.

Damit sieht man jederzeit, dass „er gerade beschäftigt ist" — genau die Rückmeldung, die gefehlt hat. Der 60-Sekunden-Wächter aus 0.9.21 lädt ohnehin nicht mehr blind neu; zusammen mit dem sichtbaren Panel ist ein laufender Upload jetzt gut aufgehoben.

Alles übersetzt (Panel-Texte in de/en.json, Beschriftungen über die Auto-Übersetzung). Geprüft: jsdom (Panel listet Server-Jobs, Badge zählt aktive, Abbrechen ruft den richtigen Endpunkt, Labels korrekt), `go test` und `-race` sauber, Herstellername 0×.

Ausblick: Netzwerk-Suche und Firmware-Prüfung laufen noch nicht über die Registry (sind aber kurz). Wenn du magst, ziehe ich die auch ins Panel — sag Bescheid.

## 52. Version 0.9.23 — Gerätename in der Netzwerksuche übernehmen

Bei der Netzwerksuche werden Drucker, die schon in der Liste stehen, nur gezeigt, wenn sich etwas geändert hat. Bisher galt das nur für eine neue IP-Adresse — ein in der Hersteller-App geänderter **Name** blieb unsichtbar.

Neu: Meldet ein bekannter Drucker (gleiche Adresse) einen anderen Namen als den gespeicherten, erscheint er im Abschnitt „Name im Drucker geändert". Dort stehen beide Namen nebeneinander — der im Programm gespeicherte und in Klammern der vom Drucker gemeldete — und ein Knopf „Namen übernehmen" schreibt den Gerätenamen in die gespeicherte Druckerliste. (Die Doppelanzeige „Name (Gerätename)" gab es schon; jetzt wird der Abweichungsfall auch tatsächlich angezeigt.)

Zur Frage, den Druckernamen über das Netzwerk zu ändern: Das geht nicht. Recherche und die Hersteller-Diskussionen bestätigen — der Name lässt sich nur über die Cloud (Bambu Studio bzw. die Handy-App) ändern; im „Nur LAN"-Betrieb ist Umbenennen nicht vorgesehen (dafür gibt es sogar einen offenen Feature-Request beim Hersteller). Das Programm kann den Namen deshalb nur auslesen und übernehmen, nicht auf den Drucker schreiben. Wer wirklich umbenennen will, muss den Drucker vorübergehend an ein Konto binden, dort umbenennen und wieder auf LAN zurückstellen.

Übersetzt (Abschnittstitel und Knopf). Geprüft: jsdom (Abweichungs-Abschnitt erscheint, beide Namen sichtbar, „Namen übernehmen" schickt den neuen Namen an den Server), `go test` und `-race` sauber, Herstellername 0×.

## 53. Version 0.9.24 — Programm-Update ab Werk eingerichtet

Die Update-Funktion war vollständig vorhanden (unter Einstellungen → „⬆ Programmversion": aktuelle Fassung, „Nach Update suchen", „Update einspielen", Fortschritt und Release-Notizen), aber ohne hinterlegte Bezugsquelle — deshalb kam „kein Repository hinterlegt".

Jetzt ist als Vorgabe das offizielle Repository **BrickshouseGmbH/druckerfarm** einkompiliert. Die Update-Suche funktioniert damit sofort ohne Einrichtung: das Programm holt die neueste Release von GitHub, vergleicht die Version, lädt bei Bedarf die neue `druckerfarm.exe` (per SHA nachweisbar), ersetzt die laufende Datei und startet neu. Über das Feld in den Einstellungen lässt sich die Quelle weiterhin überschreiben (z. B. für ein eigenes/privates Repo mit Token).

Damit Updates gefunden werden, muss zu einer Release eine `druckerfarm.exe` als Anhang hochgeladen sein und die Release-Version höher als die laufende liegen.

Geprüft: Go-Test (ohne Einstellung wird das Standard-Repo verwendet, eigene Einstellung hat Vorrang, Konfigurations-Endpunkt liefert das Standard-Repo), `go test` und `-race` sauber, Herstellername 0×.

## 54. Version 0.9.25 — Keine Edge-Fehlerseite mehr, Umbenennen ohne Code

Zwei Korrekturen.

**Reconnect zeigt keine Edge-Fehlerseite mehr.** Der Wächter hat nach dem Aufwachen die Seite per Neuladen aufgefrischt. Traf das Neuladen einen Moment, in dem der Server (127.0.0.1) gerade nicht antwortete, zeigte Edge seine eigene Seite „Hmmm… diese Seite ist leider nicht erreichbar / ERR_CONNECTION_REFUSED" — und aus der kommt unser Programm nicht mehr heraus, weil die eigene Oberfläche dann weg ist.

Jetzt wird **nicht mehr navigiert**. Solange unsere Seite geladen bleibt, kann sie sich immer wieder selbst erholen: Ist der Server erreichbar, werden nur die Daten neu geholt (Status, Fehler, go2rtc, Aufträge, Videos) — ohne Seitenwechsel. Ist der Server nicht erreichbar, bleibt das abbrechbare Overlay stehen und wartet; sobald der Server wieder antwortet, frischt die Seite still auf. So kann die Edge-Fehlerseite gar nicht mehr auftauchen. Der Knopf „Jetzt neu laden" bleibt für den Fall, dass man es doch von Hand auslösen will.

**Drucker umbenennen ohne Zugangscode.** Beim Bearbeiten wurde bisher ein Zugangscode verlangt — dadurch ließ sich ein Drucker ohne Code (z. B. frisch aus der Netzwerksuche übernommen) nicht umbenennen. Jetzt genügt der Name; Zugangscode und Modell sind fürs Umbenennen nicht mehr nötig. Damit kann man einen Drucker jederzeit benennen, auch ohne mit ihm verbunden zu sein. Vorhandener Code und vorhandenes Modell bleiben dabei erhalten.

Übersetzt. Geprüft: jsdom (Reconnect frischt bei Rückkehr des Servers auf, ohne die Seite neu zu laden — kein `location.reload`; Umbenennen ohne Code sendet den neuen Namen an den Server), `go test` und `-race` sauber, Herstellername 0×.

## 55. Version 0.9.26 — Sortierbare Spalten, natürliche Reihenfolge, Online/Offline-Pille

**Online/Offline-Pille hinter dem Namen.** In der Übersicht (Kacheln und Liste) und in FileSync steht jetzt hinter dem Druckernamen eine Pille „Online" (grün) bzw. „Offline" (rot).

**Sortierbare Spalten.** Die Listenansicht der Übersicht hat eine klickbare Tabellenüberschrift: Modell, Name, Status, Datei, Fortschritt, Restzeit, IP. Ein Klick sortiert nach dieser Spalte, ein weiterer dreht die Richtung um (Pfeil ▲/▼ zeigt es an). Neu ist dabei die Spalte „Status" (online/offline). In FileSync lässt sich die Druckerliste über eine Sortierleiste nach Name, Modell oder Status ordnen.

**Natürliche Reihenfolge.** Nach Name wird jetzt „natürlich" sortiert: „Drucker 2" kommt vor „Drucker 10" (und nicht die 10 direkt hinter der 1). Das gilt für die Druckerlisten und auch für die Dateilisten in FileSync, die standardmäßig natürlich nach Dateinamen sortiert sind.

Übersetzt (Spaltentitel und Pillen). Geprüft: jsdom (natürliche Sortierung 1→2→10, Online/Offline-Pillen vorhanden, Klick auf eine Spalte sortiert Liste und FileSync-Liste), `go test` und `-race` sauber, Herstellername 0×.

Zu den zwei Streams mit „go2rtc nicht erreichbar / context deadline exceeded" für einzelne Drucker: Das ist ein Snapshot-Zeitlimit — go2rtc/der Drucker liefert das Bild nicht rechtzeitig; die App pausiert diesen einen Stream und versucht es in 30 s erneut. Das betrifft nur die betroffenen Geräte (Last/Erreichbarkeit), nicht die Farm. Wenn es dauerhaft bei denselben zweien bleibt, lohnt ein Blick auf deren Netzw/WLAN-Signal; sag Bescheid, wenn ich das Zeitlimit oder die Wiederholrate anpassen soll.

Für die Dateiliste in FileSync sind die Spalten aktuell nach Name (natürlich) vorsortiert; klickbare Spaltenköpfe (Name/Größe) dort reiche ich nach, wenn gewünscht.

## 56. Version 0.9.27 — Sortierung auch im Grid (Druckerübersicht)

Nachtrag zu 0.9.26: Die Sortierung war nur in der Listenansicht, nicht in der Standard-Übersicht (Grid). Jetzt hat auch das Grid eine Sortierleiste („Sortieren: Name · Status · Modell · Fortschritt · Restzeit · IP"). Ein Klick sortiert die Kacheln, ein weiterer dreht die Richtung.

Die Sortierung ist jetzt einheitlich: Grid und Liste teilen sich dieselbe Auswahl (die Filterung/Sortierung sitzt zentral in der Datenquelle), sodass ein in der einen Ansicht gewählter Sortierschlüssel auch in der anderen gilt. Name wird natürlich sortiert (1→2→10).

Geprüft: jsdom (Grid zeigt die Sortierleiste, sortiert natürlich, Richtung umkehrbar, Online-Pille vorhanden), `go test` und `-race` sauber, Herstellername 0×.

## 57. Version 0.9.28 — go2rtc wird sofort beendet (nicht mehr ~10 s)

Warum go2rtc rund zehn Sekunden zum Beenden brauchte: Das Programm hat nur den **go2rtc-Elternprozess** beendet (Process.Kill). go2rtc startet für die Bilder aber **ffmpeg-Kindprozesse** — die blieben verwaist zurück und hingen noch rund zehn Sekunden an ihren RTSP-Verbindungen, bis sie von selbst aufgaben. Das war die Verzögerung.

Jetzt wird der **ganze Prozessbaum** auf einmal beendet: go2rtc samt seinen ffmpeg-Kindern, forciert (unter Windows `taskkill /F /T /PID`). Das geht praktisch sofort. Betroffen sind alle Stopp-Wege — Netzwerkpause, Neustart und Beenden des Programms. Als Nebeneffekt gibt der Port schneller frei, wodurch auch der go2rtc-Neustart flotter durchläuft.

Geprüft: `go test` und `-race` sauber, Cross-Build für Windows ok, Herstellername 0×. Den echten `taskkill`-Lauf kann ich nur am Windows-Gerät nicht selbst prüfen — sag mir, ob go2rtc jetzt zügig verschwindet.

## 58. Version 0.9.29 — Sortierung auch bei „Gespeicherte Drucker"

Nachtrag: Die Liste „Gespeicherte Drucker" (Tab „+ Drucker") hatte noch keine Sortierung. Jetzt gibt es dort eine klickbare Tabellenüberschrift — Modell, Name, Status, Zugangscode, Seriennummer, IP. Klick sortiert, nochmal Klick dreht die Richtung (Pfeil ▲/▼). Der Name wird natürlich sortiert (Drucker 2 vor Drucker 10).

Zusätzlich ist eine Status-Spalte mit Online/Offline-Pille dazugekommen, passend zu Übersicht und FileSync. Die Bearbeiten-Zeile wurde an das neue Spaltenraster angepasst.

Übersetzt (Spaltentitel). Geprüft: jsdom (Kopfzeile da, natürliche Sortierung 1→2→10, Online-Pille vorhanden, Klick sortiert z. B. nach IP, Bearbeiten-Zeile korrekt ausgerichtet), `go test` und `-race` sauber, Herstellername 0×.

## 59. Version 0.9.30 — go2rtc beim Beenden zuerst, kein Reconnect-Overlay mehr

Drei zusammenhängende Korrekturen aus der Rückmeldung.

**Beenden killt go2rtc jetzt zuerst.** Beim Schließen wurden bisher erst die Signalleuchten zurückgesetzt (mit bis zu 6 Sekunden Wartezeit auf antwortende Drucker) und go2rtc erst danach beendet — deshalb lief go2rtc so lange weiter. Die Reihenfolge ist umgedreht: go2rtc (samt ffmpeg-Kindern, forciert) wird sofort beendet, danach werden die Lampen zurückgesetzt. go2rtc verschwindet damit unmittelbar.

**Kein Reconnect-Overlay mehr.** Der Wächter hat nach Standby/Minimieren den Server geprüft und, sobald diese Prüfung auch nur kurz ins Zeitlimit lief, ein abgedunkeltes „Verbindung wird wiederhergestellt"-Fenster gezeigt. Das schlug oft grundlos an, löste nichts und war lästig — und „Abbrechen" konnte einen zurück auf die Edge-Fehlerseite werfen. Das Overlay ist ersatzlos entfernt.

**Stille Erholung stattdessen.** Nach Aufwachen oder Sichtbarwerden frischt das Programm nur noch im Hintergrund die Daten auf (Status, Fehler, go2rtc, Aufträge) und hängt die Videos wieder an — ohne Fenster, ohne Seitenwechsel. Ist der Server im Moment nicht erreichbar, greifen die vorhandenen Fehlerbehandlungen; sobald er wieder da ist, füllt sich die Anzeige beim nächsten Takt von selbst.

Warum überhaupt „reconnected" wurde: Der Wächter fragte nach einer Pause `/ready` ab; ein einzelnes Zeitlimit (z. B. unter Last oder während go2rtc gerade neu startete) reichte, um das Overlay auszulösen. Jetzt gibt es kein Overlay und keine Navigation mehr — es kann also auch nichts „Komisches" mehr aufpoppen.

Geprüft: jsdom (kein Overlay-Element mehr, stille Erholung lädt Daten neu, keine Restfunktionen/Fehler), `go test` und `-race` sauber, Herstellername 0×. Ob go2rtc beim Schließen jetzt sofort weg ist, bitte am echten Gerät gegenchecken.

## 60. Version 0.9.31 — fünf Korrekturen

**1. Modellcode „O1C" → H2C.** Die H2-Reihe meldet sich nach dem Muster O1x (O1D = H2D, O1C = H2C; in der Praxis auch als „O1C2"). Beide Varianten werden jetzt korrekt auf H2C übersetzt, statt „O1C2" anzuzeigen.

**2. Duplikatprüfung über die Seriennummer statt der IP.** Ein neuer Drucker ließ sich nicht anlegen, wenn ein anderer (offline, mit inzwischen veralteter IP) noch dieselbe IP trug. Jetzt entscheidet die Seriennummer, ob ein Drucker schon existiert — eine IP-Kollision blockiert nicht mehr, denn die Live-IP des neuen Geräts ist die richtige. Nur ohne Seriennummer (rein manuell) bleibt die IP der Anhaltspunkt. Gilt für Anlegen, Übernehmen aus der Suche und CSV-Import.

**3. FileSync aktualisiert sich mit.** Nach dem Hinzufügen oder Entfernen eines Druckers wird die Druckerliste in FileSync jetzt gleich mit aktualisiert (lief bisher nicht mit).

**4. „Fullscreen Video Only" mit Statusrahmen und Balken.** Jede Videobox bekommt einen Rahmen in Statusfarbe (grün = Druck läuft, gelb = pausiert, rot = Fehler, grau = fertig) und unten einen Fortschrittsbalken — bewusst ohne Text. Rahmen und Balken laufen live mit dem Status mit.

**5. Firmware-Update vom Tool aus anstoßen — mit deutlicher Warnung.** Nach der Update-Suche erscheint zu jedem Drucker mit offenem Update ein Knopf „⬆ Update starten". Ein Klick zeigt zuerst eine klare Warnung: besser direkt am Drucker starten, nur mit aktiviertem Developer Mode, ein unterbrochenes Update kann den Drucker beschädigen, während des Updates nicht ausschalten, und dass dieser Weg experimentell ist und nicht sicher funktioniert. Erst nach Bestätigung wird der Befehl per lokalem MQTT gesendet.

Ehrlich zu Punkt 5: Das Auslösen per LAN-MQTT ist vom Hersteller nicht offiziell dokumentiert (es gibt einen offenen Feature-Request dafür) und unterliegt derselben Signaturprüfung wie die übrigen Steuerbefehle — es funktioniert also nur mit Developer Mode und ist Best-Effort. Deshalb die Warnung. Zuverlässig bleibt der Weg direkt am Drucker.

Geprüft: Go-Tests (O1C→H2C, Serial-Duplikatlogik: gleiche IP/andere Serial legt an, gleiche Serial = 409), jsdom (FileSync zieht mit, Fullscreen-Rahmen+Balken ohne Text, Firmware-Update-Liste + Start-Knopf + Warndialog löst den Befehl aus). `go test` und `-race` sauber, Herstellername 0×.

## 61. Version 0.9.32 — Fullscreen-Feinschliff, Zugriffs-Log & Parallel-PC-Hinweis

**Fullscreen Video Only (drei Anpassungen):**

- Der Rahmen um jede Videobox ist jetzt 1 px statt 3 px, und die Zelle wird auf volle Höhe gestreckt (kein festes 16:9-Seitenverhältnis mehr).
- Im Fullscreen hat das Videobild kein `object-fit` und keine feste Höhe mehr — es füllt die Box entsprechend.
- Oben links in jeder Box steht jetzt der Druckername (kleines Schild). Statusrahmen und Fortschrittsbalken bleiben.

**Zugriffs-Protokoll und Hinweis auf parallele PCs.** Weil Druckerfarm bei euch aus einem geteilten Ordner läuft, können mehrere PCs gleichzeitig auf dieselben Drucker zugreifen. Neu:

- Im Programmordner (der geteilte Ort) wird `zugriff.log` geführt — eine Zeile je Start und Beenden mit Zeitstempel und „Benutzer@PC-Name". So ist nachvollziehbar, wann sich von welchem PC jemand eingeloggt hat.
- Jede laufende Instanz legt in `praesenz/` ein „ich lebe"-Zeichen ab (je PC eine Datei, minütlich aufgefrischt). Läuft ein anderer PC parallel, erscheint oben ein gelbes Band: „⚠ Ein anderer PC läuft ebenfalls: benutzer@PC".

Ehrlich zur Einschränkung: Die Parallel-Erkennung hängt daran, wie schnell der geteilte Ordner synchronisiert (OneDrive braucht dafür Sekunden bis Minuten). Es ist ein Hinweis, keine Echtzeit-Sperre. Liegt der Programmordner nicht auf einem geteilten Laufwerk, bleibt die Funktion einfach still (kein Fehler). Wenn du lieber einen festen gemeinsamen Pfad (Netzlaufwerk) dafür hättest, baue ich das als Einstellung ein.

Geprüft: Go-Test (andere Instanz mit frischem Zeitstempel wird erkannt, eigener/alter Eintrag nicht), jsdom (Fullscreen-Zelle mit Namensschild, Statusrahmen und Balken; Parallel-Banner erscheint/verschwindet). `go test` und `-race` sauber, Herstellername 0×.

## 62. Version 0.9.33 — Gemeinsamer Ordner für die Mehr-PC-Erkennung einstellbar

Nachtrag zur Parallel-PC-Erkennung: Der gemeinsame Ordner lässt sich jetzt einstellen (Einstellungen → „👥 Mehrere PCs").

Zwei PCs bemerken sich nur, wenn sie einen gemeinsam sichtbaren Ort haben, an dem sie ein „ich laufe"-Zeichen ablegen. Bisher war das fest der Programmordner (bei euch OneDrive) — was funktioniert, aber je nach OneDrive-Synchronisation verzögert ist. Neu: Im Feld „Ordner" lässt sich ein beliebiger gemeinsamer Pfad angeben, z. B. ein **Netzlaufwerk** (`\\server\druckerfarm`), das nahezu sofort synchronisiert. Leer lassen heißt weiterhin: Programmordner. Die Einstellung greift sofort, ohne Neustart, und die Karte zeigt den tatsächlich genutzten Ordner sowie welche anderen PCs gerade laufen.

Das ist die bessere Lösung, nach der ich gefragt hatte: Sie funktioniert auch, wenn die Exe lokal (nicht geteilt) liegt — dann trägt man einfach einen Netzwerkpfad ein — und ein schnelles Laufwerk macht die Erkennung praktisch in Echtzeit statt mit OneDrive-Verzögerung.

Geprüft: jsdom (Einstellungsfeld vorhanden, Speichern sendet den Pfad, Info-Zeile und Banner werden gefüllt), `go test` und `-race` sauber, Herstellername 0×.

## 63. Version 0.9.34 — Parallele PCs ohne gemeinsamen Ordner (über den Drucker)

Auf deinen Wunsch: kein gemeinsamer Ordner mehr. Jede Instanz läuft rein lokal.

**Wie es jetzt erkannt wird.** Der einzige gemeinsame Punkt ohne geteilte Dateien ist der Drucker selbst. Jede Instanz hört zusätzlich auf dem Befehlskanal des Druckers mit (`device/<serial>/request`) — nur lesend, es wird nichts Zusätzliches gesendet. Jeder Befehl, den DIESE Instanz schickt, trägt in seiner `sequence_id` eine zufällige Instanz-Kennung. Kommt auf dem Kanal ein Befehl an, dessen Kennung nicht die eigene ist, steuert gerade ein anderer Zugriff denselben Drucker (anderer PC, Bambu Studio, Handy-App). Dann erscheint oben das Band: „⚠ Anderer Zugriff auf Drucker erkannt: <Druckername>". Das passt genau zum „Kreuzen der Prozesse" — es meldet sich, wenn wirklich parallel gesteuert wird.

**Zugriffs-Log bleibt, aber lokal.** `zugriff.log` liegt jetzt im lokalen Datenverzeichnis (`%APPDATA%\Druckerfarm`), nicht mehr in einem geteilten Ordner — eine Zeile je Start/Stop mit Zeitstempel und Benutzer@PC.

Der gemeinsame Ordner und die zugehörige Einstellung sind entfallen.

Ehrlich zur Grenze: Die Erkennung greift nur, wenn der andere Zugriff auch tatsächlich einen Befehl sendet (nicht bloß zuschaut), und sie setzt voraus, dass der MQTT-Broker des Druckers Anfragen an mithörende Clients weiterreicht. Tut er das nicht, bleibt die Meldung still — ohne Schaden und ohne Fehlalarm. Sag mir, ob bei euch das Band erscheint, wenn ein zweiter PC an denselben Drucker etwas schickt; dann weiß ich, ob der Weg auf eurer Firmware trägt.

Geprüft: Go-Test (eigener Befehl mit Instanz-Kennung zählt nicht als fremd, fremder Befehl schon, alter Eintrag verfällt), jsdom (Band erscheint/verschwindet, keine Ordner-Einstellung mehr). `go test` und `-race` sauber, Herstellername 0×.

## 64. Version 0.9.35 — Ladekreis im Fullscreen

Im Fullscreen (beide Varianten) erscheint jetzt ein Ladekreis, solange das Bild noch nicht da ist — statt einer schwarzen Fläche. Er verschwindet, sobald das Video den ersten Frame liefert bzw. der erste Snapshot geladen ist. Das schwarze Warten kommt daher, dass ein Videostream/Snapshot erst nach dem ersten Bild etwas anzeigt; jetzt sieht man wenigstens, dass geladen wird.

Kleine Optimierung nebenbei: Die Live-Aktualisierung der Fullscreen-Kacheln (Rahmen/Balken) läuft nur noch, wenn tatsächlich eine Fullscreen-Ansicht offen ist, statt bei jedem Statustakt für jeden Drucker eine DOM-Abfrage zu machen.

Zur „langsamer geworden?"-Frage: Einen klaren Bremsklotz habe ich nicht gefunden — die schwarze Fläche im Fullscreen war fehlender Ladekreis, kein echtes Tempoproblem. Falls sich etwas anderes zäh anfühlt (welche Ansicht, wie viele Drucker?), sag es mir konkret, dann messe ich gezielt nach.

Geprüft: jsdom (Ladekreis in beiden Fullscreen-Zelltypen vorhanden, wird durch das Bild-/Video-Ereignis entfernt), `go test` und `-race` sauber, Herstellername 0×.

## 65. Version 0.9.36 — Fullscreen-Ladekreis: Bug behoben

Der Ladekreis im Fullscreen blieb schwarz stehen, auch wenn das Bild längst da war. Ursache: Beim Aktualisieren eines Snapshots setzt der Snapshot-Code den `onload`-Handler des Bildes neu — und überschrieb damit den Handler, der den Ladekreis entfernt. Jetzt macht der `onload` beides: alte Bild-URL freigeben UND den Ladekreis der Zelle ausblenden. Zusätzlich reagiert der Video-Weg jetzt auf zwei Ereignisse (`playing` und `loadeddata`), damit der Kreis auch bei Videostreams zuverlässig verschwindet. Der Hintergrund des Ladekreises ist außerdem halbtransparent statt voll schwarz — ein geladenes Bild scheint sofort durch.

Geprüft: Build ok, `go test` und `-race` unverändert grün (keine Go-Änderung), Herstellername 0×.

## 66. Version 0.9.37 — Mehr-PC-Erkennung über die SD-Karte des Druckers

Auf deinen Vorschlag umgestellt: Statt eines geteilten Ordners auf dem PC dient jetzt die SD-Karte EINES Druckers als gemeinsame Pinnwand. Dort liegt `multi-access-controll.json` mit einer kurzen Liste aktiver Sitzungen (Host, Benutzer, Zeitstempel).

Ablauf: Jede Instanz frischt in Abständen (etwa alle 90 Sekunden) auf genau einem erreichbaren Drucker ihren eigenen Eintrag auf — das ist dein „Cookie": ist er schon da mit deiner Kennung, ändert sich nichts. Dabei werden veraltete Einträge (älter als ein paar Minuten) verworfen. Trägt ein anderer PC einen frischen Eintrag, läuft er parallel — oben erscheint „⚠ Ein anderer PC läuft ebenfalls: benutzer@PC".

Bewusst nur EIN Drucker als Pinnwand (der erste erreichbare, mit automatischem Fallback), damit nicht auf allen 42 Geräten FTP-Verkehr entsteht. Alles best-effort: ist der Drucker offline oder FTP nicht möglich (kein Python), bleibt die Erkennung still — ohne Fehler.

Das bisherige MQTT-Mithören auf dem Befehlskanal (0.9.34) ist damit entfallen — das entlastet auch MQTT wieder. Das lokale `zugriff.log` (Start/Stop je PC) bleibt.

Geprüft: Go-Test (Zusammenführen der Sitzungen: veraltete raus, eigener Eintrag aufgefrischt, andere PCs korrekt erkannt; eigener Eintrag wird angelegt, wenn keiner da ist), `go test` und `-race` sauber, Herstellername 0×.

Bitte gegenchecken: Startest du das Tool auf zwei PCs, sollte nach spätestens ~1–2 Minuten auf beiden das Band erscheinen. Sag mir, ob die Datei auf der SD auftaucht und das Band greift.

## 67. Version 0.9.38 — „Gespeicherte Drucker" zeigt Online/Offline korrekt

In der Liste „Gespeicherte Drucker" (Tab „+ Drucker") stand überall „Offline", obwohl die Drucker in der Übersicht online waren. Grund: Die Online-Pille dort wurde nur einmal beim Zeichnen gesetzt — und das passiert beim Start, bevor der erste Status da ist. Anders als in der Übersicht wurde sie danach nicht aktualisiert.

Jetzt läuft die Pille mit: Bei jedem Statustakt (im Sekundenbereich) wird die Online/Offline-Anzeige in „Gespeicherte Drucker" aus derselben Statusliste nachgezogen wie die Übersicht. Im Bearbeiten-Modus einer Zeile wird sie nicht angefasst.

Geprüft: jsdom (Pille zeigt vor dem Status „Offline", nach dem Statusabruf „Online"), `go test` und `-race` sauber, Herstellername 0×.

## 68. Version 0.9.39 — Fortschritt bei Pause/Fehler & „In Reparatur"-Markierung

Zwei Punkte aus deiner Nachricht.

**Der Fortschrittsbalken verschwand bei Pause/Fehler.** Wenn ein Drucker wegen eines Fehlers pausierte, wurde der Balken auf 0 gesetzt und die Prozentzahl war weg — man sah nicht mehr, wie weit der Druck war. Jetzt bleibt bei Fehler/Pause der Balken samt Prozentzahl stehen (rot eingefärbt, mit „✗ 47 %"), sowohl in der Kachel als auch in der Zeile. Der Stand geht also nicht mehr verloren.

**Neue Markierung „In Reparatur" — auf dem Drucker hinterlegt.** Im Detailmenü eines Druckers gibt es jetzt einen Schalter „🔧 In Reparatur setzen" / „✓ Reparatur beenden". Beim Setzen kann optional eine kurze Notiz eingegeben werden. Ist die Markierung aktiv, erscheint an Name (Übersicht, Liste, Gespeicherte Drucker) eine orange Marke „🔧 In Reparatur".

Damit alle PCs denselben Stand sehen, wird die Markierung als JSON-Datei auf der SD-Karte des Druckers abgelegt: `maintenance.json` (Feld `in_repair`, optional `note`, `by`, `since`) — Dateiname in Englisch, wie gewünscht. Eine Hintergrundschleife gleicht in Abständen ab (SD ↔ lokal, der jeweils neuere Zeitstempel gewinnt), sodass eine an PC A gesetzte Reparatur nach kurzer Zeit auch auf PC B auftaucht.

Best-effort wie bei der Mehr-PC-Erkennung: Ist der Drucker offline oder FTP nicht möglich (kein Python), wird die Markierung sofort lokal gesetzt und später auf die SD geschrieben — ein Hinweis erscheint dann kurz. Kein Herstellername, keine Steuerbefehle an den Drucker; nur Lesen/Schreiben einer Datei auf der SD.

Geprüft: jsdom (Prozent bleibt bei Fehler sichtbar; Umschalter sendet an `/api/repair`; Marke erscheint nach dem Setzen), Go-Test (Pfad `maintenance.json`, Cache setzen/lesen/zurücksetzen), `go test` und `-race` sauber, Herstellername 0×.

Bitte gegenchecken: Setz an einem Drucker „In Reparatur", schau, ob `maintenance.json` auf der SD auftaucht, und ob die Marke nach kurzer Zeit auf einem zweiten PC erscheint.

## 69. Version 0.9.40 — Fortschritt im Fullscreen bei Pause/Fehler & „In Reparatur" in der Drucker-Liste

Zwei Nachträge.

**Fortschrittsbalken im Fullscreen blieb bei Unterbrechung leer.** In der Ansicht „Fullscreen (nur Video)" wurde der Balken bei einer Störung/Pause auf 0 gesetzt — man sah nur den roten Rahmen, aber nicht mehr, wie weit der Druck war. Jetzt bleibt der Balken auch bei Pause/Fehler auf dem tatsächlichen Prozentwert stehen (rot bei Fehler, gold bei Pause), passend zu Kachel und Zeile.

**„In Reparatur" jetzt auch in der Drucker-Liste umschaltbar.** Im Tab „+ Drucker" (Gespeicherte Drucker) hat jede Zeile jetzt einen 🔧-Knopf: einmal drücken markiert den Drucker als „in Reparatur" (Knopf wird grün, ✓🔧), erneut drücken beendet sie. Die Marke erscheint wie gewohnt am Namen und wird als `maintenance.json` auf der SD-Karte hinterlegt — also derselbe Stand auf allen PCs. Bisher ging das nur über das Detailmenü in der Übersicht.

Geprüft: jsdom (Fullscreen-Balken behält bei Fehler die Breite 63 %; 🔧-Knopf in der Liste sendet an `/api/repair`, Marke erscheint nach dem Setzen; Spaltenraster stimmt), `go test` und `-race` sauber, Herstellername 0×.

## 70. Version 0.9.41 — Kein Ladekreis bei Offline, „in repair"-Hinweis im Bild

Zwei Feinheiten in der Übersicht, im Kamerabereich der Kachel.

**Offline-Drucker zeigen keinen Ladekreis mehr.** Bisher versuchte die Kachel auch bei einem nicht erreichbaren Drucker, einen Videostream aufzubauen — der Ladekreis drehte sich dann endlos. Jetzt wird bei einem offline gemeldeten Drucker gar kein Stream mehr angestoßen; im Bildbereich steht ruhig „Offline" (📷). Sobald der Drucker wieder online ist, baut sich der Stream von selbst auf.

**In Reparatur: Hinweis statt Bild.** Ist ein Drucker als „in Reparatur" markiert, erscheint im Kamerabereich statt Ladekreis oder Stream ein zentrierter Hinweis „🔧 In Reparatur" (mit Warn-Schraffur). Auch hier wird kein Stream aufgebaut. Wird die Reparatur beendet und der Drucker ist online, kommt das Livebild von selbst zurück.

Technisch: Der Stream-Aufbau (WebRTC wie Snapshot) überspringt jetzt offline- und Reparatur-Drucker; die Kachel baut sich bei jedem Wechsel online↔offline bzw. Reparatur an↔aus neu auf, damit Hinweis und Bild ohne manuelles Neuladen umschalten.

Geprüft: jsdom (offline → kein Spinner/kein `conn`-Overlay; Reparatur → `cam-repair`-Hinweis, kein Spinner; online → normales Videoelement), `go test` und `-race` sauber, Herstellername 0×.

## 71. Version 0.9.42 — Ladekreis bei Offline endgültig weg, breitere Namensspalte

Nachbesserung zu 0.9.41.

**Der Ladekreis blieb bei Druckern, die während des Betriebs offline gingen.** Ursache: Wechselt ein Drucker von online auf offline, tauscht die Übersicht normalerweise nur die Statuszellen aus — der Kamerabereich blieb im alten Bild stehen, also mit dem ewig drehenden „Verbinde …"-Kreis. Jetzt prüft jede Statusaktualisierung, was im Kamerabereich gerade zu sehen ist, und baut die Kachel bei einer echten Abweichung neu: offline → „Offline" (📷), in Reparatur → „🔧 In Reparatur", online → Livebild. Kein manuelles Neuladen nötig, und im Normalbetrieb flackert nichts (nur bei echtem Wechsel wird neu gezeichnet). Für offline und in Reparatur wird weiterhin gar kein Stream mehr aufgebaut.

**Namensspalte in der Listenansicht breiter.** In der Druckerübersicht (Liste) war die Spalte „Name" mit fester Breite oft zu schmal. Sie ist jetzt flexibel und deutlich breiter (mind. 160 px, wächst mit dem Fenster), sodass längere Druckernamen nicht mehr abgeschnitten werden.

Geprüft: jsdom (online→offline: Ladekreis/Video weg, „Offline" erscheint; offline→online: Livebild kommt zurück; Wechsel in Reparatur: Hinweis statt Bild), `go test` und `-race` sauber, Herstellername 0×.

## 72. Version 0.9.43 — Kamera-Hinweise als Text statt Icon

Im Kamerabereich der Übersicht stand bei einem nicht erreichbaren oder in Reparatur befindlichen Drucker ein Symbol (📷 bzw. 🔧) über dem Hinweis. Auf Wunsch sind die Symbole raus — dort steht jetzt nur noch der Text: „Offline" bzw. „In Reparatur" (etwas größer und mittig).

Geprüft: jsdom (offline/Reparatur zeigen reinen Text ohne Symbol, kein Ladekreis), `go test` und `-race` sauber, Herstellername 0×.

## 73. Version 0.9.44 — Ladekreis läuft nicht mehr endlos, Text statt Kreis

Der eigentliche Grund, warum der Ladekreis blieb: Ein ausgeschalteter Drucker gilt über MQTT noch ~30–45 s als „online" (Keepalive), und ein Drucker, dessen Kamera nicht erreichbar ist (Kamera aus, falsche Adresse), meldet sich per MQTT als online, liefert aber nie ein Bild. In beiden Fällen drehte sich der „Verbinde …"-Kreis endlos.

Jetzt gibt es ein Zeitlimit: Kommt binnen 10 Sekunden kein Bild, verschwindet der Ladekreis und im Kamerabereich steht Text — „Offline", wenn der Drucker per MQTT nicht mehr erreichbar ist, sonst „Kein Bild". Sobald doch ein Bild kommt, ist der Kreis sofort weg (unverändert). Der Wechsel online↔offline↔Reparatur baut die Kachel weiterhin passend um.

Geprüft: jsdom (Ladekreis wird durch Text ersetzt — „Kein Bild" bei erreichbarem, „Offline" bei nicht erreichbarem Drucker; Selbstheilung online/offline/Reparatur weiterhin ok), `go test` und `-race` sauber, Herstellername 0×.

## 74. Version 0.9.45 — Kamera-Bereich strikt aus dem bekannten Zustand (kein Ladekreis bei Offline)

Du hattest recht: Die App weiß, welcher Drucker offline ist (die Online/Offline-Pille stimmt) — also muss das der Kamerabereich auch ohne Umweg über einen Videoversuch zeigen.

Der eigentliche Fehler: Beim Start des Kameraservers (go2rtc) wurde der „startet …"-Ladekreis für ALLE Kacheln gezeigt — dieser Zweig kam VOR der Offline-Prüfung. Ein längst offline gemeldeter Drucker bekam also trotzdem den Kreis, und weil sich sein bekannter Zustand (offline) nicht mehr änderte, blieb der Kreis stehen.

Jetzt ist die Reihenfolge klar: „in Reparatur" und „offline" hängen NICHT vom Kameraserver ab und werden zuerst geprüft. Ein offline gemeldeter Drucker zeigt sofort „Offline", ein Drucker in Reparatur „In Reparatur" — unabhängig davon, ob go2rtc noch startet. Nur wirklich erreichbare Drucker bauen einen Stream auf.

Zusätzlich merkt sich jede Kachel ihren Kamera-Zustand (online/offline/Reparatur). Ändert er sich, wird die Kachel bei der nächsten Statusmeldung neu gebaut und ein evtl. laufender Ladekreis/Stream gestoppt — kein Neuladen nötig, kein Flackern im Normalbetrieb. Das 10-Sekunden-Zeitlimit von 0.9.44 bleibt als Fangnetz für den Sonderfall „online, aber Kamera liefert kein Bild".

Geprüft: jsdom (offline zeigt „Offline" auch während der go2rtc-Startphase und bleibt Text; Reparatur zeigt Text; Wechsel offline→online startet den Stream wieder), `go test` und `-race` sauber, Herstellername 0×.

## 75. Version 0.9.46 — Auch im Vollbild: Text statt Ladekreis, „Offline" wie „In Reparatur"

Die Übersicht war korrigiert, das Vollbild aber nicht — dort drehte sich für offline/in Reparatur befindliche Drucker weiter der Ladekreis. Die normale Vollbild-Kachelansicht nutzte bereits die reparierte Kachel; die reine Video-Ansicht („Fullscreen Video Only") hatte jedoch eine eigene Zelle mit eigenem Ladekreis.

Jetzt zeigen beide Vollbild-Varianten denselben Text: „Offline" bzw. „In Reparatur", zentriert und groß — kein Ladekreis, kein Stream-Versuch. Wechselt ein Drucker im Vollbild den Zustand (online↔offline↔Reparatur), wird die Zelle sofort umgebaut und ein laufender Ladekreis/Stream gestoppt.

Außerdem: In der Übersicht steht „Offline" jetzt in derselben Schrift wie „In Reparatur" (groß, mittig, Großbuchstaben) — nur in neutralem Grau statt der Warnfarbe.

Geprüft: jsdom (Vollbild-Zelle zeigt bei offline/Reparatur Text ohne Ladekreis; Wechsel online→offline baut die Zelle im Vollbild um; erreichbarer Drucker zeigt weiterhin Bild), `go test` und `-race` sauber, Herstellername 0×.

## 76. Version 0.9.47 — Nach dem Bearbeiten erreichbar prüfen, manueller Ping, breitere Namensspalte

Drei Punkte in der Drucker-Liste („+ Drucker" / Gespeicherte Drucker).

**Nach dem Bearbeiten wird sofort geprüft, ob der Drucker erreichbar ist.** Speicherst du eine Änderung (z. B. neue IP), klopft das Programm direkt an die Ports des Druckers (MQTT/FTP/Kamera) und meldet „Online" bzw. „Offline" samt kurzer Diagnose — ohne auf den nächsten Statustakt zu warten. Zusätzlich wird die Online-Pille in den folgenden Sekunden mehrmals nachgezogen, weil der Server die MQTT-Verbindung nach einer Adress-/Code-Änderung neu aufbaut.

**Manueller Ping.** Ein Klick auf die Status-Pille („Online/Offline") in der Liste prüft die Erreichbarkeit jederzeit auf Knopfdruck — dieselbe Diagnose wie oben (Gerät aus / im Netz, Kamera an/aus).

**Namensspalte breiter.** In der Drucker-Liste war die Spalte „Name" zu schmal. Sie ist jetzt deutlich breiter (mind. 200 px, wächst mit dem Fenster); Zugangscode- und Seriennummer-Spalte etwas schmaler.

Geprüft: jsdom (Status-Pille ist klickbar und ruft `/api/reach`; Ping wertet den MQTT-Port aus; Namensspalte breiter), `go test` und `-race` sauber, Herstellername 0×.

## 77. Version 0.9.48 — Buttons „Bearbeiten"/„Löschen" laufen nicht mehr über den Rand

Nach der breiteren Namensspalte (0.9.47) wurden die Aktions-Buttons in der Drucker-Liste in enge Rasterspalten gequetscht — der Text stand teils über dem Button-Rand.

Behoben: Die drei Aktions-Schaltflächen (🔧, Bearbeiten, Löschen) sitzen jetzt gemeinsam in einer Zelle (flexibel nebeneinander) statt in drei starren Rasterspalten. Dadurch behält jeder Button seine volle Breite, und der Text bleibt sauber im Rahmen. Das Raster hat dadurch acht statt zehn Spalten — Kopf, Zeile und Bearbeiten-Modus wurden entsprechend angepasst. Die Namensspalte bleibt breit (mind. 160 px, wächst mit).

Geprüft: jsdom (Kopf und Zeile haben je 8 Spalten; die Aktionszelle enthält im Normalmodus 3, im Bearbeiten-Modus 2 Buttons), `go test` und `-race` sauber, Herstellername 0×.

## 78. Version 0.9.49 — Drucker-Liste breiter & ausgerichtet, Ping erklärt „gefunden, aber offline"

**Tabelle breiter, Überschriften wieder passend.** Zwei Probleme: Die „+ Drucker"-Ansicht war auf 860 px begrenzt (zu schmal), und Kopfzeile und Datenzeilen sind je ein eigenes CSS-Raster — mit flexiblen Spalten richteten sie sich unterschiedlich aus, sodass die IP verdeckt wurde und die Überschriften verrutschten. Jetzt ist die Ansicht bis 1200 px breit, die Spalten haben feste Breiten (nur „Name" wächst mit), sodass Kopf und Zeilen exakt übereinanderliegen. Ist das Fenster sehr schmal, lässt sich die Tabelle waagerecht scrollen statt zu überlappen.

**Ping/Erreichbarkeit — Ursache für „gefunden, aber offline".** Der Statuskanal (MQTT) braucht die **Seriennummer** des Druckers; ohne sie baut das Programm gar keine Statusverbindung auf — der Drucker steht dann dauerhaft auf „Offline", auch wenn er im Netz ist. Der Ping (Klick auf die Status-Pille oder automatisch nach dem Bearbeiten) sagt das jetzt klar:

- „Online" — im Netz und Statusport offen.
- „Im Netz erreichbar — Statusverbindung wird aufgebaut …" — Gerät antwortet, MQTT verbindet neu.
- „Im Netz erreichbar, aber ohne Seriennummer kein Status möglich — Seriennummer eintragen" — genau der Fall „gefunden, aber offline".
- „Offline — Gerät aus oder nicht im Netz" — keine Antwort auf irgendeinem Port.

Zusätzlich stößt der Ping bei erreichbaren Druckern eine frische MQTT-Verbindung an (neuer Endpunkt `/api/reconnect`), damit eine hängengebliebene Verbindung nicht fälschlich „offline" zeigt. „Erreichbar" gilt jetzt, sobald das Gerät auf irgendeinem seiner Ports (Kamera/Dateien/Status) antwortet — nicht nur am Statusport.

Geprüft: jsdom (Kopf/Zeile je 8 Spalten; Ping-Meldungen für online / im Netz / ohne Seriennummer / offline; Reconnect nur bei erreichbarem Gerät), Go-Test (Reconnect-Handler), `go test` und `-race` sauber, Herstellername 0×.

## 79. Version 0.9.50 — „In Reparatur" optisch wie „Offline"

Auf Wunsch: Die Marke „In Reparatur" im Kamerabereich hat jetzt dieselbe Schrift und Farbe wie „Offline" (neutrales Grau, gleiche Größe/Großbuchstaben) und keinen eigenen Hintergrund mehr — die orange Warn-Schraffur ist weg. Gilt für die Übersicht und beide Vollbild-Ansichten. Unterschieden werden die beiden Zustände nur noch am Text.

Geprüft: jsdom (cam-repair und fsv-hinweis.repair sind transparent und nutzen die Textfarbe wie Offline), `go test` und `-race` sauber, Herstellername 0×.

## 80. Version 0.9.51 — Standalone-Kameras (z. B. Xiaomi CW300) als eigene Kacheln

Neben den Druckerkameras lassen sich jetzt eigenständige Kameras einbinden, die keinen Drucker darstellen (z. B. eine Outdoor-Kamera im Büronetz). Sie laufen über dieselbe go2rtc-Mechanik (WebRTC + Snapshot) wie die Druckerbilder, tauchen aber in nichts Druckerbezogenem auf.

Was neu ist:

- **Konfiguration statt Hardcoding.** Kameras liegen in der App-Config unter `cameras` (ID, Name, Streamname, Quelle inkl. Zugangsdaten). Beliebig erweiterbar, ohne Code anzufassen.
- **go2rtc-Eintrag schreibt das Tool.** Die Quelle (z. B. `xiaomi://…`) wird beim Speichern automatisch in die `go2rtc.yaml` geschrieben und übersteht jedes Neuschreiben der Datei — der Anwender fasst go2rtc nicht an. Die Zugangsdaten werden nicht ins Log geschrieben; in der Liste erscheint das Passwort maskiert.
- **Eigene Kachel.** Gleiche Größe/Theming wie Druckerkacheln, aber nur Name + Livebild — kein Fortschritt, keine Temperatur, kein AMS, keine Statusfarbe. Snapshot-Knopf funktioniert unverändert (`/api/snapshot/<id>`). Standalone-Kameras stehen immer ans Ende, nach den Druckern, und bleiben auch beim Filter „Online" sichtbar.
- **Getrennt vom Drucker-Teil.** Keine SSDP-Discovery, keine MQTT-Verbindung, kein FTP-Sync für Kameras. Die Druckerzählung, der Discovery-Lauf und die MQTT-Verbindungen bleiben unverändert.
- **Verwaltung im Tool.** Im Tab „+ Drucker" gibt es einen Abschnitt „Kameras": Name + Quelle eintragen → hinzufügen; Löschen entfernt die Kachel wieder (laufender Stream wird gestoppt), ohne Fehler.
- **Ausfallverhalten.** Reißt die Verbindung ab (Kamera aus, Internet für den Xiaomi-Schlüssel weg), zeigt die Kachel „Kamera nicht verfügbar" und versucht es weiter — die restliche Oberfläche läuft normal weiter.
- **Vollbild.** Beide Vollbild-Ansichten zeigen die Kamera als reines Bild (ohne Fortschrittsbalken/Statusfarbe), ans Ende sortiert.

Für die CW300 ist die vorgesehene Quelle:
`xiaomi://6782331241:de@192.168.189.189?did=1109718403&model=mxiang.camera.moc006`
Hinweis: Feste IP — bei der Kamera eine DHCP-Reservierung setzen, sonst ist der Stream nach einem Lease-Wechsel tot (eine Auto-Discovery gibt es für die CW300 nicht). Bei Bandbreitenproblemen kann `&subtype=sd` an die Quelle gehängt werden (Substream) — erst messen, nicht vorsorglich.

Ergänzt: `.gitignore` schließt `config.json`, `go2rtc.yaml`, Logs und Binärdateien aus (keine Zugangsdaten im Repo).

Geprüft: Go-Tests (go2rtc.yaml enthält Drucker- UND Kamera-Stream; `/api/cameras` POST/GET/DELETE; Snapshot löst Kamera-ID zum Streamnamen auf; Druckerzählung unverändert), jsdom (Kamera-Kachel ohne Drucker-Overlays, steht am Ende, aus der Druckerzählung raus, beim Online-Filter sichtbar, Quelle maskiert, Hinzufügen sendet an die API), JS-Syntaxcheck, `go test` und `-race` sauber, Herstellername 0×.

## 81. Version 0.9.52 — Kamera-Snapshot sauber getrennt, kein Ladekreis bei 503

Zwei Nachbesserungen zur Kamera-Einbindung.

**„kein Drucker zum Stream …" bei einer Kamera.** Die Snapshot-Diagnose (die erklärt, warum kein Bild kommt) war komplett druckerbezogen: sie suchte zum Streamnamen einen Drucker, prüfte dessen MQTT-Erreichbarkeit, Port 322 und den Zugangscode. Für eine Standalone-Kamera (z. B. „dowell3d") gibt es keinen Drucker — deshalb die irreführende Meldung. Jetzt ist der Kamerafall sauber abgetrennt (in `camera.go`): Erkennt die Diagnose einen Kamera-Stream, gibt sie eine kameraspezifische Meldung aus („Kamera … liefert gerade kein Bild — erreichbar? Beim Verbindungsaufbau braucht go2rtc kurz Internet für den Xiaomi-Sitzungsschlüssel.") und fasst keine Drucker-Prüfungen an.

**Kein Ladekreis bei 503.** Liefert der Snapshot einer Kamera einen Fehler (503 Service Unavailable, z. B. Kamera aus oder Schlüssel-Abruf gescheitert), dreht sich jetzt kein Ladekreis mehr endlos: Der Mini-Ring wird ausgeblendet und es steht ruhig „Kamera nicht verfügbar". Kommt wieder ein Bild, erscheint der Ring von selbst zurück. Für Drucker bleibt die bisherige Fehleranzeige unverändert.

Geprüft: Go-Test (Kamera-Stream wird erkannt, Diagnose nennt den Kameranamen und nicht „Drucker"), jsdom (Kamera-Snapshotfehler blendet den Ring aus und zeigt „nicht verfügbar"; Drucker-Fehler unverändert), `go test` und `-race` sauber, Herstellername 0×.

## 82. Version 0.9.53 — Kamera-Snapshot ohne 503-Konsolenfehler

Der Browser protokolliert jede fehlgeschlagene Anfrage (`503 Service Unavailable`) als roten Fehler in der Konsole. Bei einer Außenkamera, die zeitweise kein Bild liefert, war das nur Lärm.

Jetzt antwortet der Server für Standalone-Kameras nicht mehr mit 503, sondern mit `200` und einem Hinweis-Header (`X-Snapshot-Unavailable`). Die Oberfläche zeigt weiterhin „nicht verfügbar" ohne Ladekreis und hält dieselbe Pause bis zum nächsten Versuch ein — aber in der Konsole erscheint kein Fehler mehr. Für Drucker bleibt das Verhalten (inkl. 503) unverändert.

Geprüft: Go-Test (Kamera-Snapshot liefert 200 statt 503 und setzt den Hinweis-Header), `go test` und `-race` sauber, JS-Syntaxcheck, Herstellername 0×.

## 83. Version 0.9.54 — Xiaomi-Login in go2rtc bleibt erhalten, echter Fehler wird angezeigt

„Camera unavailable" heißt: go2rtc bekommt kein Bild von der Kamera. Zwei Sachen dazu, die in unserer Hand liegen:

**Der Mi-Home-Login in go2rtc wird beim Neuschreiben nicht mehr überschrieben.** Das Tool erzeugt die `go2rtc.yaml` bei jeder Änderung/Neustart neu (Abschnitte `api` und `streams`). go2rtc legt aber beim Mi-Home-Login eigene Daten (Konto/Token für den Xiaomi-Cloud-Schlüssel) in dieselbe Datei. Bisher gingen die beim Neuschreiben verloren — dann kann go2rtc den Kameraschlüssel nicht mehr holen und liefert kein Bild. Jetzt werden alle fremden Top-Level-Abschnitte aus der bestehenden Datei übernommen, nur `api`/`streams` werden neu geschrieben. Ein einmal hinterlegter Xiaomi-Login bleibt damit bestehen.

**Die Kachel zeigt jetzt den echten Grund.** Statt nur „nicht verfügbar" fragt das Tool go2rtc nach dem konkreten Fehler des Streams und zeigt ihn an (z. B. ein Login-/Schlüssel-Fehler von Xiaomi). So ist erkennbar, ob die Kamera aus ist, das Internet für den Schlüssel fehlt oder der Mi-Home-Login in dieser go2rtc-Instanz nicht hinterlegt ist.

Wichtig für die Inbetriebnahme: Der Mi-Home-Login muss in genau der go2rtc-Instanz vorliegen, die dieses Tool startet (Daten liegen in `%APPDATA%\Druckerfarm\go2rtc.yaml`). Ist er dort noch nicht hinterlegt, kann go2rtc den Session-Key nicht holen — dann bleibt die Kachel „nicht verfügbar", egal ob die Stream-URL stimmt.

Geprüft: Go-Test (fremde Abschnitte wie ein `xiaomi:`-Konto bleiben erhalten, `api`/`streams` werden ersetzt; Zusammenbau stimmt), `go test` und `-race` sauber, Herstellername 0×.

## 84. Version 0.9.55 — Xiaomi Mi-Home-Login im Tool (über go2rtc)

Der Log der Kamera zeigte `401 Unauthorized` — kein Netz-/IP-Problem, sondern ein fehlender/abgelaufener Xiaomi-Login. Xiaomi-Kameras brauchen einen Mi-Home-Login: go2rtc holt damit bei jedem Verbindungsaufbau den Sitzungsschlüssel aus der Cloud. Der Token läuft nach ~3 Tagen ab (bekannte go2rtc-Eigenheit), dann kommt wieder 401.

Neu: Im Tab „+ Drucker" → „Kameras" gibt es jetzt einen aufklappbaren Abschnitt „Xiaomi Mi-Home-Login". Dort meldet man sich mit Mi-Konto, Passwort und Region an — **ohne go2rtc selbst zu öffnen**. Damit Xiaomis Anmeldeprotokoll (Verschlüsselung, Captcha, 2FA) nicht neu gebaut werden muss, nutzt das Tool go2rtcs eigene, bereits vorhandene Anmeldung: Die Eingaben werden an die lokale go2rtc-Instanz (`/api/xiaomi`) weitergereicht, die den eigentlichen Login macht und den Token speichert. Zusammen mit dem Erhalt fremder Config-Abschnitte (0.9.54) bleibt der Login bestehen.

Datenschutz: Zugangsdaten laufen ausschließlich an das lokale go2rtc; sie werden nicht gespeichert und nicht geloggt, stehen nie in einer URL (nur im Anfrage-Body) und das Passwortfeld wird nach dem Absenden geleert. Die Antwort von go2rtc (Erfolg, oder nötiger Bestätigungscode/Captcha) wird im Abschnitt angezeigt.

Ehrliche Einordnung: Diesen Login konnte ich hier nicht gegen echte Xiaomi-Server testen, und go2rtc 1.9.14 hat beim Xiaomi-Login teils eigene Probleme (Captcha/„need_verify"). Bitte auf dem Zielrechner ausprobieren; falls go2rtc einen Bestätigungsschritt verlangt, erscheint der Hinweis im Ergebnisfeld — dann bitte melden, dann bauen wir den Folgeschritt (Code-Eingabe) und die automatische Erneuerung vor Ablauf der ~3 Tage.

Geprüft: Go-Test (Login-Proxy reicht den Body an go2rtcs `/api/xiaomi` weiter; Zugangsdaten landen NICHT in der URL/Query), jsdom (Login sendet die Daten im Body, nie in der URL; Passwortfeld wird geleert; Ergebnis wird angezeigt), `go test` und `-race` sauber, JS-Syntaxcheck, Herstellername 0×.

## 85. Version 0.9.56 — Xiaomi-Login korrekt (Formulardaten, Captcha/2FA, Kameras laden)

Der Login aus 0.9.55 scheiterte mit „HTTP 400 wrong request". Grund (aus dem go2rtc-Quelltext geprüft): go2rtcs `/api/xiaomi` erwartet die Anmeldung als **Formulardaten** (`application/x-www-form-urlencoded`), nicht als JSON. Das ist jetzt korrigiert und der komplette Ablauf abgebildet:

- **Anmelden** mit Mi-Konto + Passwort (Formulardaten, nur im Body — nie in der URL).
- **Captcha:** Verlangt Xiaomi ein Captcha, zeigt das Tool das Captcha-Bild direkt an; man tippt die Zeichen ein und bestätigt.
- **Bestätigungscode (2FA):** Schickt Xiaomi einen Code per SMS/E-Mail, zeigt das Tool an, wohin (maskiert), und man gibt den Code ein.
- **Kameras laden:** Nach erfolgreicher Anmeldung listet das Tool die Kameras des Kontos (über go2rtc, `GET ?id=&region=`) mit fertigen Stream-URLs. Ein Klick auf „Hinzufügen" legt die Kamera als Kachel an — die `xiaomi://…`-URL muss man nicht mehr von Hand eintippen.

Der eigentliche Login macht weiterhin go2rtc (Verschlüsselung/Token); das Tool reicht nur die Eingaben an die lokale Instanz weiter und zeigt deren Antworten. Zugangsdaten werden nicht gespeichert, nicht geloggt, stehen nie in einer URL; Passwort- und Code-Felder werden nach dem Absenden geleert. Der von go2rtc gespeicherte Token bleibt dank 0.9.54 über Config-Neuschreibungen erhalten (läuft xiaomi-seitig nach ~3 Tagen ab → dann erneut anmelden).

Geprüft: Go-Test (Login-Proxy reicht Body an `/api/xiaomi`, Zugangsdaten nicht in der URL), jsdom (Login sendet Formulardaten mit Konto/Passwort; Captcha-Bild wird angezeigt; Captcha→Verify→Erfolg; danach Kameraliste; „Hinzufügen" legt Kamera mit der xiaomi://-URL an; Passwort/Code werden geleert), `go test` und `-race` sauber, JS-Syntaxcheck, Herstellername 0×.

## 86. Version 0.9.57 — Kameras über alle Xiaomi-Regionen finden

Nach erfolgreicher Anmeldung meldete das Tool „Keine Kameras im Konto gefunden". Grund: go2rtc listet Kameras pro Mi-Region und antwortet für eine Region ohne Kameras mit 404. Die CW300 hängt aber offenbar nicht in „de".

Jetzt probiert das „Kameras des Kontos laden" die gewählte Region zuerst und danach automatisch die übrigen bekannten Regionen (de, us, cn, i2, ru, sg und global) durch, sammelt alle gefundenen Kameras ein (ohne Dubletten) und zeigt hinter jeder Kamera die Region an. Die Kamera wird damit gefunden, egal in welcher Region das Konto sie führt; die eingetragene Region bleibt der erste Versuch.

Geprüft: jsdom (Kamera wird gefunden, obwohl „de" gewählt war und sie in „us" liegt; „de" wird zuerst probiert, „us" danach; jede Kamera nur einmal), `go test` sauber, JS-Syntaxcheck, Herstellername 0×.

## 87. Version 0.9.58 — X2D: Kamera-Schema wie H2/P2 (rtsps zuerst)

Zur Meldung, dass bei den X2D-Druckern Streams und Uploads nicht laufen. Zwei getrennte Dinge:

- **Kamera (Stream):** X2D gehört zur selben Generation wie H2/P2 und antwortet auf `rtspx://` mit einer Umleitung — deshalb muss `rtsps://` zuerst versucht werden. Bisher fiel X2D in den Standardfall (rtspx zuerst). Jetzt steht X2 (wie H2/P2/H1) auf „rtsps zuerst". Das hilft aber nur, wenn der Kamera-Port des Druckers überhaupt antwortet (siehe unten).

- **Upload (Port 990):** Der Datei-Upload hängt NICHT am Modell — er spricht bei allen Druckern gleich FTPS auf Port 990 mit Zugangscode. „Zeitüberschreitung – Drucker antwortet nicht auf Port 990" heißt: der Drucker nimmt auf Port 990 keine Verbindung an. Das ist eine Geräte-/Netzsache, kein Fehler im Tool: bei den H2-/X2-Geräten müssen „LAN Mode Liveview" und der lokale Dateizugriff am Drucker aktiv sein, und Port 322 (Kamera) bzw. 990 (Dateien) dürfen nicht durch Firewall/VLAN blockiert sein. Dass MQTT (Port 8883, Status/Fortschritt) geht, heißt nur, dass der Drucker im Netz ist — nicht, dass 322/990 offen sind.

So lässt es sich eingrenzen: In „+ Drucker" auf die Status-Pille eines woobly-Druckers klicken (Ping). Das klopft an 322/990/8883 an und sagt, ob der Port „abgelehnt" (Dienst aus → am Drucker einschalten) oder „keine Antwort" (Firewall/VLAN/Modus) ist.

Geprüft: Go-Test (X2D liefert rtsps:// zuerst), `go test` und `-race` sauber, Herstellername 0×.

## 88. Version 0.9.59 — X2D: Video-Umschalten schaltet die Kamera nicht mehr ab

Bei den X2D-Druckern ging die Kamera „von selbst wieder auf aus", sobald man den Video-Modus (720p/1080p) wählte. Ursache: Das Umschalten schickte an ALLE Drucker den alten Befehl `ipcam_resolution_set` (720p/1080p). Die neue Generation (H2/X2/P2) kennt diesen Befehl nicht — bekommt ein X2D ihn trotzdem, schaltet er seine Kamera/Liveview ab.

Jetzt wird `ipcam_resolution_set` nur noch an Modelle geschickt, die es unterstützen (X1/P1/A1). Für H2/X2/P2 wird der Befehl übersprungen — der Livestream läuft davon unabhängig weiter, die Kamera bleibt an. (Diese Modelle haben eine andere Kamera ohne dieses 720p/1080p-Umschalten.)

Geprüft: Go-Test (H2D/X2D/P2S werden als neue Generation erkannt und übersprungen; X1/P1/A1 nicht), `go test` und `-race` sauber, Herstellername 0×.

## 89. Version 0.9.60 — X2D-Kamera-Reihenfolge zurückgenommen (Rückkehr zum funktionierenden Stand)

Rückmeldung: „Die X2D sind gar nicht erreichbar, in einer Version vorher ging das." Nachgeprüft mit einem Diff gegen 0.9.38: Die MQTT-/Erreichbarkeits-Logik ist unverändert. Die einzige X2D-spezifische Änderung war 0.9.58 (Kamera-Reihenfolge auf „rtsps zuerst"). Diese Vermutung ist zurückgenommen: X2D nutzt wieder die Standard-Reihenfolge (rtspx:// zuerst). Der 0.9.59-Fix (kein ipcam_resolution_set an H2/X2/P2) bleibt.

Geprüft: go test sauber, Herstellername 0×.

## 90. Version 0.9.61 — Status-Pille auch in der Übersicht anklickbar (Ping)

Rückmeldung: die Status-Pille sei nicht auffindbar. Zur Klarstellung: Die Online/Offline-Pille ist unverändert vorhanden — in den Kacheln der Übersicht und in der „+ Drucker"-Liste (dort seit 0.9.47 anklickbar für den Ping). Sie wurde nicht entfernt.

Damit der Ping leichter zu finden ist, ist die Pille jetzt AUCH in der Hauptübersicht (Kachel und Liste) anklickbar: ein Klick auf „Online/Offline" prüft die Erreichbarkeit des Druckers (Ports 322/990/8883) und zeigt das Urteil an — dieselbe Prüfung wie in der „+ Drucker"-Liste.

Geprüft: jsdom (Pille in Übersicht und „+ Drucker"-Liste vorhanden; in der Übersicht jetzt anklickbar und ruft den Ping), `go test` sauber, Herstellername 0×.

## 91. Version 0.9.62 — Ping-Ergebnis bleibt stehen (Fenster statt kurzer Meldung)

Das Ping-Ergebnis verschwand als kurze Meldung, bevor man es lesen konnte. Jetzt öffnet ein Klick auf die Status-Pille ein Fenster, das offen bleibt, bis man es schließt. Es listet die drei Ports einzeln auf: Kamera (322), Dateien (990), Status (8883) — jeweils „offen" (grün) oder mit Grund („keine Antwort" / „abgelehnt", rot) — plus das Gesamturteil. Genau das brauchen wir, um beim X2D zu sehen, was Port 322 sagt.

Geprüft: jsdom (Fenster erscheint, listet 322/990/8883 mit Status, bleibt offen), `go test` sauber, Herstellername 0×.

## 92. Version 0.9.63 — X2D-Kamera: rtsps zuerst (jetzt korrekt begründet)

Der Ping zeigt: Bei woobly 28 sind alle drei Ports offen — Kamera (322), Dateien (990), Status (8883). Die Kamera ist also erreichbar; es fehlte nur die richtige RTSP-Verhandlung.

Damit ist klar, dass meine Rücknahme in 0.9.60 falsch war: Ich hatte X2D auf `rtspx://` zuerst zurückgestellt, weil „nicht erreichbar" gemeldet wurde — Ursache war aber die abgeschaltete Kamera, nicht die URL. Der X2D gehört zur H2-Generation, die auf `rtspx` mit einer Umleitung antwortet und kein Bild liefert; sie braucht `rtsps://` zuerst. Beim Snapshot blieb go2rtc auf `rtspx` hängen (i/o-Timeout) und kam nicht mehr zu `rtsps`.

X2 steht jetzt wieder (wie H2/P2/H1) auf „rtsps zuerst". Nach dem Start schreibt das Tool die go2rtc.yaml neu und startet go2rtc — bei eingeschalteter Kamera sollte das Bild jetzt kommen.

Geprüft: Go-Test (X2D liefert rtsps:// zuerst), `go test` sauber, Herstellername 0×.

## 93. Version 0.9.64 — X2D: Standard-Kameraschema zurück + Hinweis auf den Kamera-Test

Korrektur: Die H2D-Drucker funktionieren mit Kamera — also ist die „H2-Generation kann kein RTSP"-These falsch. H2D und X2D verhalten sich bei der Kamera unterschiedlich: H2D braucht `rtsps://` zuerst (funktioniert), der X2D lief früher mit `rtspx://` zuerst (Standard). In 0.9.63 hatte ich X2D fälschlich zu H2D gruppiert (rtsps zuerst) — dadurch hängt go2rtc auf `rtsps` und läuft in den Timeout.

Zurückgenommen: X2D nutzt wieder die Standard-Reihenfolge (`rtspx://` zuerst), wie in der Version, in der es lief. H2D bleibt auf `rtsps` zuerst.

Der zuverlässige Weg pro Drucker ist ohnehin der eingebaute Kamera-Test: Im Detailmenü eines Druckers „🔍 Kamera-Adresse testen" — das probiert alle Formen (rtsps/rtspx, live/1 und live/0) durch, merkt sich die, die ein Bild liefert, und nimmt sie künftig zuerst. Damit ist das Modell-Raten hinfällig.

Geprüft: `go test` sauber, Herstellername 0×.

## 94. Version 0.9.65 — Auto-Reconnect, N6→X2D, Modal statt Dropdown im Grid

**Drucker kommt wieder online → MQTT verbindet sich neu.** Bisher galt: War ein Drucker beim ersten Verbindungsversuch offline, gab die MQTT-Bibliothek auf und versuchte es nie wieder — der Drucker blieb „offline" und die Steuerung meldete „braucht eine MQTT-Verbindung", obwohl er längst wieder am Netz war. (Die automatische Wiederverbindung griff nur nach einer einmal geglückten Verbindung.) Jetzt versucht das Tool die Erstverbindung dauerhaft erneut (alle 20 s), bis der Drucker erreichbar ist. Kommt ein Drucker zurück, verbindet er sich von selbst und wird wieder online/steuerbar — ohne Neustart.

**Modellcode N6 wird als X2D erkannt.** Beim Hinzufügen stand „N6" statt „X2D". Die Kennung N6 ist jetzt dem Modell X2D zugeordnet.

**Grid: „mehr" öffnet eine Modal-Box statt Dropdown.** In der Kachel-Ansicht (Grid) öffnet ein Klick auf „mehr ▾" jetzt eine zentrierte Modal-Box mit abgedunkeltem Hintergrund (Klick daneben schließt) — statt des kleinen Dropdown-Menüs. In den anderen Ansichten bleibt es beim bisherigen Verhalten. Inhalt (Steuerung, Temperaturen, AMS, Reparatur) ist identisch und aktualisiert sich weiter live.

**X2D-Video:** weiterhin offen — dazu brauche ich das Ergebnis des Kamera-Tests am X2D (Detailmenü → „🔍 Kamera-Adresse testen").

Geprüft: Go-Test (N6→X2D; MQTT baut mit ConnectRetry), jsdom (Grid „mehr" öffnet Modal + Backdrop, schließt sauber), `go test` und `-race` sauber, Herstellername 0×.

## 95. Version 0.9.66 — Zuverlässiger MQTT-Reconnect (eigener Takt statt paho-Automatik)

Die Umstellung auf paho-ConnectRetry (0.9.65) hat die offline gebliebenen Drucker nicht zurückgeholt. Deshalb jetzt ein eigener, deterministischer Reconnect-Takt: Alle 30 Sekunden prüft das Tool für jeden Drucker mit Seriennummer, ob eine lebende MQTT-Verbindung besteht. Wenn nicht, wird die alte (tote/hängende) Verbindung sauber getrennt und frisch aufgebaut. Kommt ein Drucker wieder ans Netz, ist er damit spätestens nach ~30 s von selbst wieder online und steuerbar — ohne Programm-Neustart. Der Takt pausiert, während „Netzwerkverkehr anhalten" aktiv ist.

Geprüft: Go-Test (ein Reconnect-Durchlauf überspringt Drucker ohne Seriennummer und legt keine Verbindung an), `go test` und `-race` sauber, Herstellername 0×.

## 96. Version 0.9.67 — Ping zeigt den echten MQTT-Fehler (warum „offline"?)

Der Reconnect-Takt ist logisch korrekt; bleibt ein Drucker trotzdem offline, scheitert die MQTT-Verbindung selbst. Den Grund kennt das Programm bereits (`connect_err`), zeigte ihn aber nirgends. Das Ping-Fenster (Klick auf die Status-Pille) blendet jetzt eine Zeile „MQTT (Steuerung)" ein: „verbunden" oder „nicht verbunden — <Grund>". Typische Gründe: „not authorized"/„bad user name or password" → der LAN-Zugangscode stimmt nicht mehr (kommt bei Bambu nach Firmware-Update/LAN-Reset vor); „connection refused"/„timeout" → Port/Modus. Damit sehen wir endlich, warum ein Drucker offline bleibt, obwohl Port 8883 offen ist.

Geprüft: `go test` sauber, JS-Syntaxcheck, Herstellername 0×.

## 97. Version 0.9.68 — Firmware-Update direkt am Drucker (Detailmenue)

Update-Knopf jetzt im Detailmenue: steht ein Update an, erscheint "Firmware installieren (N)" -> upgrade_confirm. Geprueft: jsdom, go test, Herstellername 0x.

## 98. Version 0.9.69 — Stabile MQTT-Client-ID: Drucker mit offenem Port kommen wieder online

Der Ping hat den Fall geklärt: Ein Drucker mit ALLEN Ports offen (auch 8883) blieb „MQTT not connected" — die Verbindung hing, obwohl der Drucker erreichbar war. Ursache: Die Client-ID enthielt einen Zeitstempel, war also bei jedem Versuch neu. Bambu-Drucker erlauben aber nur EINE lokale MQTT-Verbindung. Eine hängengebliebene alte Verbindung (z. B. nach vielen Programm-Neustarts) belegte den Platz, und eine neue mit anderer ID konnte ihn nicht übernehmen → sie hing.

Jetzt ist die Client-ID stabil (`druckerfarm-<seriennummer>`, ohne Zeitstempel). Beim erneuten Verbinden trennt der Drucker laut MQTT-Standard die alte Sitzung mit gleicher ID selbst und nimmt die neue an. Zusammen mit dem 30-s-Reconnect-Takt (0.9.66) sollte ein erreichbarer Drucker damit zuverlässig wieder online kommen.

Zusätzlich: Das Ping-Fenster zeigt die MQTT-Zeile jetzt live — nach dem Auslösen des Reconnects wird sie ein paar Sekunden lang nachgezogen, sodass man sofort sieht, ob die Verbindung steht.

Geprüft: `go test` und `-race` sauber, JS-Syntaxcheck, Herstellername 0×.

## 99. Version 0.9.70 — Ping zeigt den echten MQTT-Fehler + „Neu verbinden"-Knopf

Bisher hat der Ping selbst einen Reconnect ausgelöst und dabei den Fehlergrund zurückgesetzt — deshalb stand bei „MQTT (Steuerung)" nur „not connected" ohne Grund. Jetzt holt der Ping nur den frischen Status und zeigt den TATSÄCHLICHEN aktuellen Fehler (z. B. „connection timeout", „not authorized"). Der Reconnect ist ein eigener Knopf „🔄 Neu verbinden" im Ping-Fenster; drückt man ihn, wird die MQTT-Zeile in den nächsten ~12 s live nachgezogen, bis die Verbindung steht.

Damit sehen wir bei einem Drucker mit offenem Port 8883, der trotzdem offline bleibt, endlich den echten Grund — die Grundlage, um es gezielt zu beheben.

Geprüft: `go test` sauber, JS-Syntaxcheck, Herstellername 0×.

## 100. Version 0.9.71 — „not Authorized" klar erklärt + Direktknopf zum Zugangscode

Der Ping hat den Fall woobly 27 eindeutig gemacht: alle Ports offen, aber MQTT „not Authorized" — der Drucker lehnt die Anmeldung ab, weil der gespeicherte LAN-Zugangscode nicht mehr stimmt (Bambu ändert ihn z. B. nach Firmware-Update oder LAN-Reset). Das ist kein Tool-Fehler, sondern eine Drucker-Einstellung.

Damit das nicht rätselhaft bleibt, zeigt das Ping-Fenster bei einem Anmeldefehler jetzt einen deutlichen Hinweis („Zugangscode stimmt nicht mehr — aktuellen LAN-Code am Drucker ablesen und hier eintragen") und einen Knopf „🔧 Zugangscode ändern", der direkt in die „+ Drucker"-Liste springt und das Bearbeiten dieses Druckers öffnet.

Geprüft: jsdom (bei „not Authorized" erscheinen Hinweis und Zugangscode-Knopf), `go test` sauber, JS-Syntaxcheck, Herstellername 0×.

## 101. Version 0.9.72 — Schließen-X in den Modal-Boxen

Die Modal-Boxen bekommen oben rechts ein „✕" zum Schließen: die Detailbox im Grid („mehr") und das Ping-Fenster. Klick aufs ✕ schließt die Box; die bisherigen Wege (Klick daneben bzw. „Close"-Knopf) bleiben.

Geprüft: jsdom (✕ ist in Grid-Modal und Ping-Fenster vorhanden; Klick schließt), `go test` sauber, JS-Syntaxcheck, Herstellername 0×.

## 102. Version 0.9.73 — Vier Wünsche: Blinken je Drucker, Video-Vollbild, Startfilter, Kamera „Privat"

1. **Blinken pro Drucker einzeln ausschaltbar.** Im Detailmenü eines blinkfähigen Druckers (X1E/X2D) gibt es jetzt „🔕 Blinken aus" / „🔔 Blinken ein" — zusätzlich zum globalen Schalter. Der Zustand wird gespeichert; beim Abschalten geht das Licht sofort auf normal zurück.
2. **Vollbild pro Video.** Unten rechts im Videobild jeder Kachel (Live oder Snapshot) erscheint beim Überfahren ein ⛶-Knopf, der genau dieses Video im Browser-Vollbild öffnet.
3. **Startfilter wählbar.** In den Einstellungen unter „🏁 Startansicht" lässt sich einstellen, welcher Übersichts-Filter beim Start aktiv ist: Alle / Nur Online / Nur Offline. Standard ist jetzt **Alle** (vorher fest „Online").
4. **Kamera pro Drucker aus („Privat").** Im Detailmenü „🚫 Kamera aus (Privat)" schaltet die Kamera dauerhaft ab: die Kachel zeigt dann mittig „Privat" (wie „Offline"/„In Reparatur"), es wird kein Stream aufgebaut. „📷 Kamera einschalten" holt sie zurück. Gilt in Übersicht und Vollbild, mit Selbstheilung beim Umschalten.

Geprüft: jsdom (alle vier: Startfilter geladen/gespeichert, Video-Vollbild-Knopf, Blink-Knopf + POST, Kamera-Knopf + POST + „Privat"-Overlay), Go-Test (Backend-Schalter, Settings, errorLightLoop respektiert Per-Drucker-Aus), `go test` und `-race` sauber, Herstellername 0×.

## 103. Version 0.9.74 — „Im Netz suchen" im Ping-Fenster (neue IP nach DHCP)

Ein Gerät von außen zu einer Neuanmeldung/neuen IP zu zwingen, geht nicht — solange es auf keinem Port antwortet, ist es unerreichbar. Der häufigste Grund für „an der gespeicherten IP tot, aber eingeschaltet" ist aber eine per DHCP geänderte IP. Dafür gibt es jetzt im Ping-Fenster den Knopf „🔍 Im Netz suchen": Er startet die Netzwerksuche (SSDP) und findet den Drucker über die Seriennummer. Hat er eine neue IP, wird sie automatisch übernommen und das Fenster geschlossen; der Drucker verbindet sich dann neu. Findet die Suche nichts, meldet das Tool klar: Gerät aus, WLAN weg oder anderes Netz/VLAN — das kann kein Monitor-Tool umgehen. (Dieselbe Erkennung „umgezogener" Drucker gibt es weiterhin in der großen Netzwerksuche unter „+ Drucker".)

Geprüft: jsdom (Knopf erscheint; Suche findet die Seriennummer an neuer IP und übernimmt sie per PUT), `go test` sauber, JS-Syntaxcheck, Herstellername 0×.
