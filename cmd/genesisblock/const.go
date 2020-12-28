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
package genesisblock

import (
	"github.com/spf13/cobra"
)

var (
	genesisCmd = &cobra.Command{
		Use:   "genesis",
		Short: "command used to generate a new genesis block.",
	}

	/*
		// for genesis generate command
	*/
	withDbLastState                                 bool
	dbPath, applicationCodeName, applicationVersion string
	extraNodesCount                                 int
	genesisTimestamp                                int

	/*
		// for genesis generate-consul-kv command
	*/
	logLevels              string
	wellKnownPeers         string
	deploymentName         string
	kvFileCustomConfigFile string

	/*
		ENV Target
	*/
	envTarget      string
	output         string
	envTargetValue = map[string]uint32{
		"develop":      0,
		"staging":      1,
		"alpha":        2,
		"local":        3,
		"experimental": 4,
		"beta":         5,
	}
)

type (
	genesisEntry struct {
		AccountAddress     string
		AccountSeed        string
		AccountBalance     int64
		NodeSeed           string
		NodePublicKey      string
		NodePublicKeyBytes []byte
		LockedBalance      int64
		ParticipationScore int64
		Smithing           bool
		Message            string
	}
	clusterConfigEntry struct {
		NodePublicKey  string `json:"NodePublicKey"`
		NodeSeed       string `json:"NodeSeed"`
		AccountAddress string `json:"AccountAddress"`
		Smithing       bool   `json:"Smithing,omitempty"`
	}
	accountNodeEntry struct {
		NodePublicKey  string `json:"NodePublicKey"`
		AccountAddress string `json:"AccountAddress"`
	}
	parseErrorLog struct {
		AccountAddress    string `json:"AccountAddress"`
		ConfigPublicKey   string `json:"ConfigPublicKey"`
		ComputedPublicKey string `json:"ComputedPublicKey"`
	}
)
