package main

import (
	"context"
	"fmt"
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
	"go.uber.org/zap"
)

var (
	logger, _ = zap.NewProduction(zap.Fields(zap.String("type", "main")))
	shutdowns []func() error
)

func main() {
	ctx := context.Background()
	port := os.Getenv("PORT")
	conn := initConn()
	app := gin.New()
	app.Use(logWithZap(logger), recoveryWithZap(logger, true))
	api.InitRouter(app, conn)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: app,
	}

	shutdown := make(chan struct{})

	go gracefulShutdown(ctx, server, shutdown)

	logger.Info("server starting: http://localhost" + server.Addr)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal("server error", zap.Error(err))
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
		logger.Fatal(err.Error(), zap.Error(err))
	}
	// add to graceful shutdown list.
	shutdowns = append(shutdowns, conn.Close)

	return conn
}

func gracefulShutdown(ctx context.Context, server *http.Server, shutdown chan struct{}) {
	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint

	logger.Info("shutting down server gracefully")

	// stop receiving any request.
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("shutdown error", zap.Error(err))
	}

	// close any other modules.
	for i := range shutdowns {
		_ = shutdowns[i]()
	}

	close(shutdown)
}

func logWithZap(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		latency := time.Since(start)

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			for _, e := range c.Errors.Errors() {
				logger.Error(e)
			}
			return
		}

		logger.Info(c.Request.Method,
			zap.Int("status", c.Writer.Status()),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("latency", latency.String()),
		)
	}
}

func recoveryWithZap(logger *zap.Logger, stack bool) gin.HandlerFunc {
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
					logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
