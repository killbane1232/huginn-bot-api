package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/killbane1232/huginn-bot-api/internal/core"
)

const (
	maxJSONBody   = 1 << 20
	maxUploadBody = 64 << 20
	maxLongPoll   = 30 * time.Second
)

type Server struct {
	core      core.Client
	token     string
	uploadDir string
	logger    *slog.Logger
	mux       *http.ServeMux
}

func New(client core.Client, token, uploadDir string, logger *slog.Logger) (*Server, error) {
	if client == nil {
		return nil, errors.New("core client is required")
	}
	if token == "" {
		return nil, errors.New("BOT_API_TOKEN is required")
	}
	if uploadDir == "" {
		return nil, errors.New("upload directory is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	if err := os.MkdirAll(uploadDir, 0o750); err != nil {
		return nil, fmt.Errorf("create upload directory: %w", err)
	}

	s := &Server{core: client, token: token, uploadDir: uploadDir, logger: logger, mux: http.NewServeMux()}
	s.routes()
	return s, nil
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.health)
	s.mux.Handle("GET /api/v1/me", s.authorize(http.HandlerFunc(s.getMe)))
	s.mux.Handle("GET /api/v1/peers", s.authorize(http.HandlerFunc(s.getPeers)))
	s.mux.Handle("GET /api/v1/messages", s.authorize(http.HandlerFunc(s.getMessages)))
	s.mux.Handle("POST /api/v1/messages", s.authorize(http.HandlerFunc(s.sendMessage)))
	s.mux.Handle("POST /api/v1/files", s.authorize(http.HandlerFunc(s.sendFile)))
	s.mux.Handle("GET /api/v1/updates", s.authorize(http.HandlerFunc(s.getUpdate)))
	s.mux.Handle("GET /api/v1/groups", s.authorize(http.HandlerFunc(s.getGroups)))
	s.mux.Handle("POST /api/v1/groups", s.authorize(http.HandlerFunc(s.createGroup)))
	s.mux.Handle("POST /api/v1/groups/invitations", s.authorize(http.HandlerFunc(s.inviteToGroup)))
	s.mux.Handle("POST /api/v1/messages/read", s.authorize(http.HandlerFunc(s.markRead)))
}

func (s *Server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(response, request)
}

func (s *Server) authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		header := request.Header.Get("Authorization")
		provided := strings.TrimPrefix(header, "Bearer ")
		validFormat := strings.HasPrefix(header, "Bearer ") && provided != ""
		providedHash := sha256.Sum256([]byte(provided))
		expectedHash := sha256.Sum256([]byte(s.token))
		validToken := subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) == 1
		if !validFormat || !validToken {
			response.Header().Set("WWW-Authenticate", "Bearer")
			writeError(response, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(response, request)
	})
}

func (s *Server) health(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, json.RawMessage(`{"status":"ok"}`))
}

func (s *Server) getMe(response http.ResponseWriter, _ *http.Request) {
	result, err := s.core.GetMe()
	s.respondNative(response, result, err)
}

func (s *Server) getPeers(response http.ResponseWriter, request *http.Request) {
	result, err := s.core.GetPeers(request.URL.Query().Get("q"))
	s.respondNative(response, result, err)
}

func (s *Server) getMessages(response http.ResponseWriter, request *http.Request) {
	chatID := request.URL.Query().Get("chat_id")
	if chatID == "" {
		writeError(response, http.StatusBadRequest, "chat_id is required")
		return
	}
	limit, err := queryInt(request, "limit", 50, 1, 200)
	if err != nil {
		writeError(response, http.StatusBadRequest, err.Error())
		return
	}
	offset, err := queryInt(request, "offset", 0, 0, 1_000_000_000)
	if err != nil {
		writeError(response, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.core.GetMessages(chatID, limit, offset)
	s.respondNative(response, result, err)
}

type sendMessageRequest struct {
	ChatID     string `json:"chat_id"`
	Text       string `json:"text"`
	TTLSeconds int    `json:"ttl_seconds,omitempty"`
}

func (s *Server) sendMessage(response http.ResponseWriter, request *http.Request) {
	var body sendMessageRequest
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, err.Error())
		return
	}
	if body.ChatID == "" || body.Text == "" {
		writeError(response, http.StatusBadRequest, "chat_id and text are required")
		return
	}
	if body.TTLSeconds < 0 {
		writeError(response, http.StatusBadRequest, "ttl_seconds must not be negative")
		return
	}
	result, err := s.core.SendMessage(body.ChatID, body.Text, body.TTLSeconds)
	s.respondNative(response, result, err)
}

