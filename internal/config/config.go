package config

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Address     string
	Token       string
	LibraryPath string
	Username    string
	MuninnAddr  string
	Database    string
	ChunkTTL    string
	TURNAddr    string
	TURNUser    string
	TURNPass    string
	UploadDir   string
}

func FromEnvironment() (Config, error) {
	config := Config{
		Address:     value("BOT_API_ADDR", ":8081"),
		Token:       os.Getenv("BOT_API_TOKEN"),
		LibraryPath: value("HUGINN_CORE_LIBRARY", filepath.Join("build", "libhuginn_messenger.so")),
		Username:    os.Getenv("HUGINN_USERNAME"),
		MuninnAddr:  strings.TrimSpace(os.Getenv("MUNINN_ADDR")),
		Database:    value("HUGINN_DB_PATH", filepath.Join("data", "huginn.db")),
		ChunkTTL:    value("HUGINN_CHUNK_TTL", "1w"),
		TURNAddr:    os.Getenv("HUGINN_TURN_ADDR"),
		TURNUser:    os.Getenv("HUGINN_TURN_USER"),
		TURNPass:    os.Getenv("HUGINN_TURN_PASS"),
		UploadDir:   value("BOT_API_UPLOAD_DIR", filepath.Join("data", "uploads")),
	}
	if config.Token == "" {
		return Config{}, errors.New("BOT_API_TOKEN is required")
	}
	if config.Username == "" {
		return Config{}, errors.New("HUGINN_USERNAME is required")
	}
	if config.MuninnAddr == "" {
		return Config{}, errors.New("MUNINN_ADDR is required")
	}
	addr, err := url.Parse(config.MuninnAddr)
	if err != nil || (addr.Scheme != "http" && addr.Scheme != "https") || addr.Hostname() == "" || addr.RawQuery != "" || addr.Fragment != "" {
		return Config{}, errors.New("MUNINN_ADDR must be an http(s) URL without query or fragment")
	}
	config.MuninnAddr = strings.TrimRight(config.MuninnAddr, "/")
	return config, nil
}

func value(name, fallback string) string {
	if result := os.Getenv(name); result != "" {
		return result
	}
	return fallback
}
