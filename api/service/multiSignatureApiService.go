package service

import (
	"database/sql"
	"encoding/hex"
	"strings"

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
		GetPendingTransactionDetailByTransactionHash(
			param *model.GetPendingTransactionDetailByTransactionHashRequest,
		) (*model.GetPendingTransactionDetailByTransactionHashResponse, error)
		GetMultisignatureInfo(
			param *model.GetMultisignatureInfoRequest,
		) (*model.GetMultisignatureInfoResponse, error)
		GetMultisigAddressByParticipantAddresses(
			param *model.GetMultisigAddressByParticipantAddressesRequest,
		) (*model.GetMultisigAddressByParticipantAddressesResponse, error)
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
	if param.GetSenderAddress() != "" {
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
	// sub query for getting addresses from multisignature_participant
	subQ := query.NewCaseQuery()
	subQ.Select("multisignature_participant", "GROUP_CONCAT(account_address, ',')")
	subQ.Where(subQ.Equal("multisig_address", pendingTx.SenderAddress))
	subQ.GroupBy("multisig_address", "block_height")
	subQ.OrderBy("account_address_index", model.OrderBy_DESC)
	subQ.As("addresses")
	subStr, subArgs := subQ.SubBuild()

	caseQuery = query.NewCaseQuery()
	caseQuery.Select(multisigInfoQuery.TableName, append(multisigInfoQuery.Fields, subStr)...)
	caseQuery.Args = append(caseQuery.Args, subArgs...)

	caseQuery.Where(caseQuery.Equal("latest", true))

	if pendingTx.Status == model.PendingTransactionStatus_PendingTransactionPending {
		caseQuery.Where(caseQuery.GreaterEqual("block_height", validStartHeight))
	}
	selectMultisigInfoQuery, args := caseQuery.Build()
	multisigInfoRow, _ := ms.Executor.ExecuteSelectRow(selectMultisigInfoQuery, false, args...)
	err = ms.MultisignatureInfoQuery.Scan(multisigInfo, multisigInfoRow)
	if err != nil {
		if err != sql.ErrNoRows {
			ms.Logger.Error(err)
			return nil, status.Error(codes.Internal, "server error")
		}
	}
	return &model.GetPendingTransactionDetailByTransactionHashResponse{
		PendingTransaction: pendingTx,
		PendingSignatures:  pendingSigs,
		MultiSignatureInfo: multisigInfo,
	}, nil
}

func (ms *MultisigService) GetMultisignatureInfo(
	param *model.GetMultisignatureInfoRequest,
) (*model.GetMultisignatureInfoResponse, error) {
	var (
		result            []*model.MultiSignatureInfo
		caseQuery         = query.NewCaseQuery()
		multisigInfoQuery = query.NewMultisignatureInfoQuery()
		selectQuery       string
		args              []interface{}
		totalRecords      uint32
		err               error
	)
	// sub query for getting addresses from multisignature_participant
	subQ := query.NewCaseQuery()
	subQ.Select("multisignature_participant", "GROUP_CONCAT(account_address, ',')")
	subQ.GroupBy("multisig_address", "block_height")
	subQ.OrderBy("account_address_index", model.OrderBy_DESC)
	subQ.As("addresses")
	subStr, subArgs := subQ.SubBuild()

	caseQuery.Select(multisigInfoQuery.TableName, append(multisigInfoQuery.Fields, subStr)...)
	caseQuery.Args = append(caseQuery.Args, subArgs...)
	if param.GetMultisigAddress() != "" {
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
	result, err = ms.MultisignatureInfoQuery.BuildModel(result, multisigInfoRows)
	if err != nil {
		return nil, err
	}
	return &model.GetMultisignatureInfoResponse{
		Count:              totalRecords,
		Page:               param.GetPagination().GetPage(),
		MultisignatureInfo: result,
	}, err
}

func (ms *MultisigService) GetMultisigAddressByParticipantAddresses(
	param *model.GetMultisigAddressByParticipantAddressesRequest,
) (*model.GetMultisigAddressByParticipantAddressesResponse, error) {
	var (
		multiSignatureAddresses        = make(map[string]*model.Addresses)
		caseQuery                      = query.NewCaseQuery()
		multisignatureParticipantQuery = query.NewMultiSignatureParticipantQuery()
		selectQuery                    string
		args                           []interface{}
		totalRecords                   uint32
		err                            error
	)

	caseQuery.Select(multisignatureParticipantQuery.TableName, append([]string{"account_address"}, "GROUP_CONCAT(multisig_address, ',')")...)
	var accountAddresses []interface{}
	for _, v := range param.Addresses {
		accountAddresses = append(accountAddresses, v)
	}
	caseQuery.Where(caseQuery.In("account_address", accountAddresses...))
	caseQuery.GroupBy("account_address")
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
		var accountAddress string
		var multisigAddressesString string
		var multisigAddresses model.Addresses
		err = multiSignatureAddressesRows.Scan(
			&accountAddress,
			&multisigAddressesString,
		)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil, status.Error(codes.Internal, err.Error())
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		multisigAddresses.Addresses = strings.Split(multisigAddressesString, ",")
		multiSignatureAddresses[accountAddress] = &multisigAddresses
	}
	return &model.GetMultisigAddressByParticipantAddressesResponse{
		Total:                   totalRecords,
		MultiSignatureAddresses: multiSignatureAddresses,
	}, err
}
