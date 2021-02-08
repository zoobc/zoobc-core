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
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestInitLogger(t *testing.T) {
	type args struct {
		path     string
		filename string
		levels   []string
		logOnCLI bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "wantError",
			args: args{
				path:     "",
				filename: "",
			},
			wantErr: true,
		},
		{
			name: "wantSuccess:FullLevels",
			args: args{
				path:     "./testdata/",
				filename: "test.log",
				levels:   []string{"info", "warn", "error", "fatal", "panic"},
				logOnCLI: true,
			},

			wantErr: false,
		},
		{
			name: "wantSuccess:NoLevels",
			args: args{
				path:     "./testdata/",
				filename: "test.log",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := InitLogger(tt.args.path, tt.args.filename, tt.args.levels, tt.args.logOnCLI); (got == nil) != tt.wantErr {
				t.Errorf("InitLogger() = %v, wantError %v", tt.name, tt.wantErr)
			}
		})
	}
}

func Test_hooker_Fire(t *testing.T) {
	type fields struct {
		Writer      *os.File
		EntryLevels []logrus.Level
	}
	type args struct {
		entry *logrus.Entry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "wantFail:Write",
			fields: fields{},
			args: args{
				entry: &logrus.Entry{
					Logger: logrus.New(),
				},
			},
			wantErr: true,
		},
		{
			name:   "wantFail:EntryToString",
			fields: fields{},
			args: args{
				entry: &logrus.Entry{
					Logger: &logrus.Logger{
						Formatter: &logrus.TextFormatter{},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := hooker{
				Writer:      tt.fields.Writer,
				EntryLevels: tt.fields.EntryLevels,
			}
			if err := h.Fire(tt.args.entry); (err != nil) != tt.wantErr {
				t.Errorf("hooker.Fire() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
