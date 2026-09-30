package browser

import (
	"ant-chrome/backend/internal/config"
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// This fixture has no page/content scripts, external requests or account data.
func testMultiProfilePackage(t *testing.T) []byte {
	t.Helper()
	key := []byte("multi-profile-test-public-key")
	var archive bytes.Buffer
	w := zip.NewWriter(&archive)
	f, err := w.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte(`{"manifest_version":3,"name":"Multi profile fixture","version":"1.0.0","permissions":["storage"],"optional_permissions":["tabs"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 16+len(key)+4)
	copy(data, "Cr24")
	binary.LittleEndian.PutUint32(data[4:8], 2)
	binary.LittleEndian.PutUint32(data[8:12], uint32(len(key)))
	binary.LittleEndian.PutUint32(data[12:16], 4)
	copy(data[16:], key)
	return append(data, archive.Bytes()...)
}

func TestPrepareExtensionsAcross25ProfilesDeleteAndReimport(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows uses its external CRX installer; real-browser coverage is platform-specific")
	}
	root := t.TempDir()
	m := NewManager(config.DefaultConfig(), root)
	m.ExtensionDAO = newTestExtensionDAO(t, root)
	data := testMultiProfilePackage(t)
	e, err := m.InstallExtensionPackageBytes("", "test fixture", data)
	if err != nil {
		t.Fatal(err)
	}
	if !e.DefaultInstall {
		t.Fatal("newly imported package should default to inherited installation")
	}
	for i := 0; i < 25; i++ {
		id := fmt.Sprintf("profile-%02d", i)
		p := &Profile{ProfileId: id, ProfileName: id, UserDataDir: filepath.Join(root, id)}
		m.Profiles[id] = p
		for start := 0; start < 2; start++ {
			dirs, warnings := m.PrepareProfileExtensions(p, "unused-browser", p.UserDataDir, nil)
			if len(warnings) > 0 {
				t.Fatalf("%s start %d: %v", id, start, warnings)
			}
			state, err := m.ExtensionDAO.GetProfileExtensionRuntime(id, e.ExtensionID)
			if err != nil {
				t.Fatal(err)
			}
			if state.Status != ExtensionRuntimeStatusInstalled || state.RuntimeExtensionID != e.ExtensionID {
				t.Fatalf("runtime: %#v", state)
			}
			if len(dirs) != 1 || dirs[0] != persistentExtensionCodePath(p.UserDataDir, e.ExtensionID, e.Version) {
				t.Fatalf("wrong profile-local launch dirs: %v", dirs)
			}
		}
	}
	// Keep a profile-local setting through deletion/reimport, as user data.
	p := m.Profiles["profile-00"]
	storage := filepath.Join(p.UserDataDir, "Default", "Local Extension Settings", e.ExtensionID, "fixture.txt")
	if err = os.MkdirAll(filepath.Dir(storage), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(storage, []byte("keep me"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = m.RemoveExtensionFromStoppedProfiles(e.ExtensionID); err != nil {
		t.Fatal(err)
	}
	if err = m.ExtensionDAO.Delete(e.ExtensionID); err != nil {
		t.Fatal(err)
	}
	for _, p := range m.Profiles {
		if persistentExtensionCodeID(p.UserDataDir, e.ExtensionID) != "" {
			t.Fatal("code not removed")
		}
		states, err := m.ExtensionDAO.ListProfileExtensionRuntime(p.ProfileId)
		if err != nil || len(states) != 0 {
			t.Fatalf("runtime not deleted: %v %v", states, err)
		}
	}
	if _, err = os.Stat(storage); err != nil {
		t.Fatalf("local data was deleted: %v", err)
	}
	if _, err = m.InstallExtensionPackageBytes("", "test fixture", data); err != nil {
		t.Fatal(err)
	}
	for _, p := range m.Profiles {
		_, warnings := m.PrepareProfileExtensions(p, "unused-browser", p.UserDataDir, nil)
		if len(warnings) > 0 {
			t.Fatal(warnings)
		}
	}
	if b, err := os.ReadFile(storage); err != nil || string(b) != "keep me" {
		t.Fatalf("local data changed: %q %v", b, err)
	}
}

func TestFindInstalledExtensionDoesNotMatchUnrelatedVersion(t *testing.T) {
	root := t.TempDir()
	otherID := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	path := persistentExtensionCodePath(root, otherID, "1.0.0")
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "manifest.json"), []byte(`{"name":"Unrelated","version":"1.0.0"}`), 0600); err != nil {
		t.Fatal(err)
	}
	id, err := findInstalledRuntimeExtensionID(root, Extension{ExtensionID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "Expected", Version: "1.0.0"})
	if err != nil || id != "" {
		t.Fatalf("matched unrelated extension %s: %v", id, err)
	}
}

func TestExtensionOptionalPermissionsAreNotPreGranted(t *testing.T) {
	permissions := extensionPermissionSnapshot(profileExtensionJSON{"permissions": []any{"storage"}, "optional_permissions": []any{"tabs", "<all_urls>"}, "host_permissions": []any{"https://example.test/*"}})
	if fmt.Sprint(permissions["api"]) != "[storage]" || fmt.Sprint(permissions["explicit_host"]) != "[https://example.test/*]" {
		t.Fatalf("unexpected grants: %v", permissions)
	}
}
