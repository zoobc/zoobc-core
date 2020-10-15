package service

import (
	"database/sql"
	"encoding/hex"

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	MultisigServiceInterface interface {
		GetPendingTransactions(
			param *model.GetPendingTransactionsRequest,
		) (*model.GetPendingTransactionsResponse, error)
		GetPendingTransactionsByHeight(
			fromHeight, toHeight uint32,
		) ([]*model.PendingTransaction, error)
		GetPendingTransactionDetailByTransactionHash(
			param *model.GetPendingTransactionDetailByTransactionHashRequest,
		) (*model.GetPendingTransactionDetailByTransactionHashResponse, error)
		GetMultisignatureInfo(
			param *model.GetMultisignatureInfoRequest,
		) (*model.GetMultisignatureInfoResponse, error)
		GetMultisigAddressByParticipantAddress(
			param *model.GetMultisigAddressByParticipantAddressRequest,
		) (*model.GetMultisigAddressByParticipantAddressResponse, error)
		GetMultisigAddressesByBlockHeightRange(
			param *model.GetMultisigAddressesByBlockHeightRangeRequest,
		) (*model.GetMultisigAddressesByBlockHeightRangeResponse, error)
		GetParticipantsByMultisigAddresses(
			param *model.GetParticipantsByMultisigAddressesRequest,
		) (*model.GetParticipantsByMultisigAddressesResponse, error)
	}

	MultisigService struct {
		Executor                       query.ExecutorInterface
		BlockService                   coreService.BlockServiceInterface
		PendingTransactionQuery        query.PendingTransactionQueryInterface
		PendingSignatureQuery          query.PendingSignatureQueryInterface
		MultisignatureInfoQuery        query.MultisignatureInfoQueryInterface
		MultiSignatureParticipantQuery query.MultiSignatureParticipantQueryInterface
		Logger                         *logrus.Logger
	}
)

func NewMultisigService(
	executor query.ExecutorInterface,
	blockService coreService.BlockServiceInterface,
	pendingTransactionQuery query.PendingTransactionQueryInterface,
	pendingSignatureQuery query.PendingSignatureQueryInterface,
	multisignatureQuery query.MultisignatureInfoQueryInterface,
	multiSignatureParticipantQuery query.MultiSignatureParticipantQueryInterface,
) *MultisigService {
	return &MultisigService{
		Executor:                       executor,
		BlockService:                   blockService,
		PendingTransactionQuery:        pendingTransactionQuery,
		PendingSignatureQuery:          pendingSignatureQuery,
		MultisignatureInfoQuery:        multisignatureQuery,
		MultiSignatureParticipantQuery: multiSignatureParticipantQuery,
	}
}

func (ms *MultisigService) GetPendingTransactions(
	param *model.GetPendingTransactionsRequest,
) (*model.GetPendingTransactionsResponse, error) {
	var (
		totalRecords uint32
		result       []*model.PendingTransaction
		err          error
		musigQuery   = query.NewPendingTransactionQuery()
		caseQuery    = query.NewCaseQuery()
		selectQuery  string
		args         []interface{}
	)
	caseQuery.Select(musigQuery.TableName, musigQuery.Fields...)
	if param.GetSenderAddress() != nil {
		caseQuery.Where(caseQuery.Equal("sender_address", param.GetSenderAddress()))
	}
	caseQuery.Where(caseQuery.Equal("status", param.GetStatus()))
	caseQuery.Where(caseQuery.Equal("latest", true))

	selectQuery, args = caseQuery.Build()

	countQuery := query.GetTotalRecordOfSelect(selectQuery)

	countRow, _ := ms.Executor.ExecuteSelectRow(countQuery, false, args...)
	err = countRow.Scan(
		&totalRecords,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "FailToGetTotalItemInPendingTransaction")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	caseQuery.OrderBy(param.GetPagination().GetOrderField(), param.GetPagination().GetOrderBy())
	caseQuery.Paginate(
		param.GetPagination().GetLimit(),
		param.GetPagination().GetPage(),
	)
	selectQuery, args = caseQuery.Build()
	pendingTransactionsRows, err := ms.Executor.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer pendingTransactionsRows.Close()
	result, err = ms.PendingTransactionQuery.BuildModel(result, pendingTransactionsRows)
	if err != nil {
		return nil, err
	}
	return &model.GetPendingTransactionsResponse{
		Count:               totalRecords,
		Page:                param.GetPagination().GetPage(),
		PendingTransactions: result,
	}, err
}

