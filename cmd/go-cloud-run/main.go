package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/dev-templates/go-cloud-run/api"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	logger    = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	shutdowns []func() error
)

func main() {
	slog.SetDefault(logger)
	ctx := context.Background()
	port := os.Getenv("PORT")
	conn := initConn()
	app := gin.New()
	app.Use(logWithZap(), recoveryWithZap(true))
	api.InitRouter(app, conn)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: app,
	}

	shutdown := make(chan struct{})

	go gracefulShutdown(ctx, server, shutdown)

	slog.Info("server starting: http://localhost" + server.Addr)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", slog.Any("err", err))
	}
	<-shutdown
}

func initConn() *sqlx.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DATABASE"),
	)

	conn, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		slog.Error(err.Error(), slog.Any("err", err))
		os.Exit(1)
	}
	// add to graceful shutdown list.
	shutdowns = append(shutdowns, conn.Close)

	return conn
}

func gracefulShutdown(ctx context.Context, server *http.Server, shutdown chan struct{}) {
	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint

	slog.Info("shutting down server gracefully")

	// stop receiving any request.
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", slog.Any("err", err))
		os.Exit(1)
	}

	// close any other modules.
	for i := range shutdowns {
		_ = shutdowns[i]()
	}

	close(shutdown)
}

func logWithZap() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		latency := time.Since(start)

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			for _, e := range c.Errors.Errors() {
				slog.Error(e)
			}
			return
		}

		slog.Info(c.Request.Method,
			slog.Int("status", c.Writer.Status()),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.String("user-agent", c.Request.UserAgent()),
			slog.String("latency", latency.String()),
		)
	}
}

func recoveryWithZap(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					slog.Error(c.Request.URL.Path,
						slog.Any("error", err),
						slog.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					slog.Error("[Recovery from panic]",
						slog.Any("error", err),
						slog.String("request", string(httpRequest)),
						slog.String("stack", string(debug.Stack())),
					)
				} else {
					slog.Error("[Recovery from panic]",
						slog.Any("error", err),
						slog.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
