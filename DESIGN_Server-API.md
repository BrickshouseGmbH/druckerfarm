# Design-Konzept: Server-/API-Modell für Druckerfarm

Stand: Entwurf zum Draufschauen. **Nichts hiervon ist gebaut.** Erst wenn die
offenen Entscheidungen am Ende geklärt und freigegeben sind, wird umgesetzt —
und dann in kleinen Phasen, damit das laufende Tool (0.9.75) nichts einbüßt.

---

## 1. Das Grundproblem, das wir lösen

Heute verbindet **jeder PC** direkt zu **jedem Drucker**: eigener MQTT-Status,
eigene FTP-Zugriffe, eigener Kamera-Stream, und die Zugangscodes liegen auf
jedem Rechner. Das hat drei Nachteile:

- **Verbindungs-Kollisionen.** Mehrere FTP-Clients auf denselben Drucker → genau
  das 421 „too many connections", das wir gerade hatten.
- **Passwörter überall.** Jeder Rechner kennt alle Zugangscodes.
- **Kein gemeinsamer Stand.** Jeder sieht/steuert für sich, nichts ist koordiniert.

## 2. Das Zielbild

Eine Instanz läuft als **Server** (der „Hub"). Nur sie redet mit den Druckern.
Alle anderen — Menschen im Browser oder andere Programme — reden mit dem **Server**,
nie direkt mit den Druckern.

```
        Drucker 1 ─┐
        Drucker 2 ─┤   (nur EIN Client pro Drucker: der Server)
          ...      ├──►  DRUCKERFARM-SERVER  ◄─── Browser (Person A, Viewer)
        Drucker 53─┘        (Hub, hält                ◄─── Browser (Person B, Operator)
                             Zugangscodes)             ◄─── Fremdprogramm (API-Token)
```

Aus Sicht der Drucker gibt es dann **einen** FTP-/MQTT-Client statt zehn → das
421-Problem verschwindet als ganze Klasse, nicht nur im Einzelfall.

**Wichtig:** Das Tool ist dafür schon **halb gebaut.** Es ist bereits ein
Webserver mit internen `/api/*`-Endpunkten, und die `/broadcast`-Ansicht ist
schon eine entkoppelte Nur-Zuschauen-Seite. Was fehlt, ist im Wesentlichen:
**Anmeldung, Rollen und ein sauber abgesicherter API-Zugang.**

## 3. Rollen

Drei Stufen reichen erfahrungsgemäß:

| Rolle | Darf | Beispiel |
|------|------|---------|
| **Viewer** | nur ansehen (Status, Kameras, Listen) | Kollege, der zuschaut |
| **Operator** | ansehen + Drucke steuern (pausieren/fortsetzen/abbrechen, Datei senden, Sync) | Person an der Farm |
| **Admin** | alles inkl. Drucker anlegen/löschen, Zugangscodes, Einstellungen, Update | du |

Die Zugangscodes der Drucker sieht **nur der Admin/der Server** — Viewer und
Operator brauchen sie nie.

## 4. Anmeldung & Zugang

Zwei getrennte Wege, weil Menschen und Programme unterschiedlich zugreifen:

- **Menschen (Browser):** Login mit Benutzer + Passwort, danach ein Sitzungs-Cookie.
  Wer nicht angemeldet ist, sieht höchstens die (optionale) offene Broadcast-Seite.
- **Programme (API):** ein **API-Token** pro Programm (langer Zufallsschlüssel,
  im Header mitgeschickt). Jederzeit im Tool erzeugbar und widerrufbar. Kein
  Passwort im Klartext in fremden Skripten.

Beides bindet an dieselben Rollen: ein Token kann „nur lesen" oder „steuern".

## 5. Die API

Die internen `/api/*`-Endpunkte gibt es schon — sie werden dokumentiert,
versioniert (`/api/v1/...`) und mit der Rollenprüfung versehen. Grob:

- **Lesen (Viewer+):** Druckerliste & Status, Jobs, Job-Planer, Health, Kamera-Snapshot.
- **Steuern (Operator+):** Druck pausieren/fortsetzen/abbrechen, Datei senden,
  File-Sync starten, Timelapse.
- **Verwalten (Admin):** Drucker/Konnektoren/Einstellungen, Zugangscodes, Update.

Jede Antwort ist JSON, jeder schreibende Aufruf prüft die Rolle. So kann ein
ERP/MES oder ein eigenes Skript andocken, ohne je einen Drucker direkt zu berühren.

## 6. Sicherheit (ehrlich, das ist der heikle Teil)

- **Standardmäßig nur im LAN.** Der Server lauscht im lokalen Netz; nach außen
  nur bewusst (der bestehende Cloudflare-Tunnel bleibt der kontrollierte Weg).
- **HTTPS/TLS** als Option, damit Passwörter/Tokens nicht im Klartext übers Netz gehen.
- **Zugangscodes bleiben serverseitig**, verschlüsselt gespeichert, nie in einer
  API-Antwort.
- **Tokens widerrufbar**, mit Ablaufdatum möglich, jede Nutzung protokolliert.
- **Rate-Limit** gegen Missbrauch.
- Kurz: Eine API verlagert das Passwort-Problem nur, wenn sie schlecht
  abgesichert ist — deshalb ist Auth + TLS + Token-Verwaltung Teil des Kern-Designs,
  kein Nachgedanke.

## 7. Die Kehrseite

- **Eine Maschine wird kritisch.** Läuft der Server-PC nicht, sieht keiner mehr was.
  (Gegenmittel: der Server-PC ist einfach der stabilste; optional Auto-Neustart.)
- **Etwas mehr Komplexität** beim ersten Einrichten (Benutzer anlegen, Rollen vergeben).
- Für „ein, zwei Leute am selben Rechner" wäre das Overkill — es lohnt ab
  mehreren Personen/Rechnern, was bei deiner Farm gegeben ist.

## 8. Vorschlag: Rollout in Phasen (bleibt 0.9.75, nichts bricht)

1. **Phase 1 — Anmeldung & Rollen** auf der bestehenden Oberfläche.
   Login-Seite, Benutzerverwaltung, „Viewer/Operator/Admin". Nach außen ändert
   sich sonst nichts. Kleinster, sofort nützlicher Schritt.
2. **Phase 2 — API mit Tokens.** `/api/v1/...` dokumentiert, Token-Verwaltung im
   Tool, Rollenprüfung auf jedem Endpunkt. Damit können Fremdprogramme andocken.
3. **Phase 3 — „Nur-Zuschauen"-Verteilung.** Die anderen brauchen dann nur die
   Server-Adresse im Browser, keine Installation, keine Zugangscodes. (Teilweise
   über `/broadcast` schon da.)

Jede Phase ist für sich fertig und nutzbar; wir müssen nicht alles auf einmal bauen.

---

## 9. Offene Entscheidungen (bitte vor dem Bauen)

1. **Benutzer:** echte einzelne Konten (Name + Passwort je Person) — oder erstmal
   ein gemeinsames Passwort pro Rolle (einfacher, aber weniger nachvollziehbar)?
2. **Zuschauen ohne Login:** soll die reine Ansicht offen bleiben (wie heute
   `/broadcast`) und nur das **Steuern** einen Login verlangen — oder alles hinter Login?
3. **TLS/HTTPS im LAN:** brauchen wir Verschlüsselung schon im lokalen Netz
   (empfohlen, aber etwas Einrichtungsaufwand mit Zertifikat) — oder vorerst nur HTTP intern?
4. **Womit anfangen:** nur Phase 1 (Anmeldung/Rollen) zuerst bauen und den Rest
   später — oder gleich Phase 1+2 zusammen?

Sag mir zu diesen vier Punkten deine Richtung, dann setze ich die erste Phase um.