func (ms *MultisigService) GetPendingTransactionDetailByTransactionHash(
	param *model.GetPendingTransactionDetailByTransactionHashRequest,
) (*model.GetPendingTransactionDetailByTransactionHashResponse, error) {
	var (
		validStartHeight        uint32
		pendingTx               = &model.PendingTransaction{}
		pendingSigs             []*model.PendingSignature
		multisigInfo            = &model.MultiSignatureInfo{}
		err                     error
		pendingTransactionQuery = query.NewPendingTransactionQuery()
		pendingSignatureQuery   = query.NewPendingSignatureQuery()
		multisigInfoQuery       = query.NewMultisignatureInfoQuery()
		caseQuery               = query.NewCaseQuery()
	)
	// get current block height
	lastBlock, err := ms.BlockService.GetLastBlock()
	if err != nil {
		ms.Logger.Error(err)
		return nil, status.Error(codes.Internal, "server error")
	}
	if lastBlock.Height > constant.MinRollbackBlocks {
		validStartHeight = lastBlock.Height - constant.MinRollbackBlocks
	}
	// get pending transaction
	txHash, err := hex.DecodeString(param.GetTransactionHashHex())
	if err != nil {
		ms.Logger.Error(err)
		return nil, status.Error(codes.Internal, "server error")
	}
	caseQuery.Select(pendingTransactionQuery.TableName, pendingTransactionQuery.Fields...)
	caseQuery.Where(caseQuery.Equal("transaction_hash", txHash))
	caseQuery.Where(caseQuery.Equal("latest", true))
	selectPendingTxQuery, args := caseQuery.Build()
	pendingTxRow, _ := ms.Executor.ExecuteSelectRow(selectPendingTxQuery, false, args...)
	err = ms.PendingTransactionQuery.Scan(pendingTx, pendingTxRow)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "tx not found")
		}
		ms.Logger.Error(err)
		return nil, status.Error(codes.Internal, "server error")
	}
	// get pending signatures
	caseQuery = query.NewCaseQuery()
	caseQuery.Select(pendingSignatureQuery.TableName, pendingSignatureQuery.Fields...)
	caseQuery.Where(caseQuery.Equal("transaction_hash", txHash))
	caseQuery.Where(caseQuery.Equal("latest", true))
	if pendingTx.Status == model.PendingTransactionStatus_PendingTransactionPending {
		caseQuery.Where(caseQuery.GreaterEqual("block_height", validStartHeight))
	}
	selectPendingSigQuery, args := caseQuery.Build()
	pendingSigRows, err := ms.Executor.ExecuteSelect(selectPendingSigQuery, false, args...)
	if err != nil {
		ms.Logger.Error(err)
		return nil, status.Error(codes.Internal, "server error")
	}
	defer pendingSigRows.Close()
	pendingSigs, err = ms.PendingSignatureQuery.BuildModel(pendingSigs, pendingSigRows)
	if err != nil {
		ms.Logger.Error(err)
		return nil, status.Error(codes.Internal, "server error")
	}
	// get multisig info if exist
	caseQuery = query.NewCaseQuery()
	caseQuery.Select(multisigInfoQuery.TableName, multisigInfoQuery.Fields...)
	caseQuery.Where(caseQuery.Equal("latest", true))
	caseQuery.Where(caseQuery.Equal("multisig_address", pendingTx.SenderAddress))
	if pendingTx.Status == model.PendingTransactionStatus_PendingTransactionPending {
		caseQuery.Where(caseQuery.GreaterEqual("block_height", validStartHeight))
	}
	selectMultisigInfoQuery, args := caseQuery.Build()
	multisigInfoRow, _ := ms.Executor.ExecuteSelectRow(selectMultisigInfoQuery, false, args...)
	err = ms.MultisignatureInfoQuery.Scan(multisigInfo, multisigInfoRow)
	if err != nil && err != sql.ErrNoRows {
		ms.Logger.Error(err)
		return nil, status.Error(codes.Internal, "server error")
	}
	if err != sql.ErrNoRows {
		multisigInfo.Addresses, err = ms.getMultisigAddressParticipants(pendingTx.SenderAddress)
		if err != nil {
			if err != sql.ErrNoRows {
				ms.Logger.Error(err)
				return nil, status.Error(codes.Internal, "server error")
			}
		}
	}

	return &model.GetPendingTransactionDetailByTransactionHashResponse{
		PendingTransaction: pendingTx,
		PendingSignatures:  pendingSigs,
		MultiSignatureInfo: multisigInfo,
	}, nil
}

