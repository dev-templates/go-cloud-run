package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dev-templates/go-cloud-run/api"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var (
	log, _    = zap.NewProduction(zap.Fields(zap.String("type", "main")))
	shutdowns []func() error
)

func main() {
	ctx := context.Background()
	port := os.Getenv("PORT")
	conn := initConn()
	app := gin.Default()
	api.InitRouter(app, conn)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: app,
	}

	shutdown := make(chan struct{})

	go gracefulShutdown(ctx, server, shutdown)

	log.Info("server starting: http://localhost" + server.Addr)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("server error", zap.Error(err))
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
		log.Fatal(err.Error(), zap.Error(err))
	}
	// add to graceful shutdown list.
	shutdowns = append(shutdowns, conn.Close)

	return conn
}

func gracefulShutdown(ctx context.Context, server *http.Server, shutdown chan struct{}) {
	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint

	log.Info("shutting down server gracefully")

	// stop receiving any request.
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("shutdown error", zap.Error(err))
	}

	// close any other modules.
	for i := range shutdowns {
		_ = shutdowns[i]()
	}

	close(shutdown)
}
