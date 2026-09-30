//go:build !windows

package browser

import (
	"ant-chrome/backend/internal/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Opt-in real-browser smoke test. Only temporary profiles and a harmless local
// MV3 fixture are used; there is no web traffic or interaction with user data.
func TestExtensionRealBrowserMultipleProfiles(t *testing.T) {
	binary := os.Getenv("ANT_EXTENSION_TEST_BROWSER")
	if binary == "" {
		t.Skip("set ANT_EXTENSION_TEST_BROWSER to a Chromium binary")
	}
	root := t.TempDir()
	m := NewManager(config.DefaultConfig(), root)
	m.ExtensionDAO = newTestExtensionDAO(t, root)
	source := filepath.Join(root, "fixture")
	if err := os.MkdirAll(source, 0700); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"manifest.json": `{"manifest_version":3,"name":"Ant QA Fixture","version":"1.0.0","permissions":["storage"]}`,
		"qa.html":       `<!doctype html><title>Ant Extension QA</title><h1>Installed</h1>`,
	}
	for name, value := range files {
		if err := os.WriteFile(filepath.Join(source, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	e, err := m.InstallExtensionDirectory(source)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.ExtensionDAO.SetDefaultInstall(e.ExtensionID, true); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		p := &Profile{ProfileId: fmt.Sprintf("real-%d", i), UserDataDir: filepath.Join(root, fmt.Sprintf("profile-%d", i))}
		m.Profiles[p.ProfileId] = p
		for restart := 0; restart < 2; restart++ {
			dirs, warnings := m.PrepareProfileExtensions(p, binary, p.UserDataDir, nil)
			if len(warnings) > 0 {
				t.Fatal(warnings)
			}
			state, err := m.ExtensionDAO.GetProfileExtensionRuntime(p.ProfileId, e.ExtensionID)
			if err != nil {
				t.Fatal(err)
			}
			if len(dirs) != 1 {
				t.Fatalf("expected one managed directory: %v", dirs)
			}
			checkRealExtensionPage(t, binary, p.UserDataDir, state.RuntimeExtensionID, dirs[0])
		}
	}
}

func checkRealExtensionPage(t *testing.T, binary, userData, id, launchDir string) {
	t.Helper()
	url := "chrome-extension://" + id + "/qa.html"
	cmd := exec.Command(binary, "--user-data-dir="+userData, "--no-first-run", "--no-default-browser-check", "--disable-background-networking", "--headless=new", "--remote-debugging-port=0", "--load-extension="+launchDir, url)
	var logs strings.Builder
	cmd.Stdout = &logs
	cmd.Stderr = &logs
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	deadline := time.Now().Add(15 * time.Second)
	base := ""
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(filepath.Join(userData, "DevToolsActivePort"))
		lines := strings.Split(string(data), "\n")
		if err == nil && len(lines) > 1 && lines[0] != "" {
			base = "http://127.0.0.1:" + lines[0]
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if base == "" {
		t.Fatalf("no debugging endpoint: %s", logs.String())
	}
	client := http.Client{Timeout: time.Second}
	loaded := false
	var lastTargets any
	for time.Now().Before(deadline) {
		res, err := client.Get(base + "/json/list")
		if err == nil {
			var targets []struct {
				Title string `json:"title"`
				URL   string `json:"url"`
				WS    string `json:"webSocketDebuggerUrl"`
			}
			_ = json.NewDecoder(res.Body).Decode(&targets)
			lastTargets = targets
			res.Body.Close()
			for _, target := range targets {
				if target.URL == url && target.Title != "Ant Extension QA" {
					conn, _, err := websocket.DefaultDialer.Dial(target.WS, nil)
					if err == nil {
						_ = conn.WriteJSON(map[string]any{"id": 1, "method": "Page.reload"})
						conn.Close()
					}
				}
				if target.URL == url && target.Title == "Ant Extension QA" {
					loaded = true
				}
			}
			if loaded {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	// Graceful browser shutdown flushes Preferences and releases all profile locks.
	res, err := client.Get(base + "/json/version")
	if err == nil {
		data, _ := io.ReadAll(res.Body)
		res.Body.Close()
		var v struct {
			WS string `json:"webSocketDebuggerUrl"`
		}
		_ = json.Unmarshal(data, &v)
		conn, _, err := websocket.DefaultDialer.Dial(v.WS, nil)
		if err == nil {
			_ = conn.WriteJSON(map[string]any{"id": 1, "method": "Browser.close"})
			conn.Close()
			_ = cmd.Wait()
			closed = true
		}
	}
	if !loaded {
		prefs, _ := readProfileJSON(filepath.Join(userData, "Default", "Secure Preferences"), false)
		t.Fatalf("Chromium did not load extension %s targets=%v extensions=%v: %s", id, lastTargets, prefs["extensions"], logs.String())
	}
	_ = os.Remove(filepath.Join(userData, "DevToolsActivePort"))
}
