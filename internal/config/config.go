package config

import (
	"errors"
	"os"
	"path/filepath"
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
		MuninnAddr:  value("MUNINN_ADDR", "https://muninn.evil-bread.ru"),
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
	return config, nil
}

func value(name, fallback string) string {
	if result := os.Getenv(name); result != "" {
		return result
	}
	return fallback
}
