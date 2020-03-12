package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/common/model"
)

type HealthCheckHandler struct {
}

func (hc *HealthCheckHandler) HealthCheck(context.Context, *model.Empty) (*model.HealthCheckResponse, error) {
	return &model.HealthCheckResponse{
		Reply: "pong",
	}, nil
}
