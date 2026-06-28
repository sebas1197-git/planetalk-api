// Command api is the entrypoint of the Planetalk backend.
//
// Keep this file thin: it should only WIRE things together. The actual logic
// lives in the internal/* packages.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/auth"
	"github.com/sebas1197-git/planetalk/internal/config"
	"github.com/sebas1197-git/planetalk/internal/db"
	"github.com/sebas1197-git/planetalk/internal/user"
	"github.com/sebas1197-git/planetalk/pkg/twilio"
)

func main() {
	// 1. Load configuration. Check the error BEFORE using cfg.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// 2. Run DB migrations (creates/updates tables to the latest version).
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	// 3. Open the PostgreSQL connection pool. Close it when the program exits.
	pool, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}
	defer pool.Close()

	// 4. Build shared dependencies (tokens + SMS sender).
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	sms := twilio.New(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioFrom)

	// 5. Build the Gin router and register routes.
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Health stays UNVERSIONED at the root (for load balancers / monitoring).
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// All app endpoints live under /api/v1 so we can ship /api/v2 later
	// without breaking apps already calling v1.
	v1 := r.Group("/api/v1")

	// Auth module (Step 3): /api/v1/auth/otp/request, .../verify, .../refresh
	authSvc := auth.NewService(auth.NewRepository(pool), tokens, sms)
	auth.RegisterRoutes(v1, auth.NewHandler(authSvc))

	// User module (Step 4): /api/v1/me, /users/:id, /interests, /friends/*
	userSvc := user.NewService(user.NewRepository(pool))
	user.RegisterRoutes(v1, user.NewHandler(userSvc), tokens)

	// 6. Wrap the router in an http.Server so we can shut it down cleanly.
	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: r,
	}

	// 7. Start the server in the background so main() can wait for a stop signal.
	go func() {
		log.Printf("Planetalk API listening on http://localhost:%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 8. Wait for Ctrl+C / stop signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// 9. Give in-flight requests up to 5 seconds to finish, then exit.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exiting")
}
