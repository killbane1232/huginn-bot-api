package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestBotPeerFlagIsAlwaysVeryThick(t *testing.T) {
	if botPeerFlag != "very_thick" {
		t.Fatalf("botPeerFlag = %q, want very_thick", botPeerFlag)
	}
}

func TestNativeSendReturnsPreparationErrorsAndCommitsBeforeSuccess(t *testing.T) {
	library := os.Getenv("HUGINN_CORE_TEST_LIBRARY")
	if library == "" {
		t.Skip("set HUGINN_CORE_TEST_LIBRARY to a rebuilt core")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/peers" {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	client, err := Open(Options{LibraryPath: library, Username: "test-bot", MuninnAddr: srv.URL, Database: filepath.Join(t.TempDir(), "test.db"), ChunkTTL: "1w"})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if _, err := client.SendMessage("unknown-recipient", "draft", 0); err == nil {
		t.Fatal("unknown recipient reported success")
	}
	raw, err := client.CreateGroup("delivery regression")
	if err != nil {
		t.Fatal(err)
	}
	var group struct {
		UID string `json:"uid"`
	}
	if err := json.Unmarshal(raw, &group); err != nil || group.UID == "" {
		t.Fatal("group creation failed")
	}
	if _, err := client.SendFile(group.UID, "attachment", filepath.Join(t.TempDir(), "missing.txt"), 0); err == nil {
		t.Fatal("missing attachment reported success")
	}
	if _, err := client.SendMessage(group.UID, "durably queued", 0); err != nil {
		t.Fatal(err)
	}
	raw, err = client.GetMessages(group.UID, 64, 0)
	if err != nil {
		t.Fatal(err)
	}
	var history []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &history); err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Text != "durably queued" {
		t.Fatal("send returned success before committing message")
	}
}
