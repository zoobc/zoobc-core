package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/util"
)

type (
	BlockSpinePublicKeyServiceInterface interface {
		BuildSpinePublicKeysFromNodeRegistry(
			fromTimestamp,
			toTimestamp int64,
			spineBlockHeight uint32,
		) (spinePublicKeys []*model.SpinePublicKey, err error)
		GetSpinePublicKeysByBlockHeight(height uint32) (spinePublicKeys []*model.SpinePublicKey, err error)
		GetValidSpinePublicKeyByBlockHeightInterval(
			fromHeight, toHeight uint32,
		) (
			[]*model.SpinePublicKey, error,
		)
		InsertSpinePublicKeys(block *model.Block) error
	}

	BlockSpinePublicKeyService struct {
		Signature             crypto.SignatureInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Logger                *log.Logger
	}
)

func NewBlockSpinePublicKeyService(
	signature crypto.SignatureInterface,
	queryExecutor query.ExecutorInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	spinePublicKeyQuery query.SpinePublicKeyQueryInterface,
	logger *log.Logger,
) *BlockSpinePublicKeyService {
	return &BlockSpinePublicKeyService{
		Signature:             signature,
		QueryExecutor:         queryExecutor,
		NodeRegistrationQuery: nodeRegistrationQuery,
		SpinePublicKeyQuery:   spinePublicKeyQuery,
		Logger:                logger,
	}
}

// GetValidSpinePublicKeyByBlockHeightInterval return the spine_public_key rows that were valid
func (bsf *BlockSpinePublicKeyService) GetValidSpinePublicKeyByBlockHeightInterval(
	fromHeight, toHeight uint32,
) (
	[]*model.SpinePublicKey, error,
) {
	var validSpinePublicKeys []*model.SpinePublicKey
	// get all registered nodes with participation score > 0
	rows, err := bsf.QueryExecutor.ExecuteSelect(bsf.SpinePublicKeyQuery.GetValidSpinePublicKeysByHeightInterval(fromHeight, toHeight), false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	validSpinePublicKeys, err = bsf.SpinePublicKeyQuery.BuildModel(validSpinePublicKeys, rows)
	return validSpinePublicKeys, err
}

func (bsf *BlockSpinePublicKeyService) GetSpinePublicKeysByBlockHeight(height uint32) (spinePublicKeys []*model.SpinePublicKey, err error) {
	rows, err := bsf.QueryExecutor.ExecuteSelect(bsf.SpinePublicKeyQuery.GetSpinePublicKeysByBlockHeight(height), false)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	spinePublicKeys, err = bsf.SpinePublicKeyQuery.BuildModel(spinePublicKeys, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return spinePublicKeys, nil
}

// BuildSpinePublicKeysFromNodeRegistry build the list of spine public keys from the node registry
func (bsf *BlockSpinePublicKeyService) BuildSpinePublicKeysFromNodeRegistry(
	fromTimestamp,
	toTimestamp int64,
	spineHeight uint32,
) (spinePublicKeys []*model.SpinePublicKey, err error) {
	var (
		nodeRegistrations []*model.NodeRegistration
	)
	qry := bsf.NodeRegistrationQuery.GetNodeRegistrationsByBlockTimestampInterval(fromTimestamp, toTimestamp)
	rows, err := bsf.QueryExecutor.ExecuteSelect(
		qry,
		false,
	)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	nodeRegistrations, err = bsf.NodeRegistrationQuery.BuildModel(nodeRegistrations, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	spinePublicKeys = make([]*model.SpinePublicKey, 0)
	for _, nr := range nodeRegistrations {
		spinePublicKey := &model.SpinePublicKey{
			NodePublicKey:   nr.NodePublicKey,
			NodeID:          nr.NodeID,
			PublicKeyAction: util.GetAddRemoveSpineKeyAction(nr.RegistrationStatus),
			MainBlockHeight: nr.Height, // (node registration) transaction's height
			Height:          spineHeight,
			Latest:          true,
		}
		spinePublicKeys = append(spinePublicKeys, spinePublicKey)
	}
	return spinePublicKeys, nil
}

// InsertSpinePublicKeys insert all spine block publicKeys into spinePublicKey table
// Note: at this stage the spine pub keys have already been parsed into their model struct
func (bsf *BlockSpinePublicKeyService) InsertSpinePublicKeys(block *model.Block) error {
	queries := make([][]interface{}, 0)
	for _, spinePublicKey := range block.SpinePublicKeys {
		insertSpkQry := bsf.SpinePublicKeyQuery.InsertSpinePublicKey(spinePublicKey)
		queries = append(queries, insertSpkQry...)
	}
	if err := bsf.QueryExecutor.ExecuteTransactions(queries); err != nil {
		return err
	}
	return nil
}
