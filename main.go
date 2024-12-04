package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	port := getEnv("PORT", "8080")

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%v", port),
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		Handler:           routing(),
	}
	go func() {
		slog.InfoContext(ctx, "http server started", slog.String("addr", srv.Addr))
		if httpErr := srv.ListenAndServe(); httpErr != nil {
			switch {
			case errors.Is(httpErr, http.ErrServerClosed):
				slog.InfoContext(ctx, "http server stopped", slog.String("addr", srv.Addr))
			default:
				slog.ErrorContext(ctx, httpErr.Error(), slog.String("addr", srv.Addr))
				os.Exit(1)
			}
		}
	}()
	defer func() { _ = srv.Close() }()

	wait()
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func handler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"hello": "world"})
}

func routing() http.Handler {
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.AllowAll())

	r.GET("/foo", handler)

	return r
}

func wait() {
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
	<-termChan
}
