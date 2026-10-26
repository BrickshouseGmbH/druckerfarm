# Druckerfarm – Navigations- & Seitenkonzept

Stand: 0.9.76. Beschreibt **jeden Menüpunkt/Link, wohin er führt und was auf der jeweiligen
Seite passiert**. Immer mitpflegen, wenn sich Navigation oder Seiten ändern.

Technik-Kurzinfo: Einzelseite (`ui.go`, `dashboardHTML`). „Seiten" sind
`<div class="view" id="view-…">`, umgeschaltet per `sv(name)`. Grid/Liste/Fullscreen
schaltet `setView(name)`.

---

## 1. Kopfzeile (Hauptmenü)

| Link / Button | Ziel (Funktion) | Was passiert |
|---|---|---|
| **⊞ Grid** (links) | `sv('grid')` → `view-uebersicht`, `setView('grid')` | Kachelansicht aller Drucker mit Live-Bild/Snapshot. Stream-Leiste + Filterpillen sichtbar. Standard beim Start. |
| **≡ Liste** (links) | `sv('list')` → `view-uebersicht`, `setView('list')` | Gleiche Drucker als sortierbare Tabelle (kein Video). Stream-Leiste ausgeblendet. |
| **🔍 Suche** (Mitte) | `flt(value)` (id `searchInput`) | Zentrales Suchfeld in der Hauptnavigation. Filtert Grid UND Liste nach Name/IP. Solange ein Begriff gesetzt ist, zeigt ein **Hinweisbanner** (`#search-active-hint`, via `updateSearchActiveHint`) über dem Grid „Suche aktiv: »…« — der Filter wirkt zusätzlich" + Button „Suche leeren" — damit ein Modellwechsel nicht grundlos leer wirkt. |
| **⚙ Aufträge** (rechts) | `sv('jobs')` → `view-jobs` | Jobs-Seite: laufende Aufträge (Upload/Sync/FW) + **Upload-Historie** + Job-Planer. Badge = Anzahl laufender Jobs. |
| **➕ Drucker** (rechts, nur Admin) | `openActionModal('import')` | Öffnet das **gemeinsame Aktions-Modal** auf dem Tab „Drucker hinzufügen". |
| **⬆ Datei hochladen** (rechts) | `openActionModal('sync')` | Öffnet dasselbe **Aktions-Modal** auf dem Tab „Datei hochladen". |
| **⚙ Einstellungen** (rechts) | `sv('setup')` → `view-setup` | Sprache, Startfilter, Streaming/go2rtc, Auth (Benutzer/Rollen/Tokens), KI-Konnektoren, Cloudflare, YouTube, Version/Build, Lizenzen. |
| **Benutzer-Chip** (rechts) | – | Nur bei Remote-Login: Name + Rolle + „Abmelden". Bei lokalem Zugriff ausgeblendet. |

Hinweise:
- „➕ Drucker" und „⬆ Datei hochladen" öffnen **ein einziges Modal mit zwei Tabs** – die
  Buttons springen nur direkt auf den jeweiligen Tab. Navigiert NICHT weg. Schließen: ✕,
  Klick auf den Hintergrund oder **Esc**.
- „➕ Drucker" ist nur für Rolle **admin** sichtbar.
- Einfach/Experte-Umschalter wurde entfernt – die fortgeschrittenen Regler sind immer sichtbar.

## 2. Unterleiste der Übersicht (`#sub-nav`, nur in Grid/Liste)

| Element | Ziel | Was passiert |
|---|---|---|
| **⊡ Fullscreen** | `setView('fsc')` → `view-fsc` | Vollbild-Kachelansicht (ohne Kopf/Leisten). |
| **⛶ Fullscreen Video Only** | `setView('fsv')` → `view-fsv` | Reines Vollbild-Video (Rahmen + Balken, kein Text). |

## 3. Stream-Leiste (`#stream-bar`, nur in Grid)

| Element | Ziel | Was passiert |
|---|---|---|
| ▶ / ⏹ | `startAll()` / `stopAll()` | Alle Videostreams starten / stoppen. |
| Profile (📷 2s/5s/10s, 720p/1080p) | `setSnapMode()` / `setVideoMode()` | Snapshot-Intervall bzw. Videoauflösung. |

## 4. Filter-/Statusleiste (`#top-bar`, Grid & Liste)

| Element | Ziel | Was passiert |
|---|---|---|
| Pillen **Alle / Online / Offline / Pausiert / Favoriten / Fertig** | `setPillFilter(...)` | Filtert die angezeigten Drucker. Zahl = Anzahl je Zustand. Startfilter aus Einstellungen. |
| Modell-Pillen | `setPillFilter` | Nach Modell filtern. (Suchfeld ist in die Hauptnavigation gewandert.) |

## 5. Views (Seiten) – Inhalt & Aktionen

| View-ID | Erreichbar über | Inhalt / was passiert |
|---|---|---|
| `view-uebersicht` | Grid / Liste | Container für Grid + Liste inkl. Sub-Nav/Stream-/Top-Bar. |
| `view-grid` | ⊞ Grid | Kacheln je Drucker: Video/Snapshot, Status, Fortschritt, Pillen, „mehr" → Detail-Modal. |
| `view-list` | ≡ Liste | Tabelle je Drucker, sortierbar; Name/„mehr" → Detail-Modal. |
| `view-jobs` | ⚙ Aufträge | Laufende Aufträge (Pille), **Upload-Historie** (verschoben aus File Sync), Job-Planer pro Figur/Modell. |
| `view-import` | ➕ Drucker (Modal-Tab) | Drucker hinzufügen: SSDP-Suche, CSV-Import, Optionen Kamera/Blink/Reparatur. **Lebt im Aktions-Modal.** |
| `view-sync` | ⬆ Datei hochladen (Modal-Tab) | File Sync: Datei-Upload auf gewählte Drucker, SD-Dateiliste, Speicher-/FTP-Check. **Lebt im Aktions-Modal.** |
| `view-setup` | ⚙ Einstellungen | Siehe Kopfzeile. |
| `view-fsc` | ⊡ Fullscreen | Vollbild-Kacheln. |
| `view-fsv` | ⛶ Fullscreen Video Only | Vollbild-Video, minimaler Rahmen. |

