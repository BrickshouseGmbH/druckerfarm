# Übersetzungen / Translations

Die Oberfläche von Druckerfarm wird aus diesen JSON-Dateien übersetzt:

- `de.json` — Deutsch
- `en.json` — Englisch

Beide werden beim Bauen fest in die Programmdatei eingebettet (`go:embed`).
Zum Anpassen also **vor** dem Erstellen der Exe hier bearbeiten und neu bauen.

## Eigene Übersetzung ändern

Jeder Eintrag ist ein Schlüssel-Wert-Paar. Nur den Wert (rechts) ändern, den
Schlüssel (links) so lassen:

    "btnAdd": "Hinzufügen",     ->  "btnAdd": "Add",

Beide Dateien müssen dieselben Schlüssel haben. Fehlt ein Schlüssel in einer
Sprache, fällt die Oberfläche für diesen Eintrag auf Deutsch zurück.

## Neuen Text übersetzbar machen

Im HTML bekommt ein Element das Attribut `data-i18n="meinSchluessel"` — sein
Text wird dann automatisch aus der JSON gesetzt. Varianten:

- `data-i18n="k"`      setzt den Textinhalt
- `data-i18n-html="k"` setzt HTML (für Text mit Formatierung)
- `data-i18n-ph="k"`   setzt den Platzhalter eines Eingabefelds

Dann in `de.json` UND `en.json` denselben Schlüssel `meinSchluessel` mit dem
jeweiligen Text ergänzen. Kein weiterer Eingriff nötig.

## Neue Sprache hinzufügen

Aktuell sind Deutsch und Englisch verdrahtet. Eine weitere Sprache erfordert
zusätzlich eine kleine Anpassung im Programm (Einbetten der neuen Datei und die
Auswahl) — sag Bescheid, dann richte ich das ein.

## de-en.json — die vollflächige Übersetzung

`de-en.json` ordnet **deutschen Text → englischen Text** zu. Bei englischer
Sprache übersetzt das Programm damit die komplette Oberfläche automatisch —
auch Text, der erst zur Laufzeit entsteht (Listen, Meldungen, Menüs). Deutsch
ist dabei die Ausgangssprache; Englisch entsteht aus dieser Tabelle.

Eintrag = deutscher Text links, englischer Text rechts, jeweils exakt so, wie er
angezeigt wird (mit Symbolen). Beispiel:

    "Gespeicherte Drucker": "Saved printers",
    "🧯 Alle go2rtc beenden": "🧯 Kill all go2rtc"

Fehlt ein Eintrag, bleibt dieser Text auf Deutsch. Findet sich also irgendwo
noch Deutsch im englischen Modus: den genauen Text hier als neuen Eintrag
ergänzen und neu bauen — mehr ist nicht nötig.
