package launchcode

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ant-chrome/backend/internal/browser"
)

type copyTestStarter struct{}

func (copyTestStarter) StartInstance(profileID string) (*browser.Profile, error) {
	return &browser.Profile{ProfileId: profileID, Running: true, DebugReady: true, DebugPort: 9222}, nil
}

func (copyTestStarter) CopyWithMode(_ string, name string, mode string) (*browser.Profile, error) {
	return &browser.Profile{ProfileId: "new-profile", ProfileName: name, ProxyId: "proxy-1", LaunchCode: "NEW001", Running: false}, nil
}

func TestHandleCopyProfileReturnsNewProfileAndRetainsSource(t *testing.T) {
	mgr := &browser.Manager{Profiles: map[string]*browser.Profile{
		"source-profile": {ProfileId: "source-profile", ProfileName: "source", ProxyId: "proxy-1", LaunchCode: "SRC001"},
	}}
	server := NewLaunchServer(nil, copyTestStarter{}, mgr, 0)
	body := bytes.NewBufferString(`{"name":"replacement","mode":"auto_fingerprint"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles/source-profile/copy", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	server.buildMux().ServeHTTP(recorder, req)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	var payload struct {
		OK              bool   `json:"ok"`
		SourceProfileID string `json:"sourceProfileId"`
		ProfileID       string `json:"profileId"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.OK || payload.SourceProfileID != "source-profile" || payload.ProfileID != "new-profile" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if _, ok := mgr.Profiles["source-profile"]; !ok {
		t.Fatal("source profile was removed")
	}
}

func TestHandleCopyProfileRejectsUnknownMode(t *testing.T) {
	mgr := &browser.Manager{Profiles: map[string]*browser.Profile{
		"source-profile": {ProfileId: "source-profile", ProfileName: "source"},
	}}
	server := NewLaunchServer(nil, copyTestStarter{}, mgr, 0)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles/source-profile/copy", bytes.NewBufferString(`{"mode":"unknown"}`))
	recorder := httptest.NewRecorder()

	server.buildMux().ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