## 6. Modale Dialoge

| Modal | Ausgelöst durch | Inhalt / was passiert |
|---|---|---|
| **Aktions-Modal** (`#action-modal`) | ➕ Drucker / ⬆ Datei hochladen | **Ein** Modal mit zwei Tabs: „➕ Drucker hinzufügen" (`view-import`) und „⬆ Datei hochladen" (`view-sync`). Beide Ansichten werden beim Laden **einmalig** ins Modal gehängt (`hostActionViews`), nie wieder im DOM verschoben → Grid bleibt unberührt. Tab-Wechsel: `switchActionTab()`. **Über** der gemeinsamen Druckerliste: ein Suchfeld (`#modalListSearch`) mit **zwei Buttons** — der Begriff wird **nicht** beim Tippen angewendet (das ließ früher alle Drucker verschwinden, sobald man einen Dateinamen eintrug). **„🔍 Drucker suchen"** (`searchPrinters`, auch per Enter) filtert die Liste nach Name/IP (`filterUprList` → `UPR_FILTER`); **„🔎 Datei suchen"** (`modalFileSearch` → `startSearch`, min. 3 Zeichen) sucht die Datei/Figur über die SD-Karten der ausgewählten Drucker, ohne die Liste zu filtern. ✕ leert Feld und Filter. Darunter die Modell-Filter-Pillen (`#sync-model-filter`, via `renderSyncModelPills`). Findet der Live-Filter nichts: „**Kein Drucker gefunden**" statt leerer Liste. Kopfzeile rechts: **▁ Minimieren** (`minimizeActionModal` — versteckt das Modal, **Upload läuft weiter**, Wiederherstellen über Chip `#amx-restore` unten rechts) + **✕ Schließen**. Schließen: ✕ / Hintergrund / Esc (bricht einen Upload NICHT ab — der läuft im Hintergrund). |
| **Detail-Modal** (`openDetailModal(ip)`) | „mehr"/Name in Grid & Liste | Pro Drucker: Status (FAILED bestätigen), farbige Druckvorschau (.3mf/.gcode.3mf, **alle Platten**, über `/api/plates`+`/api/preview`, PC-Cache mit SIZE/MDTM-Invalidierung), Farbabgleich Datei↔AMS, Steuerung (Pause/Weiter/Abbrechen, Datei wählen/hochladen+starten), Fehler-Verlauf, Geräte-/KI-Blöcke. |
| **Medien-Modal** | aus Detail-Modal / File Sync | Video-/Snapshot-Verwaltung, Timelapse-Toggle. |

---

## Datei-Upload & FTP (gegen „421 too many connections")
- **Farmweit nur EINE FTP-Session gleichzeitig** (`ftpGlobalMax=1` in ftpgate.go). go2rtc nutzt kein FTP (nur Video), ist also nie die Ursache.
- **Upload = eine Session pro Drucker für ALLE Dateien** (`/api/sync/uploadmany` + Python-Action `stormany`): einmal verbinden, nacheinander STOR, NDJSON-Stream pro Datei zurück an den Client (Live-Schritte). Kein connect/quit pro Datei mehr.
- **421-Retry** schon beim Verbinden im Python-Helper (kurz warten, erneut) — kein Drucker-Neustart nötig.
- **Stopp** bricht die laufende Datei sofort ab (AbortController → Request-Kontext endet → Python-Prozess stirbt). Pause wirkt zwischen Druckern.
- **Zeilen-Hintergrund im Upload (4 Farben):** blau = ausgewählt · gelb animiert = Verbindungsaufbau · gelb Stufen = Upload-Fortschritt · grün = fertig · rot = Fehler/keine Verbindung. Der **Druckfortschritt** ist ein dünner grüner Balken wie im Grid (`.upr-pbar`).

## UI-Invarianten (nicht brechen!)
- **Kacheln im Grid immer gleich hoch:** `#printerGrid{align-items:stretch}`, und die Fuß-Zeilen
  der Kachel reservieren Höhe per `min-height` (`.df-fuss`/`-r1`/`-r2`). Fehlt Text (IDLE/offline),
  bleibt die Zeile LEER — die Karte zieht sich NICHT zusammen. Nie wieder `align-items:start`.

## Entscheidungen / Notizen
- **Add Printer + Upload File = ein Modal** mit zwei Tabs (statt zwei getrennter Modals).
- **Upload-Historie** liegt jetzt in **Jobs** (nicht mehr in File Sync); wird beim Öffnen von
  Jobs geladen (`loadUploadLog()` in `sv('jobs')`).
- **Einfach/Experte** wurde komplett entfernt.
- Bugfix: Beide Modal-Ansichten werden nur **einmal** ins Modal gehängt (vorher Hin-und-Her-
  Verschieben → „alle Drucker verschwanden").
- Default-Ansicht beim Start = **Grid**.
