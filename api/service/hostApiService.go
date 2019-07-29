package service

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	HostServiceInterface interface {
	}

	HostService struct {
		Query query.ExecutorInterface
	}
)

var hostServiceInstance *HostService

// NewHostService create a singleton instance of HostService
func NewHostService(queryExecutor query.ExecutorInterface) *HostService {
	if hostServiceInstance == nil {
		hostServiceInstance = &HostService{Query: queryExecutor}
	}
	return hostServiceInstance
}

func (*HostService) GetHostInfo() *model.Host {
	return &model.Host{}
}
