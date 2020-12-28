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
package util

import (
	"bytes"
	"encoding/binary"
)

// ConvertBytesToUint64 takes array of bytes and return the uint64 representation of the given bytes
func ConvertBytesToUint64(dataBytes []byte) uint64 {
	return binary.LittleEndian.Uint64(dataBytes)
}

// ConvertBytesToUint32 takes array of bytes and return the uint32 representation of the given bytes
func ConvertBytesToUint32(dataBytes []byte) uint32 {
	return binary.LittleEndian.Uint32(dataBytes)
}

// ConvertBytesToUint16 takes array of bytes and return the uint16 representation of the given bytes
func ConvertBytesToUint16(dataBytes []byte) uint16 {
	return binary.LittleEndian.Uint16(dataBytes)
}

// ConvertUint64ToBytes takes the uint64 decimal number and return the byte array representation of the given number
func ConvertUint64ToBytes(number uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, number)
	return buf
}

// ConvertUint32ToBytes takes the uint32 decimal number and return the byte array representation of the given number
func ConvertUint32ToBytes(number uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, number)
	return buf
}

// ConvertUint16ToBytes takes the uint16 decimal number and return the byte array representation of the given number
func ConvertUint16ToBytes(number uint16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, number)
	return buf
}

// ConvertIntToBytes takes the int decimal number and return the byte array representation of the given number
func ConvertIntToBytes(number int) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(number))
	return buf
}

// ConvertStringToBytes takes a string and return the byte array representation of the given string
func ConvertStringToBytes(str string) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(ConvertUint32ToBytes(uint32(len(str))))
	buffer.Write(bytes.NewBufferString(str).Bytes())
	return buffer.Bytes()
}
