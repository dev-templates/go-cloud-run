package handler

import "go.uber.org/zap"

var log, _ = zap.NewProduction(zap.Fields(zap.String("type", "handler")))
