package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type Echo struct {
	conn *sqlx.DB
}

func NewEcho(conn *sqlx.DB) *Echo {
	return &Echo{
		conn: conn,
	}
}

func (e *Echo) Echo(c *gin.Context) {
	c.String(http.StatusOK, c.ClientIP())
}
