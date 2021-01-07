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
package blocker

import (
	"encoding/json"
	"fmt"

	"github.com/zoobc/zoobc-core/common/monitoring"
)

type (
	TypeBlocker string

	Blocker struct {
		Type    TypeBlocker
		Message string
		Data    interface{}
	}
)

var (
	isDebugMode bool

	DBErr                     TypeBlocker = "DBErr"
	DBRowNotFound             TypeBlocker = "DBRowNotFound"
	NotFound                  TypeBlocker = "NotFound"
	BlockErr                  TypeBlocker = "BlockErr"
	BlockNotFoundErr          TypeBlocker = "BlockNotFoundErr"
	RequestParameterErr       TypeBlocker = "RequestParameterErr"
	AppErr                    TypeBlocker = "AppErr"
	AuthErr                   TypeBlocker = "AuthErr"
	ValidationErr             TypeBlocker = "ValidationErr"
	DuplicateMempoolErr       TypeBlocker = "DuplicateMempoolErr"
	DuplicateReceiptErr       TypeBlocker = "DuplicateReceiptErr"
	ParserErr                 TypeBlocker = "ParserErr"
	ServerError               TypeBlocker = "ServerError"
	SmithingErr               TypeBlocker = "SmithingErr"
	ZeroParticipationScoreErr TypeBlocker = "ZeroParticipationScoreErr"
	ChainValidationErr        TypeBlocker = "ChainValidationErr"
	P2PPeerError              TypeBlocker = "P2PPeerError"
	P2PPeerErrorDownload      TypeBlocker = "P2PPeerErrorDownload"
	P2PNetworkConnectionErr   TypeBlocker = "P2PNetworkConnectionErr"
	SmithingPending           TypeBlocker = "SmithingPending"
	InvalidBlockTimestamp     TypeBlocker = "InvalidBlockTimestamp"
	TimeoutExceeded           TypeBlocker = "TimeoutExceeded"
	PushMainBlockErr          TypeBlocker = "PushMainBlockErr"
	ValidateMainBlockErr      TypeBlocker = "ValidateMainBlockErr"
	PushSpineBlockErr         TypeBlocker = "PushSpineBlockErr"
	ValidateSpineBlockErr     TypeBlocker = "ValidateSpineBlockErr"
	SchedulerError            TypeBlocker = "SchedulerError"
	CacheEmpty                TypeBlocker = "CacheEmpty"
	IgnoredError              TypeBlocker = "IgnoredError"
)

func SetIsDebugMode(val bool) {
	isDebugMode = val
}

func NewBlocker(typeBlocker TypeBlocker, message string, data ...interface{}) error {
	monitoring.IncrementBlockerMetrics(string(typeBlocker))
	blocker := Blocker{
		Type:    typeBlocker,
		Message: message,
	}
	if isDebugMode {
		blocker.Data = data
	}
	return blocker
}

func (e Blocker) Error() string {
	if isDebugMode {
		j, _ := json.Marshal(e.Data)
		return fmt.Sprintf("%v: %v > %s", e.Type, e.Message, j)
	}
	return fmt.Sprintf("%v: %v", e.Type, e.Message)
}