// getMultisigAddressParticipants returns multisignature participants addresses for a given multisig address and (
// optional) a start block height (for transaction in 'pending' status)
func (ms *MultisigService) getMultisigAddressParticipants(
	multisigAddress []byte,
) (multisigParticipants [][]byte, err error) {
	var (
		caseQuery                = query.NewCaseQuery()
		multisigParticipantQuery = query.NewMultiSignatureParticipantQuery()
	)
	caseQuery.Select(multisigParticipantQuery.TableName, multisigParticipantQuery.Fields...)
	// TODO: multisig participants should not have latest field
	caseQuery.Where(caseQuery.Equal("latest", true))
	caseQuery.Where(caseQuery.Equal("multisig_address", multisigAddress))
	caseQuery.OrderBy("account_address_index", model.OrderBy_DESC)
	selectMultisigParticipantsQuery, args := caseQuery.Build()
	multisigParticipantRows, err := ms.Executor.ExecuteSelect(selectMultisigParticipantsQuery, false, args...)
	if err != nil {
		return nil, err
	}
	defer multisigParticipantRows.Close()
	participants, err := ms.MultiSignatureParticipantQuery.BuildModel(multisigParticipantRows)
	if err != nil {
		return nil, err
	}
	for _, participant := range participants {
		multisigParticipants = append(multisigParticipants, participant.GetAccountAddress())
	}
	return multisigParticipants, nil

}

func (ms *MultisigService) GetMultisignatureInfo(
	param *model.GetMultisignatureInfoRequest,
) (*model.GetMultisignatureInfoResponse, error) {
	var (
		multiSignatureInfos []*model.MultiSignatureInfo
		caseQuery           = query.NewCaseQuery()
		multisigInfoQuery   = query.NewMultisignatureInfoQuery()
		selectQuery         string
		args                []interface{}
		totalRecords        uint32
		err                 error
	)
	caseQuery.Select(multisigInfoQuery.TableName, multisigInfoQuery.Fields...)
	if param.GetMultisigAddress() != nil {
		caseQuery.Where(caseQuery.Equal("multisig_address", param.GetMultisigAddress()))
	}
	caseQuery.Where(caseQuery.Equal("latest", true))
	selectQuery, args = caseQuery.Build()
	countQuery := query.GetTotalRecordOfSelect(selectQuery)
	countRow, _ := ms.Executor.ExecuteSelectRow(countQuery, false, args...)
	err = countRow.Scan(
		&totalRecords,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "FailToGetTotalItemInMultisigInfo")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	caseQuery.OrderBy(param.GetPagination().GetOrderField(), param.GetPagination().GetOrderBy())
	caseQuery.Paginate(
		param.GetPagination().GetLimit(),
		param.GetPagination().GetPage(),
	)
	selectQuery, args = caseQuery.Build()
	multisigInfoRows, err := ms.Executor.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer multisigInfoRows.Close()
	multiSignatureInfos, err = ms.MultisignatureInfoQuery.BuildModel([]*model.MultiSignatureInfo{}, multisigInfoRows)
	if err != nil {
		return nil, err
	}
	for idx, multisigInfo := range multiSignatureInfos {
		multiSignatureInfos[idx].Addresses, err = ms.getMultisigAddressParticipants(multisigInfo.GetMultisigAddress())
		if err != nil {
			if err != sql.ErrNoRows {
				ms.Logger.Error(err)
				return nil, status.Error(codes.Internal, "server error")
			}
		}
	}
	return &model.GetMultisignatureInfoResponse{
		Count:              totalRecords,
		Page:               param.GetPagination().GetPage(),
		MultisignatureInfo: multiSignatureInfos,
	}, err
}

