package model

import "github.com/zoobc/zoobc-core/common/contract"

var (
	mainchainDerivedRepositoriesInstance  *derivedRepositories
	spinechainDerivedRepositoriesInstance *derivedRepositories
)

type derivedRepositories struct {
	Repositories []contract.DerivedRepository
}

// DerivedRepositories initialize or get instance of derived repository list
// by supplying chain type

// fetch all derived repositories
func (dr derivedRepositories) GetDerivedRepositories() []contract.DerivedRepository {
	return dr.Repositories
}

// register repository as derived repository
func (dr *derivedRepositories) RegisterRepository(repository contract.DerivedRepository) {
	dr.Repositories = append(dr.Repositories, repository)
}
