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
	t.Setenv("MUNINN_ADDR", "http://muninn.internal:8080")
	config, err := FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if config.Address != ":8081" || config.MuninnAddr != "http://muninn.internal:8080" {
		t.Fatal("unexpected default bind address or Muninn ENV address")
	}
}

func TestMuninnAddressFromEnvironment(t *testing.T) {
	t.Setenv("BOT_API_TOKEN", "token")
	t.Setenv("HUGINN_USERNAME", "weather-bot")
	for _, value := range []string{"", "   ", "muninn.internal:8080", "ftp://muninn.internal", "https://", "http://muninn.internal?x=1", "http://muninn.internal#fragment"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("MUNINN_ADDR", value)
			if _, err := FromEnvironment(); err == nil {
				t.Fatal("expected invalid or missing MUNINN_ADDR error")
			}
		})
	}
	t.Setenv("MUNINN_ADDR", " https://muninn.internal/directory/ ")
	cfg, err := FromEnvironment()
	if err != nil || cfg.MuninnAddr != "https://muninn.internal/directory" {
		t.Fatalf("normalized Muninn address = %q, err = %v", cfg.MuninnAddr, err)
	}
}