func (ms *MultisigService) GetMultisigAddressByParticipantAddress(
	param *model.GetMultisigAddressByParticipantAddressRequest,
) (*model.GetMultisigAddressByParticipantAddressResponse, error) {
	var (
		multisigAddresses              = [][]byte{}
		caseQuery                      = query.NewCaseQuery()
		multisignatureParticipantQuery = query.NewMultiSignatureParticipantQuery()
		selectQuery                    string
		args                           []interface{}
		totalRecords                   uint32
		err                            error
	)

	caseQuery.Select(multisignatureParticipantQuery.TableName, []string{"multisig_address"}...)
	caseQuery.Where(caseQuery.Equal("account_address", param.ParticipantAddress))

	selectQuery, args = caseQuery.Build()
	countQuery := query.GetTotalRecordOfSelect(selectQuery)
	countRow, _ := ms.Executor.ExecuteSelectRow(countQuery, false, args...)

	err = countRow.Scan(
		&totalRecords,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "FailToGetTotal")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	caseQuery.OrderBy(param.GetPagination().GetOrderField(), param.GetPagination().GetOrderBy())

	selectQuery, args = caseQuery.Build()
	multiSignatureAddressesRows, err := ms.Executor.ExecuteSelect(selectQuery, false, args...)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer multiSignatureAddressesRows.Close()

	for multiSignatureAddressesRows.Next() {
		var multisigAddress []byte
		err = multiSignatureAddressesRows.Scan(
			&multisigAddress,
		)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil, status.Error(codes.Internal, err.Error())
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		multisigAddresses = append(multisigAddresses, multisigAddress)
	}

	return &model.GetMultisigAddressByParticipantAddressResponse{
		Total:             totalRecords,
		MultisigAddresses: multisigAddresses,
	}, err
}

func (ms *MultisigService) GetPendingTransactionsByHeight(
	fromHeight, toHeight uint32,
) ([]*model.PendingTransaction, error) {
	var (
		result         []*model.PendingTransaction
		err            error
		pendingTxQuery = query.NewPendingTransactionQuery()
		caseQuery      = query.NewCaseQuery()
		selectQuery    string
		args           []interface{}
	)
	caseQuery.Select(pendingTxQuery.TableName, pendingTxQuery.Fields...)

	caseQuery.Where(caseQuery.Between("block_height", fromHeight, toHeight))
	caseQuery.OrderBy("block_height", model.OrderBy_ASC)
	selectQuery, args = caseQuery.Build()
	pendingTransactionsRows, err := ms.Executor.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer pendingTransactionsRows.Close()
	result, err = ms.PendingTransactionQuery.BuildModel(result, pendingTransactionsRows)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ms *MultisigService) GetMultisigAddressesByBlockHeightRange(
	param *model.GetMultisigAddressesByBlockHeightRangeRequest,
) (*model.GetMultisigAddressesByBlockHeightRangeResponse, error) {
	var (
		multiSignatureInfos []*model.MultiSignatureInfo
		caseQuery           = query.NewCaseQuery()
		multisigInfoQuery   = query.NewMultisignatureInfoQuery()
		selectQuery         string
		args                []interface{}
		totalRecords        uint32
		err                 error
	)
	caseQuery.Select(multisigInfoQuery.TableName, multisigInfoQuery.Fields...)
	caseQuery.Where(caseQuery.Equal("latest", true))
	caseQuery.And(caseQuery.Between("block_height", param.FromBlockHeight, param.ToBlockHeight))
	selectQuery, args = caseQuery.Build()
	countQuery := query.GetTotalRecordOfSelect(selectQuery)
	countRow, _ := ms.Executor.ExecuteSelectRow(countQuery, false, args...)
	err = countRow.Scan(
		&totalRecords,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "FailToGetTotalItemInMultisigInfo")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	caseQuery.OrderBy(param.GetPagination().GetOrderField(), param.GetPagination().GetOrderBy())
	caseQuery.Paginate(
		param.GetPagination().GetLimit(),
		param.GetPagination().GetPage(),
	)
	selectQuery, args = caseQuery.Build()
	multisigInfoRows, err := ms.Executor.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer multisigInfoRows.Close()
	multiSignatureInfos, err = ms.MultisignatureInfoQuery.BuildModel(multiSignatureInfos, multisigInfoRows)
	if err != nil {
		return nil, err
	}
	for idx, multisigInfo := range multiSignatureInfos {
		multiSignatureInfos[idx].Addresses, err = ms.getMultisigAddressParticipants(multisigInfo.GetMultisigAddress())
		if err != nil {
			if err != sql.ErrNoRows {
				ms.Logger.Error(err)
				return nil, status.Error(codes.Internal, "server error")
			}
		}
	}
	return &model.GetMultisigAddressesByBlockHeightRangeResponse{
		Count:              totalRecords,
		Page:               param.GetPagination().GetPage(),
		MultisignatureInfo: multiSignatureInfos,
	}, err
}

