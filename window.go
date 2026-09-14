package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func hideConsoleOnWindows() {
	// Handled by -ldflags="-H windowsgui" at build time.
}

func appUserDataDir() string {
	dir := filepath.Join(os.TempDir(), "druckerfarm-profile")
	// Das Profil wird bewusst NICHT mehr bei jedem Start geloescht. Ein kaltes
	// Profil zwingt Edge zu einer Erstinitialisierung, bei der sich der
	// gestartete Prozess selbst neu startet und sofort beendet — was frueher den
	// Server mitgerissen hat. Fuers Cache-Busting ist der Loeschvorgang ohnehin
	// unnoetig: serveUI liefert die Seite mit no-store aus.
	os.MkdirAll(dir, 0755)
	return dir
}

// buildEdgeArgs — suppress all extra Edge windows/popups
func buildEdgeArgs(url string) []string {
	profileDir := appUserDataDir()
	return []string{
		"--app=" + url,
		"--user-data-dir=" + profileDir,
		"--window-size=1400,900",
		// Prevent Edge welcome / first-run windows
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-first-run-ui",
		// Kill all background activity that spawns extra windows
		"--disable-extensions",
		"--disable-background-networking",
		"--disable-default-apps",
		"--disable-sync",
		"--disable-translate",
		"--disable-component-update",
		"--disable-client-side-phishing-detection",
		"--disable-features=TranslateUI,msImplicitSignin",
		"--metrics-recording-only",
		"--safebrowsing-disable-auto-update",
		"--no-pings",
		"--deny-permission-prompts",
		"--suppress-message-center-popups",
	}
}

// launchAppWindow opens Edge/Chrome in --app mode (no address bar, no tabs).
func launchAppWindow(url string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "windows":
		candidates := []string{
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft\\Edge\\Application\\msedge.exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Microsoft\\Edge\\Application\\msedge.exe"),
			filepath.Join(os.Getenv("LocalAppData"), "Microsoft\\Edge\\Application\\msedge.exe"),
			filepath.Join(os.Getenv("LocalAppData"), "Google\\Chrome\\Application\\chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Google\\Chrome\\Application\\chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Google\\Chrome\\Application\\chrome.exe"),
		}
		for _, browser := range candidates {
			if _, err := os.Stat(browser); err == nil {
				args := buildEdgeArgs(url)
				cmd := exec.Command(browser, args...)
				if err := cmd.Start(); err == nil {
					return cmd, nil
				}
			}
		}
		appWindowMode = "browser"
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		return nil, fmt.Errorf("Edge/Chrome nicht gefunden — im Standardbrowser geöffnet")

	case "darwin":
		candidates := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
		for _, browser := range candidates {
			if _, err := os.Stat(browser); err == nil {
				args := buildEdgeArgs(url)
				cmd := exec.Command(browser, args...)
				if err := cmd.Start(); err == nil {
					return cmd, nil
				}
			}
		}
		appWindowMode = "browser"
		exec.Command("open", url).Start()
		return nil, fmt.Errorf("Chrome/Edge nicht gefunden — im Standardbrowser geöffnet")

	default:
		for _, b := range []string{"google-chrome-stable", "google-chrome", "chromium-browser", "chromium", "microsoft-edge"} {
			if path, err := exec.LookPath(b); err == nil {
				args := buildEdgeArgs(url)
				for i, a := range args {
					args[i] = strings.ReplaceAll(a, "\\", "/")
				}
				cmd := exec.Command(path, args...)
				if err := cmd.Start(); err == nil {
					return cmd, nil
				}
			}
		}
		appWindowMode = "browser"
		exec.Command("xdg-open", url).Start()
		return nil, fmt.Errorf("Chrome/Chromium nicht gefunden — im Standardbrowser geöffnet")
	}
}

// appWindowMode sagt, wie das Fenster geoeffnet wurde: "app" = Edge/Chrome im
// App-Modus, "browser" = Standardbrowser als Rueckfallebene.
var appWindowMode = "app"

func waitForWindow(cmd *exec.Cmd, shutdown chan struct{}) {
	if cmd == nil {
		// Kein Chromium gefunden — die Seite laeuft im Standardbrowser, auf
		// dessen Prozess wir nicht warten koennen. Frueher kehrte die Funktion
		// hier einfach zurueck und schloss shutdown nie: die App lief dann
		// endlos weiter, auch nach dem Schliessen des Tabs. Stattdessen sofort
		// weitergeben, damit die Leerlauf-Erkennung uebernimmt.
		close(shutdown)
		return
	}
	cmd.Wait()
	// Small delay — Edge sometimes exits before fully closing
	time.Sleep(300 * time.Millisecond)
	close(shutdown)
}
