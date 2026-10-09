package config

import "testing"

func TestLoadBaseHostname(t *testing.T) {
	t.Setenv("DULIO_BASE_HOSTNAME", "3DreamStudio.COM.BR")
	t.Setenv("AUTH_TOKEN_TTL", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")
	t.Setenv("SONYFLAKE_MACHINE_ID", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseHostname != "3dreamstudio.com.br" {
		t.Fatalf("BaseHostname=%q", cfg.BaseHostname)
	}
}

func TestLoadRejectsInvalidBaseHostname(t *testing.T) {
	t.Setenv("AUTH_TOKEN_TTL", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")
	t.Setenv("SONYFLAKE_MACHINE_ID", "")

	for _, value := range []string{
		"https://3dreamstudio.com.br",
		"3dreamstudio.com.br:443",
		"3dreamstudio.com.br/path",
		"bad..example",
		"-bad.example",
		"bad-.example",
	} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("DULIO_BASE_HOSTNAME", value)
			if _, err := Load(); err == nil {
				t.Fatalf("Load accepted invalid hostname %q", value)
			}
		})
	}
}
