package api

import (
	"net/http"
	"time"

	"github.com/dev-templates/go-cloud-run/api/handler"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"golang.org/x/time/rate"
)

func Throttle(every time.Duration, maxBurstSize int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(every), maxBurstSize)
	return func(c *gin.Context) {
		if limiter.Allow() {
			c.Next()
			return
		}
		c.AbortWithStatus(http.StatusTooManyRequests)
	}
}

func InitRouter(r *gin.Engine, conn *sqlx.DB) {
	echo := handler.NewEcho(conn)
	api := r.Group("api", Throttle(time.Second, 60))
	{
		api.GET("echo", echo.Echo)
	}

	r.GET("/", echo.Echo)
}
