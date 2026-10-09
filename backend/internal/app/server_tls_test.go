package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

func TestRunLoadsTLSCertificate(t *testing.T) {
	missingCertificate := t.TempDir() + "/missing.pem"
	server := &Server{
		echo:            echo.New(),
		address:         "127.0.0.1:0",
		certificateFile: missingCertificate,
		privateKeyFile:  missingCertificate,
		shutdownTimeout: time.Second,
	}

	err := server.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "run HTTPS server") {
		t.Fatalf("expected TLS startup error, got %v", err)
	}
}
