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
package service

import (
	"math/big"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/core/smith/strategy"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	BlockServiceInterface interface {
		NewGenesisBlock(version uint32, previousBlockHash []byte, blockSeed, blockSmithPublicKey []byte, mRoot, mTree []byte,
			previousBlockHeight, referenceBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
			transactions []*model.Transaction, blockReceipts []*model.PublishedReceipt, spinePublicKeys []*model.SpinePublicKey,
			payloadHash []byte, payloadLength uint32, cumulativeDifficulty *big.Int, genesisSignature []byte) (*model.Block, error)
		GenerateBlock(
			previousBlock *model.Block,
			secretPhrase string,
			timestamp int64,
			empty bool,
		) (*model.Block, error)
		ValidateBlock(block, previousLastBlock *model.Block) error
		ValidatePayloadHash(block *model.Block) error
		GetPayloadHashAndLength(block *model.Block) (payloadHash []byte, payloadLength uint32, err error)
		PushBlock(previousBlock, block *model.Block, broadcast, persist bool) error
		GetBlockByID(id int64, withAttachedData bool) (*model.Block, error)
		GetBlockByHeight(uint32) (*model.Block, error)
		GetBlockByHeightCacheFormat(uint32) (*storage.BlockCacheObject, error)
		GetBlocksFromHeight(startHeight, limit uint32, withAttachedData bool) ([]*model.Block, error)
		GetLastBlock() (*model.Block, error)
		GetLastBlockCacheFormat() (*storage.BlockCacheObject, error)
		UpdateLastBlockCache(block *model.Block) error
		InitializeBlocksCache() error
		GetBlockHash(block *model.Block) ([]byte, error)
		GetBlocks() ([]*model.Block, error)
		PopulateBlockData(block *model.Block) error
		GetGenesisBlock() (*model.Block, error)
		GenerateGenesisBlock(genesisEntries []constant.GenesisConfigEntry) (*model.Block, error)
		AddGenesis() error
		CheckGenesis() (exist bool, err error)
		GetChainType() chaintype.ChainType
		ChainWriteLock(int)
		ChainWriteUnlock(actionType int)
		ReceiveBlock(
			senderPublicKey []byte,
			lastBlock, block *model.Block,
			nodeSecretPhrase string,
			peer *model.Peer,
		) (*model.Receipt, error)
		PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error)
		GetBlocksmithStrategy() strategy.BlocksmithStrategyInterface
		ReceivedValidatedBlockTransactionsListener() observer.Listener
		BlockTransactionsRequestedListener() observer.Listener
	}
)