func (s *Server) sendFile(response http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(response, request.Body, maxUploadBody)
	if err := request.ParseMultipartForm(maxUploadBody); err != nil {
		writeError(response, http.StatusBadRequest, "invalid multipart request")
		return
	}
	chatID := request.FormValue("chat_id")
	if chatID == "" {
		writeError(response, http.StatusBadRequest, "chat_id is required")
		return
	}
	ttl, err := formInt(request, "ttl_seconds", 0)
	if err != nil || ttl < 0 {
		writeError(response, http.StatusBadRequest, "ttl_seconds must be a non-negative integer")
		return
	}
	file, header, err := request.FormFile("file")
	if err != nil {
		writeError(response, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	path, err := s.storeUpload(file, header)
	if err != nil {
		s.logger.Error("store upload", "error", err)
		writeError(response, http.StatusInternalServerError, "failed to store upload")
		return
	}
	result, sendErr := s.core.SendFile(chatID, request.FormValue("caption"), path, ttl)
	if sendErr != nil {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			s.logger.Warn("remove failed upload", "error", err)
		}
	}
	s.respondNative(response, result, sendErr)
}

func (s *Server) storeUpload(source multipart.File, header *multipart.FileHeader) (path string, err error) {
	extension := filepath.Ext(filepath.Base(header.Filename))
	target, err := os.CreateTemp(s.uploadDir, "upload-*"+extension)
	if err != nil {
		return "", err
	}
	path = target.Name()
	defer func() {
		if closeErr := target.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(path)
		}
	}()
	if err = target.Chmod(0o640); err != nil {
		return "", err
	}
	_, err = io.Copy(target, source)
	return path, err
}

func (s *Server) getUpdate(response http.ResponseWriter, request *http.Request) {
	timeoutSeconds, err := queryInt(request, "timeout", 25, 0, int(maxLongPoll/time.Second))
	if err != nil {
		writeError(response, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.core.GetUpdate(time.Duration(timeoutSeconds) * time.Second)
	if err != nil {
		s.respondNative(response, nil, err)
		return
	}
	if len(result) == 0 {
		writeJSON(response, http.StatusOK, json.RawMessage(`{"event":null}`))
		return
	}
	writeJSON(response, http.StatusOK, json.RawMessage(`{"event":`+string(result)+`}`))
}

func (s *Server) getGroups(response http.ResponseWriter, _ *http.Request) {
	result, err := s.core.GetGroups()
	s.respondNative(response, result, err)
}

func (s *Server) createGroup(response http.ResponseWriter, request *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, err.Error())
		return
	}
	if body.Name == "" {
		writeError(response, http.StatusBadRequest, "name is required")
		return
	}
	result, err := s.core.CreateGroup(body.Name)
	s.respondNative(response, result, err)
}

func (s *Server) inviteToGroup(response http.ResponseWriter, request *http.Request) {
	var body struct {
		GroupID string `json:"group_id"`
		UserID  string `json:"user_id"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, err.Error())
		return
	}
	if body.GroupID == "" || body.UserID == "" {
		writeError(response, http.StatusBadRequest, "group_id and user_id are required")
		return
	}
	result, err := s.core.InviteToGroup(body.GroupID, body.UserID)
	s.respondNative(response, result, err)
}

func (s *Server) markRead(response http.ResponseWriter, request *http.Request) {
	var body struct {
		MessageID string `json:"message_id"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, err.Error())
		return
	}
	if body.MessageID == "" {
		writeError(response, http.StatusBadRequest, "message_id is required")
		return
	}
	result, err := s.core.MarkRead(body.MessageID)
	s.respondNative(response, result, err)
}

func (s *Server) respondNative(response http.ResponseWriter, result json.RawMessage, err error) {
	if err != nil {
		s.logger.Warn("Huginn core request failed", "error", err)
		writeError(response, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(response, http.StatusOK, result)
}

func decodeJSON(response http.ResponseWriter, request *http.Request, target any) error {
	request.Body = http.MaxBytesReader(response, request.Body, maxJSONBody)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func queryInt(request *http.Request, name string, defaultValue, minValue, maxValue int) (int, error) {
	value := request.URL.Query().Get(name)
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < minValue || parsed > maxValue {
		return 0, fmt.Errorf("%s must be an integer between %d and %d", name, minValue, maxValue)
	}
	return parsed, nil
}

func formInt(request *http.Request, name string, defaultValue int) (int, error) {
	value := request.FormValue(name)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(value)
}

func writeJSON(response http.ResponseWriter, status int, body json.RawMessage) {
	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("X-Content-Type-Options", "nosniff")
	response.WriteHeader(status)
	if len(body) == 0 {
		body = json.RawMessage("null")
	}
	_, _ = response.Write(append(body, '\n'))
}

func writeError(response http.ResponseWriter, status int, message string) {
	body, _ := json.Marshal(map[string]string{"error": message})
	writeJSON(response, status, body)
}
