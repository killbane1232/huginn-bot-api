package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fakeCore struct {
	messageChat string
	messageText string
	messageTTL  int
	updateWait  time.Duration
	filePath    string
	err         error
}

func (f *fakeCore) Close() error { return nil }
func (f *fakeCore) GetMe() (json.RawMessage, error) {
	return json.RawMessage(`{"id":"bot-id","username":"test-bot"}`), f.err
}
func (f *fakeCore) GetPeers(string) (json.RawMessage, error) {
	return json.RawMessage(`[]`), f.err
}
func (f *fakeCore) GetMessages(string, int, int) (json.RawMessage, error) {
	return json.RawMessage(`[]`), f.err
}
func (f *fakeCore) SendMessage(chatID, text string, ttl int) (json.RawMessage, error) {
	f.messageChat, f.messageText, f.messageTTL = chatID, text, ttl
	return json.RawMessage(`{"status":"ok"}`), f.err
}
func (f *fakeCore) SendFile(_, _, path string, _ int) (json.RawMessage, error) {
	f.filePath = path
	return json.RawMessage(`{"status":"ok"}`), f.err
}
func (f *fakeCore) GetUpdate(timeout time.Duration) (json.RawMessage, error) {
	f.updateWait = timeout
	return json.RawMessage(`{"type":"message","data":{"id":"m1"}}`), f.err
}
func (f *fakeCore) GetGroups() (json.RawMessage, error) {
	return json.RawMessage(`[]`), f.err
}
func (f *fakeCore) CreateGroup(string) (json.RawMessage, error) {
	return json.RawMessage(`{"uid":"g1"}`), f.err
}
func (f *fakeCore) InviteToGroup(string, string) (json.RawMessage, error) {
	return json.RawMessage(`{"status":"ok"}`), f.err
}
func (f *fakeCore) MarkRead(string) (json.RawMessage, error) {
	return json.RawMessage(`{"status":"ok"}`), f.err
}

func newTestServer(t *testing.T, client coreForTest) http.Handler {
	t.Helper()
	server, err := New(client, "secret-token", t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	return server
}

type coreForTest interface {
	Close() error
	GetMe() (json.RawMessage, error)
	GetPeers(string) (json.RawMessage, error)
	GetMessages(string, int, int) (json.RawMessage, error)
	SendMessage(string, string, int) (json.RawMessage, error)
	SendFile(string, string, string, int) (json.RawMessage, error)
	GetUpdate(time.Duration) (json.RawMessage, error)
	GetGroups() (json.RawMessage, error)
	CreateGroup(string) (json.RawMessage, error)
	InviteToGroup(string, string) (json.RawMessage, error)
	MarkRead(string) (json.RawMessage, error)
}

func authorized(request *http.Request) *http.Request {
	request.Header.Set("Authorization", "Bearer secret-token")
	return request
}

func TestRequiresBearerToken(t *testing.T) {
	response := httptest.NewRecorder()
	newTestServer(t, &fakeCore{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestSendMessage(t *testing.T) {
	client := &fakeCore{}
	response := httptest.NewRecorder()
	request := authorized(httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBufferString(`{"chat_id":"peer:signature","text":"hello","ttl_seconds":120}`)))
	request.Header.Set("Content-Type", "application/json")
	newTestServer(t, client).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if client.messageChat != "peer:signature" || client.messageText != "hello" || client.messageTTL != 120 {
		t.Fatalf("unexpected native call: %#v", client)
	}
}

func TestRejectsUnknownJSONFields(t *testing.T) {
	response := httptest.NewRecorder()
	request := authorized(httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBufferString(`{"chat_id":"peer","text":"hello","unknown":true}`)))
	newTestServer(t, &fakeCore{}).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestLongPollingWrapsEvent(t *testing.T) {
	client := &fakeCore{}
	response := httptest.NewRecorder()
	request := authorized(httptest.NewRequest(http.MethodGet, "/api/v1/updates?timeout=7", nil))
	newTestServer(t, client).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if client.updateWait != 7*time.Second {
		t.Fatalf("timeout = %s, want 7s", client.updateWait)
	}
	if got := response.Body.String(); got != "{\"event\":{\"type\":\"message\",\"data\":{\"id\":\"m1\"}}}\n" {
		t.Fatalf("body = %s", got)
	}
}

func TestNativeErrorBecomesBadGateway(t *testing.T) {
	client := &fakeCore{err: errors.New("delivery failed")}
	response := httptest.NewRecorder()
	request := authorized(httptest.NewRequest(http.MethodGet, "/api/v1/me", nil))
	newTestServer(t, client).ServeHTTP(response, request)
	if response.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadGateway)
	}
}

func TestFileUploadUsesPrivatePersistentFile(t *testing.T) {
	client := &fakeCore{}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("chat_id", "peer"); err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile("file", "picture.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("image-data")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	request := authorized(httptest.NewRequest(http.MethodPost, "/api/v1/files", body))
	request.Header.Set("Content-Type", writer.FormDataContentType())
	newTestServer(t, client).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if filepath.Ext(client.filePath) != ".png" {
		t.Fatalf("stored path = %q", client.filePath)
	}
	info, err := os.Stat(client.filePath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("permissions = %o, want 640", info.Mode().Perm())
	}
}
