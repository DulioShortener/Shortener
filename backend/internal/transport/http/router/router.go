package router

import (
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/handler"
	authmiddleware "github.com/DulioShortener/Shortener/backend/internal/transport/http/middleware"
	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

type Dependencies struct {
	Root           *handler.Root
	Users          *handler.User
	Auth           *handler.Auth
	Links          *handler.Link
	Authentication *authmiddleware.Authentication
}

func Register(e *echo.Echo, deps Dependencies) {
	authRateLimit := echomiddleware.RateLimiter(echomiddleware.NewRateLimiterMemoryStore(5))
	e.GET("/", deps.Root.Redirect)
	e.HEAD("/", deps.Root.Redirect)

	api := e.Group("/api/v1")
	api.POST("/users", deps.Users.Create, authRateLimit)
	api.GET("/users/me", deps.Users.Me, deps.Authentication.Require)
	api.POST("/links", deps.Links.Create, deps.Authentication.Require)
	api.GET("/links", deps.Links.List, deps.Authentication.Require)
	api.DELETE("/links/:id", deps.Links.Delete, deps.Authentication.Require)

	auth := api.Group("/auth")
	auth.POST("/login", deps.Auth.Login, authRateLimit)
	auth.POST("/logout", deps.Auth.Logout, deps.Authentication.Require)

	e.GET("/r/:code", deps.Links.Resolve)
}
