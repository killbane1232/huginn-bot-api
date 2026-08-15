package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/killbane1232/huginn-bot-api/internal/api"
	"github.com/killbane1232/huginn-bot-api/internal/config"
	"github.com/killbane1232/huginn-bot-api/internal/core"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("Bot API stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.FromEnvironment()
	if err != nil {
		return err
	}
	client, err := core.Open(core.Options{
		LibraryPath: cfg.LibraryPath,
		Username:    cfg.Username,
		MuninnAddr:  cfg.MuninnAddr,
		Database:    cfg.Database,
		ChunkTTL:    cfg.ChunkTTL,
		TURNAddr:    cfg.TURNAddr,
		TURNUser:    cfg.TURNUser,
		TURNPass:    cfg.TURNPass,
	})
	if err != nil {
		return err
	}
	defer client.Close()

	handler, err := api.New(client, cfg.Token, cfg.UploadDir, logger)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	serverError := make(chan error, 1)
	go func() {
		logger.Info("Bot API listening", "address", cfg.Address)
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			return err
		}
		return nil
	}
}
