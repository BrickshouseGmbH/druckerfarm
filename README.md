<div align="center">

# Druckerfarm

**A local dashboard for monitoring and managing a farm of FDM 3D printers.**

Live video, status, AMS/filament, file sync and more — for many printers at once, on your own network.

_Public beta 0.9.20 · Windows 10 / 11 · by Brickshouse GmbH_

### [⬇️ Download the latest release](../../releases/latest)

</div>

> **Not affiliated with Bambu Lab.** Druckerfarm is independent third‑party software. Bambu Lab is not affiliated with, does not endorse, and is not part of this program. All product names and trademarks belong to their respective owners.

---

## ✨ What it does

Druckerfarm gives you a single window to watch and control a whole room of printers on your local network — no cloud account required.

- **📹 Live video & snapshots** — WebRTC live streams and periodic snapshots (2 s / 5 s / 10 s, 720p / 1080p), in a grid, list or fullscreen view.
- **📊 Status at a glance** — print progress, layer, time remaining, nozzle/bed/chamber temperatures and print state for every printer.
- **🎨 AMS & filament** — see each AMS unit and tray; assign material, colour and profile per slot.
- **🌡️ Temperature control** — set nozzle and bed targets.
- **🖨️ Print from your PC** — send a file to a printer, with per‑colour filament mapping.
- **🗂️ File sync & SD management** — browse, upload, download, delete and search print files on the printer's storage.
- **🔄 Firmware update check** — reads what each printer itself reports and marks pending updates with an “⬆ Update” badge. It only checks — it never triggers an update.
- **📡 Discovery & import** — find printers on the network via SSDP, or add them by CSV or by hand.
- **⚠️ Error alerts** — surfaces HMS/print errors; for models without their own signal light (e.g. X1E / X2D) it can blink the chamber light on error or pause, and reset it afterwards.
- **🔎 Filters & favourites** — online, offline, paused, favourites and **done** (100 %).
- **🌍 Bilingual** — full German and English UI, chosen on first start and switchable any time.
- **🧰 Robust by design** — operating‑hours counter, a network‑pause switch for maintenance, go2rtc process management, and automatic reload after the PC wakes from sleep.

---

## 🖨️ Tested printers

Tested on **X1E, X2D, H2D and H2C** running on **Windows 10 and 11**.

Other Bambu models that expose the local LAN interface are expected to work for monitoring; please report your results.

---

## 🚀 Getting started

1. **Download** the latest `druckerfarm.exe` from [Releases](../../releases/latest). It is portable — no installer.
2. **Run it.** On first launch you choose the language; the interface opens in its own Edge/Chrome app window (or your default browser as a fallback).
3. **Prepare each printer** (on the printer's touchscreen):
   - Enable **LAN Mode** and note the **Access Code**.
   - For remote control (pause/resume/stop, temperature, print start, chamber light) also enable **Developer Mode** — newer firmware rejects control commands otherwise. *Monitoring and video work without it.*
4. **Add your printers** — scan the network, import a CSV, or enter them manually (name, IP, access code, serial, model).
5. **First video start** downloads two helpers automatically into the data folder: **go2rtc** and **ffmpeg** (verified via SHA‑256). After that, streams and snapshots are available.

---

## 🔧 Requirements

- **Windows 10 or 11**
- **Microsoft Edge or Google Chrome** (recommended, for the dedicated app window)
- Printers reachable on the **local network**, in **LAN Mode**, with their **Access Code**
- Internet access on first run to fetch go2rtc/ffmpeg

---

## 🔒 Data & privacy

- Your printer list — including access codes — is stored **only on your PC** at
  `%APPDATA%\Druckerfarm\config.json`.
- **No access data is baked into the program.** The `.exe` contains no IPs, access codes or serial numbers.
- Everything talks to your printers directly on the local network. There is no cloud account and no telemetry.
- If you share the `.exe`, you do **not** share your credentials — only `config.json` contains them, so keep that file private.

---

## 🗑️ Uninstall

Druckerfarm is portable — there is no installer and nothing is written to the registry.

1. Delete `druckerfarm.exe`.
2. Delete `%APPDATA%\Druckerfarm` (config, logs and the downloaded go2rtc/ffmpeg — the largest part).
3. Delete `%TEMP%\druckerfarm-profile` (the app‑window profile).

---

## 🌍 Translations

The interface ships in German and English. Translations live in editable JSON files under `lang/` and are compiled into the program. To adjust or extend them, edit `lang/de.json`, `lang/en.json` or `lang/de-en.json` and rebuild — see `lang/README.md`.

---

## 📦 Third‑party software

Druckerfarm builds on open‑source software (paho.mqtt.golang, gorilla/websocket, jlaffaye/ftp, hashicorp/go‑multierror & errwrap, golang.org/x/net & x/sync) and, at runtime, **go2rtc** (MIT) and **ffmpeg** (BtbN builds). go2rtc and ffmpeg are downloaded on demand and are **not** bundled with the program. Full attribution and licence texts are in [`THIRD_PARTY_LICENSES.md`](THIRD_PARTY_LICENSES.md).

---

## ⚖️ Legal

Druckerfarm interoperates with the printer's **local network interface** using publicly documented, community‑gathered knowledge of the LAN protocol. It contains no manufacturer code, firmware or keys, and it does not circumvent any protection — control commands rely on the printer's own Developer Mode.

“Bambu Lab”, model names and related marks are trademarks of their respective owners. This project is **not affiliated with, endorsed by, or supported by** Bambu Lab.

---

<div align="center">

© Brickshouse GmbH

</div>