func (ms *MultisigService) GetParticipantsByMultisigAddresses(
	param *model.GetParticipantsByMultisigAddressesRequest,
) (*model.GetParticipantsByMultisigAddressesResponse, error) {
	var (
		multiSignatureParticipants     = make(map[string]*model.MultiSignatureParticipants)
		multiSignatureParticipant      model.MultiSignatureParticipant
		caseQuery                      = query.NewCaseQuery()
		multisignatureParticipantQuery = query.NewMultiSignatureParticipantQuery()
		selectQuery                    string
		args                           []interface{}
		totalRecords                   uint32
		err                            error
		result                         []*model.MultiSignatureParticipant
	)

	caseQuery.Select(multisignatureParticipantQuery.TableName, multisignatureParticipantQuery.Fields...)
	var multisigAddressesParam []interface{}
	for _, v := range param.MultisigAddresses {
		multisigAddressesParam = append(multisigAddressesParam, v)
	}
	caseQuery.Where(caseQuery.In("multisig_address", multisigAddressesParam...))

	selectQuery, args = caseQuery.Build()
	countQuery := query.GetTotalRecordOfSelect(selectQuery)
	countRow, _ := ms.Executor.ExecuteSelectRow(countQuery, false, args...)

	err = countRow.Scan(
		&totalRecords,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "FailToGetTotal")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	caseQuery.OrderBy(param.GetPagination().GetOrderField(), param.GetPagination().GetOrderBy())

	selectQuery, args = caseQuery.Build()
	multiSignatureParticipantRows, err := ms.Executor.ExecuteSelect(selectQuery, false, args...)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer multiSignatureParticipantRows.Close()

	result, err = multisignatureParticipantQuery.BuildModel(multiSignatureParticipantRows)
	if err != nil {
		return nil, err
	}

	participantMultiSignatureAddressHex := hex.EncodeToString(multiSignatureParticipant.MultiSignatureAddress)
	for _, msParticipant := range result {
		if multiSignatureParticipants[participantMultiSignatureAddressHex] == nil {
			multiSignatureParticipants[participantMultiSignatureAddressHex] = &model.MultiSignatureParticipants{}
		}

		multiSignatureParticipants[participantMultiSignatureAddressHex].MultiSignatureParticipants = append(
			multiSignatureParticipants[participantMultiSignatureAddressHex].MultiSignatureParticipants,
			msParticipant)
	}

	return &model.GetParticipantsByMultisigAddressesResponse{
		Total:                      totalRecords,
		MultiSignatureParticipants: multiSignatureParticipants,
	}, err
}
