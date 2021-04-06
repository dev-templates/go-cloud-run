package api

import (
	"fmt"
	"net/http"
	"time"

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

		_ = c.Error(fmt.Errorf("TooManyRequests"))
		c.AbortWithStatus(http.StatusTooManyRequests)
	}
}

func InitRouter(r *gin.Engine, conn *sqlx.DB) {
	api := r.Group("api", Throttle(time.Second, 60))
	{
		api.BasePath()
	}
}
