package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/DulioShortener/Shortener/backend/internal/config"
	"github.com/DulioShortener/Shortener/backend/internal/repository/sqlite"
	"github.com/DulioShortener/Shortener/backend/internal/security/password"
	"github.com/DulioShortener/Shortener/backend/internal/security/shortcode"
	securitysonyflake "github.com/DulioShortener/Shortener/backend/internal/security/sonyflake"
	"github.com/DulioShortener/Shortener/backend/internal/security/token"
	"github.com/DulioShortener/Shortener/backend/internal/service"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/handler"
	authmiddleware "github.com/DulioShortener/Shortener/backend/internal/transport/http/middleware"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/router"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/validation"
	"github.com/DulioShortener/Shortener/database"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Server struct {
	echo            *echo.Echo
	database        *sql.DB
	address         string
	certificateFile string
	privateKeyFile  string
	shutdownTimeout time.Duration
}

const (
	tlsAddress         = ":443"
	tlsCertificateFile = "/data/tls/origin.pem"
	tlsPrivateKeyFile  = "/data/tls/origin.key"
)

func New(ctx context.Context, cfg config.Config) (*Server, error) {
	db, err := database.Open(ctx, cfg.DatabasePath)
	if err != nil {
		return nil, err
	}

	idGenerator, err := securitysonyflake.NewGenerator(cfg.SonyflakeMachineID)
	if err != nil {
		db.Close()
		return nil, err
	}
	requestValidator, err := validation.New()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create request validator: %w", err)
	}

	userRepository := sqlite.NewUserRepository(db)
	sessionRepository := sqlite.NewSessionRepository(db)
	linkRepository := sqlite.NewLinkRepository(db)
	passwordHasher := password.NewArgon2id()
	tokenGenerator := token.NewGenerator()
	clock := service.SystemClock{}

	userService := service.NewUserService(userRepository, passwordHasher, idGenerator, clock)
	authService, err := service.NewAuthService(
		userRepository, sessionRepository, passwordHasher, tokenGenerator,
		idGenerator, clock, cfg.TokenTTL,
	)
	if err != nil {
		db.Close()
		return nil, err
	}
	linkService := service.NewLinkService(
		linkRepository, idGenerator, shortcode.NewGenerator(), clock,
	)

	e := echo.New()
	e.Validator = requestValidator
	e.HTTPErrorHandler = handler.Error
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit(64 * 1024))
	e.Use(middleware.Secure())
	e.Use(middleware.CORS("*"))

	shortURLBase := "https://" + cfg.BaseHostname
	applicationURL := "https://app." + cfg.BaseHostname + "/"
	authentication := authmiddleware.NewAuthentication(authService)
	router.Register(e, router.Dependencies{
		Root:  handler.NewRoot(applicationURL),
		Users: handler.NewUser(userService), Auth: handler.NewAuth(authService),
		Links: handler.NewLink(linkService, shortURLBase), Authentication: authentication,
	})

	return &Server{
		echo: e, database: db, address: tlsAddress,
		certificateFile: tlsCertificateFile, privateKeyFile: tlsPrivateKeyFile,
		shutdownTimeout: cfg.ShutdownTimeout,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	certificate, err := os.ReadFile(s.certificateFile)
	if err != nil {
		return fmt.Errorf("run HTTPS server: read TLS certificate: %w", err)
	}
	privateKey, err := os.ReadFile(s.privateKeyFile)
	if err != nil {
		return fmt.Errorf("run HTTPS server: read TLS private key: %w", err)
	}

	startConfig := echo.StartConfig{
		Address:         s.address,
		GracefulTimeout: s.shutdownTimeout,
	}
	if err := startConfig.StartTLS(ctx, s.echo, certificate, privateKey); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("run HTTPS server: %w", err)
	}
	return nil
}

func (s *Server) Close() error {
	return s.database.Close()
}

func (s *Server) Handler() http.Handler {
	return s.echo
}
