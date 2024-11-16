// Copyright 2024 by Blank-Xu. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sqladapter

import (
	"context"
	"database/sql"
	"testing"
)

// nolint: funlen,paralleltest
func TestNewAdapterWithContext(t *testing.T) {
	type params struct {
		ctx        context.Context
		db         *sql.DB
		driverName string
		tableName  string
	}

	tests := []struct {
		name    string
		params  params
		wantErr bool
	}{
		{
			name: "01 nil context",
			params: params{
				ctx: nil,
			},
			wantErr: true,
		},
		{
			name: "02 nil db",
			params: params{
				ctx: context.TODO(),
			},
			wantErr: true,
		},
		{
			name: "03 unsupported driver",
			params: params{
				ctx:        context.TODO(),
				driverName: "mssql",
				db:         &sql.DB{},
			},
			wantErr: true,
		},
		{
			name: "04 unsupported driver",
			params: params{
				ctx:        context.TODO(),
				driverName: "oci8",
				db:         &sql.DB{},
			},
			wantErr: true,
		},
		{
			name: "05 unsupported driver",
			params: params{
				ctx:        context.TODO(),
				driverName: "MongoDB",
				db:         &sql.DB{},
			},
			wantErr: true,
		},
		{
			name: "06 empty driver",
			params: params{
				ctx:        context.TODO(),
				driverName: "",
				db:         &sql.DB{},
			},
			wantErr: true,
		},
		{
			name: "07 invalid driver",
			params: params{
				ctx:        context.TODO(),
				driverName: "MyDB",
				db:         &sql.DB{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewAdapterWithContext(tt.params.ctx, tt.params.db, tt.params.driverName, tt.params.tableName)
			if tt.wantErr {
				if a != nil || err == nil {
					t.Errorf("test case[%s] failed", tt.name)
				}
			} else {
				if a == nil || err != nil {
					t.Errorf("test case[%s] failed", tt.name)
				}
			}
		})
	}
}
