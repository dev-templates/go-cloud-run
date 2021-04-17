package handler

import "go.uber.org/zap"

var logger, _ = zap.NewProduction(zap.Fields(zap.String("type", "handler")))
