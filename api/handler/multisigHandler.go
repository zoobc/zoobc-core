// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	MultisigHandler struct {
		MultisigService service.MultisigServiceInterface
	}
)

func (msh *MultisigHandler) GetPendingTransactions(
	ctx context.Context,
	req *model.GetPendingTransactionsRequest,
) (*model.GetPendingTransactionsResponse, error) {
	if req.Pagination == nil {
		req.Pagination = &model.Pagination{
			OrderField: "block_height",
			OrderBy:    model.OrderBy_DESC,
			Page:       1,
			Limit:      constant.MaxAPILimitPerPage,
		}
	}
	if req.GetPagination().GetPage() < 1 {
		return nil, status.Error(codes.InvalidArgument, "PageCannotBeLessThanOne")
	}
	if req.GetPagination().GetOrderField() == "" {
		req.Pagination.OrderField = "block_height"
		req.Pagination.OrderBy = model.OrderBy_DESC
	}

	return msh.MultisigService.GetPendingTransactions(req)
}

func (msh *MultisigHandler) GetPendingTransactionsByHeight(
	ctx context.Context,
	req *model.GetPendingTransactionsByHeightRequest,
) (*model.GetPendingTransactionsByHeightResponse, error) {

	if req.GetToHeight() < 1 {
		return nil, status.Error(codes.InvalidArgument, "ToHeightMustBeGreaterThanZero")
	}
	if req.GetFromHeight() >= req.GetToHeight() {
		return nil, status.Error(codes.InvalidArgument, "FromHeightMustBeLowerThanToHeight")
	}
	if req.GetToHeight()-req.GetFromHeight() > constant.MaxAPILimitPerPage {
		return nil, status.Error(codes.InvalidArgument, "HeightRangeMustBeLessThanOrEqualTo500")
	}
	pendingTxs, err := msh.MultisigService.GetPendingTransactionsByHeight(req.GetFromHeight(), req.GetToHeight())
	if err != nil {
		return nil, err
	}
	return &model.GetPendingTransactionsByHeightResponse{
		PendingTransactions: pendingTxs,
	}, nil
}

func (msh *MultisigHandler) GetPendingTransactionDetailByTransactionHash(
	_ context.Context,
	req *model.GetPendingTransactionDetailByTransactionHashRequest,
) (*model.GetPendingTransactionDetailByTransactionHashResponse, error) {
	result, err := msh.MultisigService.GetPendingTransactionDetailByTransactionHash(req)
	return result, err
}

func (msh *MultisigHandler) GetMultisignatureInfo(
	_ context.Context,
	req *model.GetMultisignatureInfoRequest,
) (*model.GetMultisignatureInfoResponse, error) {
	if req.Pagination == nil {
		req.Pagination = &model.Pagination{
			OrderField: "block_height",
			OrderBy:    model.OrderBy_DESC,
			Page:       1,
			Limit:      constant.MaxAPILimitPerPage,
		}
	}
	if req.GetPagination().GetPage() < 1 {
		return nil, status.Error(codes.InvalidArgument, "PageCannotBeLessThanOne")
	}
	if req.GetPagination().GetOrderField() == "" {
		req.Pagination.OrderField = "block_height"
		req.Pagination.OrderBy = model.OrderBy_DESC
	}
	if req.GetPagination().GetPage() > 30 {
		return nil, status.Error(codes.InvalidArgument, "LimitCannotBeMoreThan30")
	}
	result, err := msh.MultisigService.GetMultisignatureInfo(req)
	return result, err
}

func (msh *MultisigHandler) GetMultisigAddressByParticipantAddress(
	_ context.Context,
	req *model.GetMultisigAddressByParticipantAddressRequest,
) (*model.GetMultisigAddressByParticipantAddressResponse, error) {
	if req.Pagination == nil {
		req.Pagination = &model.Pagination{
			OrderField: "block_height",
			OrderBy:    model.OrderBy_DESC,
		}
	}
	if req.GetPagination().GetOrderField() == "" {
		req.Pagination.OrderField = "block_height"
		req.Pagination.OrderBy = model.OrderBy_DESC
	}
	result, err := msh.MultisigService.GetMultisigAddressByParticipantAddress(req)
	return result, err
}

func (msh *MultisigHandler) GetMultisigAddressesByBlockHeightRange(
	_ context.Context,
	req *model.GetMultisigAddressesByBlockHeightRangeRequest,
) (*model.GetMultisigAddressesByBlockHeightRangeResponse, error) {
	if req.Pagination == nil {
		req.Pagination = &model.Pagination{
			OrderField: "block_height",
			OrderBy:    model.OrderBy_DESC,
		}
	}
	if req.GetPagination().GetOrderField() == "" {
		req.Pagination.OrderField = "block_height"
		req.Pagination.OrderBy = model.OrderBy_DESC
	}
	result, err := msh.MultisigService.GetMultisigAddressesByBlockHeightRange(req)
	return result, err
}

func (msh *MultisigHandler) GetParticipantsByMultisigAddresses(
	_ context.Context,
	req *model.GetParticipantsByMultisigAddressesRequest,
) (*model.GetParticipantsByMultisigAddressesResponse, error) {

	if req.Pagination == nil {
		req.Pagination = &model.Pagination{
			OrderField: "block_height",
			OrderBy:    model.OrderBy_DESC,
		}
	}

	if req.GetPagination().GetOrderField() == "" {
		req.Pagination.OrderField = "block_height"
		req.Pagination.OrderBy = model.OrderBy_DESC
	}

	if len(req.MultisigAddresses) == 0 {
		return nil, status.Error(codes.InvalidArgument, "At least 1 address is required")
	}

	result, err := msh.MultisigService.GetParticipantsByMultisigAddresses(req)
	return result, err
}
