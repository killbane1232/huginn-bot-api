package config

import "testing"

func TestFromEnvironmentRequiresCredentials(t *testing.T) {
	t.Setenv("BOT_API_TOKEN", "")
	t.Setenv("HUGINN_USERNAME", "")
	if _, err := FromEnvironment(); err == nil {
		t.Fatal("expected missing token error")
	}
}

func TestFromEnvironmentDefaults(t *testing.T) {
	t.Setenv("BOT_API_TOKEN", "token")
	t.Setenv("HUGINN_USERNAME", "weather-bot")
	t.Setenv("BOT_API_ADDR", "")
	t.Setenv("MUNINN_ADDR", "")
	config, err := FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if config.Address != ":8081" || config.MuninnAddr != "https://muninn.evil-bread.ru" {
		t.Fatalf("unexpected defaults: %#v", config)
	}
}
