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
	Reviews        *handler.Review
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
	api.GET("/reviews", deps.Reviews.List, deps.Authentication.Optional)
	api.POST("/reviews", deps.Reviews.Create, deps.Authentication.Require)
	api.GET("/reviews/me", deps.Reviews.Mine, deps.Authentication.Require)
	api.GET("/reviews/:id", deps.Reviews.Get, deps.Authentication.Optional)
	api.DELETE("/reviews/:id", deps.Reviews.Delete, deps.Authentication.Require)
	api.GET("/reviews/:id/replies", deps.Reviews.ListReplies, deps.Authentication.Require)
	api.POST("/reviews/:id/replies", deps.Reviews.CreateReply, deps.Authentication.Require)

	auth := api.Group("/auth")
	auth.POST("/login", deps.Auth.Login, authRateLimit)
	auth.POST("/logout", deps.Auth.Logout, deps.Authentication.Require)

	e.GET("/r/:code", deps.Links.Resolve)
}
