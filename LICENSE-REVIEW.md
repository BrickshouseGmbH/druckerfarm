# Lizenz- und Rechte-Prüfung

Stand: Version 0.9.14. Diese Prüfung betrifft mögliche Verletzungen von
Urheberrecht, Lizenzen und Marken — mit besonderem Blick auf Bambu Lab und
Reverse-Engineering — sowie auf die verwendeten Open-Source-Projekte.

Hinweis: Dies ist eine technische Einschätzung, keine Rechtsberatung. Für eine
verbindliche Bewertung sollte im Zweifel eine auf IT-/IP-Recht spezialisierte
Kanzlei hinzugezogen werden.

## 1. Bambu Lab — eingebettete Artefakte

Geprüft wurde, ob herstellereigenes Material in der Programmdatei steckt.
Ergebnis: nichts dergleichen gefunden.

- Keine Zertifikate, keine privaten oder öffentlichen Schlüssel, keine Firmware,
  keine `.pem`/`.crt`/`.key`/`.bin`-Dateien im Projekt.
- Kein kopierter Slicer-Code (kein OrcaSlicer/Bambu Studio/PrusaSlicer/Cura).
- Der Herstellername kommt in der Programmdatei nicht vor (geprüft: 0 Treffer).
- Der MQTT-Benutzername `bblp` ist der öffentlich bekannte, feste Kennwert des
  lokalen LAN-Protokolls — kein Geheimnis. Das Passwort ist der Zugangscode, den
  der Anwender selbst für seinen eigenen Drucker einträgt.

## 2. Reverse-Engineering des lokalen Protokolls

Druckerfarm spricht mit den Druckern über deren lokale Schnittstellen: MQTT über
TLS (Port 8883), SSDP-Erkennung, FTP (Port 990) und RTSP-Kamerastreams. Diese
Umsetzung beruht auf öffentlich dokumentiertem, von der Gemeinschaft
zusammengetragenem Wissen über das LAN-Protokoll.

Einschätzung:
- Protokolle, Nachrichtenformate und Feldnamen sind Fakten und als solche nicht
  urheberrechtlich geschützt. Es wurde kein Herstellercode übernommen.
- Die Kommunikation dient der **Interoperabilität** mit Geräten, die dem
  Anwender gehören. Das ist in der Regel zulässig (in der EU u. a. über das
  Interoperabilitäts-Interesse der Software-Richtlinie).
- **Kein Kopierschutz wird umgangen.** Die neuere Firmware verlangt für
  Steuerbefehle eine Signatur; Druckerfarm setzt dafür bewusst auf den am Gerät
  aktivierbaren **Developer Mode** (LAN-Freigabe durch den Besitzer) und
  **nicht** auf einen aus der Firmware extrahierten Herstellerschlüssel. Genau
  das wäre der heikle Punkt (Umgehung technischer Schutzmaßnahmen) gewesen — er
  wird vermieden. `InsecureSkipVerify` beim TLS ist unkritisch: es betrifft nur
  das selbstsignierte Zertifikat des lokalen Druckers, nichts wird ausgehebelt.

## 3. Marken

„Bambu Lab", „Bambu" sowie Modellnamen (X1C, P1S, H2D, AMS …) sind Marken des
Herstellers. Druckerfarm nutzt den Herstellernamen nicht zur Kennzeichnung des
eigenen Produkts (aus der Programmdatei entfernt). Modellkürzel werden nur zur
Zuordnung/Kompatibilität verwendet (nominative Nennung). Empfehlung: in Werbung
und Produktauftritt keinen Eindruck einer offiziellen Verbindung oder
Unterstützung durch den Hersteller erwecken; ggf. einen Zusatz wie „nicht mit
… verbunden oder von … unterstützt" führen.

## 4. Open-Source-Abhängigkeiten

Alle eingebundenen Go-Bibliotheken sind freizügig (BSD/MIT/ISC) oder dateiweise
schwach-copyleft (MPL-2.0, EPL-2.0/EDL-1.0). **Keine** verlangt die Offenlegung
des Druckerfarm-Quellcodes. Die einzige bestehende Pflicht ist die **Nennung**
(Copyright-Hinweise und Lizenztexte mitgeben). Diese Lücke wird mit der neuen
Datei `THIRD_PARTY_LICENSES.md` geschlossen, die mit ausgeliefert wird.

go2rtc (MIT) und ffmpeg (BtbN-Builds, GPL) werden **zur Laufzeit nachgeladen**
und nicht mit Druckerfarm gebündelt oder weiterverbreitet. Dadurch bleiben die
GPL-Pflichten von ffmpeg bei BtbN als Verteiler. Diese Trennung sollte so
bleiben — ffmpeg nicht in die Programmdatei oder das Auslieferungspaket
aufnehmen.

## 5. Eigene Bildmittel — offener Punkt

Das Startbild `assets/splash.jpg` (Gorilla) trägt keine Urheberangabe und wurde
mit ffmpeg/libavcodec kodiert. Die Herkunft lässt sich aus der Datei nicht
belegen. Bitte sicherstellen, dass dieses Bild (und `logo.svg`) selbst erstellt
oder ordnungsgemäß lizenziert ist. Das ist der einzige Punkt, den nur du klären
kannst — technisch ist er von außen nicht prüfbar.

## Ergebnis

- Kein Herstellermaterial eingebettet, kein Kopierschutz umgangen, kein
  Fremdcode kopiert → aus dieser Sicht keine erkennbare Verletzung.
- Open-Source-Pflichten werden mit `THIRD_PARTY_LICENSES.md` erfüllt.
- Zu klären bleibt allein die Herkunft der eigenen Bildmittel (Punkt 5).
