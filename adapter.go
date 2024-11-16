// Copyright 2020 by Blank-Xu. All Rights Reserved.
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
	"errors"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// the supported for Casbin interfaces
var (
	_ persist.Adapter                 = new(Adapter)
	_ persist.ContextAdapter          = new(Adapter)
	_ persist.FilteredAdapter         = new(Adapter)
	_ persist.ContextFilteredAdapter  = new(Adapter)
	_ persist.BatchAdapter            = new(Adapter)
	_ persist.ContextBatchAdapter     = new(Adapter)
	_ persist.UpdatableAdapter        = new(Adapter)
	_ persist.ContextUpdatableAdapter = new(Adapter)
)

// the supported driver names
var supportedDriverNames = map[adapterDriverNameIndex][]string{
	_SQLite:     {"sqlite", "sqlite3", "nrsqlite3"},
	_MySQL:      {"mysql", "nrmysql"},
	_PostgreSQL: {"postgres", "pgx", "pq-timeouts", "cloudsql-postgres", "ql", "nrpostgres", "cockroach"},
	_SQLServer:  {"sqlserver", "azuresql"},
}

// NewAdapter  the constructor for Adapter.
// db should connected to database and controlled by user.
// If tableName == "", the Adapter will automatically create a table named "casbin_rule".
func NewAdapter(db *sql.DB, driverName, tableName string) (*Adapter, error) {
	return NewAdapterWithContext(context.Background(), db, driverName, tableName)
}

// NewAdapterWithContext  the constructor for Adapter.
// db should connected to database and controlled by user.
// If tableName == "", the Adapter will automatically create a table named "casbin_rule".
func NewAdapterWithContext(ctx context.Context, db *sql.DB, driverName, tableName string) (*Adapter, error) {
	// check parameters first
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	if db == nil {
		return nil, errors.New("db is nil")
	}

	driverNameIndex, err := getAdapterDriverNameIndex(driverName)
	if err != nil {
		return nil, err
	}

	if tableName == "" {
		tableName = defaultTableName
	}

	dao := newDao(db, driverNameIndex, tableName)

	// check db connection
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// check adapter table
	if !dao.IsTableExist(ctx) {
		if err = dao.CreateTable(ctx); err != nil {
			return nil, err
		}
	}

	return &Adapter{ctx: ctx, dao: dao}, nil
}

func getAdapterDriverNameIndex(driverName string) (adapterDriverNameIndex, error) {
	for driverNameIndex, drivers := range supportedDriverNames {
		for _, supportedDriver := range drivers {
			if driverName == supportedDriver {
				return driverNameIndex, nil
			}
		}
	}

	switch driverName {
	case "mssql":
		return 0, errors.New("driver name mssql not support, please use sqlserver")
	case "oci8", "ora", "goracle":
		return 0, errors.New("sqladapter: please checkout 'oracle' branch")
	default:
		return 0, fmt.Errorf("unsupported driver name: %s", driverName)
	}
}

// Adapter  defines the database adapter for Casbin.
// It can load policy lines from connected database or save policy lines.
type Adapter struct {
	dao dao

	// Could not be removed until Casbin adapter interface support context as the first parameter.
	ctx context.Context

	filtered interface{}
}

// loadPolicyLine  load a policy line to model.
func (Adapter) loadPolicyLine(line rule, model model.Model) error {
	// return persist.LoadPolicyLine(strings.Join(line.Data(), ","), model)

	return persist.LoadPolicyArray(line.Data(), model)
}

// genArgs generate args from ptype and rule.
func (Adapter) genArgs(ptype string, rule []string) []interface{} {
	args := make([]interface{}, maxParameterCount)
	args[0] = ptype

	for idx := range rule {
		args[idx+1] = strings.TrimSpace(rule[idx])
	}

	for idx := len(rule) + 1; idx < maxParameterCount; idx++ {
		args[idx] = ""
	}

	return args
}

// LoadPolicy  load all policy rules from the storage.
func (adapter *Adapter) LoadPolicy(model model.Model) error {
	return adapter.LoadPolicyCtx(adapter.ctx, model)
}

// LoadPolicyCtx loads all policy rules from the storage with context.
func (adapter *Adapter) LoadPolicyCtx(ctx context.Context, model model.Model) error {
	lines, err := adapter.dao.SelectAll(ctx)
	if err != nil {
		return err
	}

	adapter.filtered = nil

	for _, line := range lines {
		if err = adapter.loadPolicyLine(line, model); err != nil {
			return err
		}
	}

	return nil
}

// SavePolicy  save policy rules to the storage.
func (adapter Adapter) SavePolicy(model model.Model) error {
	return adapter.SavePolicyCtx(adapter.ctx, model)
}

// SavePolicyCtx saves all policy rules to the storage with context.
func (adapter Adapter) SavePolicyCtx(ctx context.Context, model model.Model) error {
	if adapter.filtered != nil {
		return errors.New("could not save filtered policies")
	}

	args := make([][]interface{}, 0, 128)

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			arg := adapter.genArgs(ptype, rule)
			args = append(args, arg)
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			arg := adapter.genArgs(ptype, rule)
			args = append(args, arg)
		}
	}

	return adapter.dao.DeleteAllAndInsertRows(ctx, args)
}

// AddPolicy  add one policy rule to the storage.
func (adapter Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	return adapter.AddPolicyCtx(adapter.ctx, sec, ptype, rule)
}

// AddPolicyCtx adds a policy rule to the storage with context.
// This is part of the Auto-Save feature.
func (adapter Adapter) AddPolicyCtx(ctx context.Context, sec string, ptype string, rule []string) error {
	args := adapter.genArgs(ptype, rule)

	return adapter.dao.InsertRow(ctx, args...)
}

// AddPolicies  add multiple policy rules to the storage.
func (adapter Adapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	return adapter.AddPoliciesCtx(adapter.ctx, sec, ptype, rules)
}

// AddPoliciesCtx adds policy rules to the storage.
// This is part of the Auto-Save feature.
func (adapter Adapter) AddPoliciesCtx(ctx context.Context, sec string, ptype string, rules [][]string) error {
	args := make([][]interface{}, 0, len(rules))

	for _, rule := range rules {
		arg := adapter.genArgs(ptype, rule)
		args = append(args, arg)
	}

	return adapter.dao.InsertRows(ctx, args)
}

// RemovePolicy  remove policy rules from the storage.
func (adapter Adapter) RemovePolicy(sec, ptype string, rule []string) error {
	return adapter.RemovePolicyCtx(adapter.ctx, sec, ptype, rule)
}

// RemovePolicyCtx removes a policy rule from the storage with context.
// This is part of the Auto-Save feature.
func (adapter Adapter) RemovePolicyCtx(ctx context.Context, sec string, ptype string, rule []string) error {
	return adapter.dao.DeleteByArgs(ctx, ptype, rule)
}

// RemoveFilteredPolicy  remove policy rules that match the filter from the storage.
func (adapter Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return adapter.RemoveFilteredPolicyCtx(adapter.ctx, sec, ptype, fieldIndex, fieldValues...)
}

// RemoveFilteredPolicyCtx removes policy rules that match the filter from the storage with context.
// This is part of the Auto-Save feature.
func (adapter Adapter) RemoveFilteredPolicyCtx(ctx context.Context, sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	whereCondition, whereArgs := adapter.dao.GenFilteredCondition(ptype, fieldIndex, fieldValues...)

	return adapter.dao.DeleteByCondition(ctx, whereCondition, whereArgs...)
}

// RemovePolicies removes policy rules from the storage.
// This is part of the Auto-Save feature.
func (adapter Adapter) RemovePolicies(sec string, ptype string, rules [][]string) (err error) {
	return adapter.RemovePoliciesCtx(adapter.ctx, sec, ptype, rules)
}

// RemovePoliciesCtx removes policy rules from the storage.
// This is part of the Auto-Save feature.
func (adapter Adapter) RemovePoliciesCtx(ctx context.Context, sec string, ptype string, rules [][]string) error {
	args := make([][]interface{}, len(rules))

	for idx, rule := range rules {
		arg := adapter.genArgs(ptype, rule)
		args[idx] = arg
	}

	return adapter.dao.DeleteRows(ctx, args)
}

// LoadFilteredPolicy  load policy rules that match the Filter.
// filterPtr must be a pointer.
func (adapter *Adapter) LoadFilteredPolicy(model model.Model, filterPtr interface{}) error {
	return adapter.LoadFilteredPolicyCtx(adapter.ctx, model, filterPtr)
}

// LoadFilteredPolicyCtx loads only policy rules that match the filter.
func (adapter *Adapter) LoadFilteredPolicyCtx(ctx context.Context, model model.Model, filterPtr interface{}) error {
	if filterPtr == nil {
		return adapter.LoadPolicy(model)
	}

	filter, ok := filterPtr.(*Filter)
	if !ok {
		return errors.New("invalid filter type")
	}

	lines, err := adapter.dao.SelectByFilter(ctx, filter.genData())
	if err != nil {
		return err
	}

	for _, line := range lines {
		if err = adapter.loadPolicyLine(line, model); err != nil {
			return err
		}
	}

	adapter.filtered = struct{}{}

	return nil
}

// IsFiltered  returns true if the loaded policy rules has been filtered.
func (adapter Adapter) IsFiltered() bool {
	return adapter.IsFilteredCtx(adapter.ctx)
}

// IsFilteredCtx returns true if the loaded policy has been filtered.
func (adapter Adapter) IsFilteredCtx(ctx context.Context) bool {
	return adapter.filtered != nil
}

// UpdatePolicy update a policy rule from storage.
// This is part of the Auto-Save feature.
func (adapter Adapter) UpdatePolicy(sec, ptype string, oldRule, newRule []string) error {
	return adapter.UpdatePolicyCtx(adapter.ctx, sec, ptype, oldRule, newRule)
}

// UpdatePolicyCtx updates a policy rule from storage.
// This is part of the Auto-Save feature.
func (adapter Adapter) UpdatePolicyCtx(ctx context.Context, sec string, ptype string, oldRule, newRule []string) error {
	oldArgs := adapter.genArgs(ptype, oldRule)
	newArgs := adapter.genArgs(ptype, newRule)

	return adapter.dao.UpdateRow(ctx, append(newArgs, oldArgs...)...)
}

// UpdatePolicies updates policy rules to storage.
func (adapter Adapter) UpdatePolicies(sec, ptype string, oldRules, newRules [][]string) (err error) {
	return adapter.UpdatePoliciesCtx(adapter.ctx, sec, ptype, oldRules, newRules)
}

// UpdatePoliciesCtx updates some policy rules to storage, like db, redis.
func (adapter Adapter) UpdatePoliciesCtx(ctx context.Context, sec string, ptype string, oldRules, newRules [][]string) error {
	if len(oldRules) != len(newRules) {
		return errors.New("old rules size not equal to new rules size")
	}

	args := make([][]interface{}, 0, len(oldRules)+len(newRules))

	for idx := range oldRules {
		oldArgs := adapter.genArgs(ptype, oldRules[idx])
		newArgs := adapter.genArgs(ptype, newRules[idx])
		args = append(args, append(newArgs, oldArgs...))
	}

	return adapter.dao.UpdateRows(ctx, args)
}

// UpdateFilteredPolicies deletes old rules and adds new rules.
func (adapter Adapter) UpdateFilteredPolicies(sec, ptype string, newRules [][]string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	return adapter.UpdateFilteredPoliciesCtx(adapter.ctx, sec, ptype, newRules, fieldIndex, fieldValues...)
}

// UpdateFilteredPoliciesCtx deletes old rules and adds new rules.
func (adapter Adapter) UpdateFilteredPoliciesCtx(ctx context.Context, sec string, ptype string, newRules [][]string, fieldIndex int, fieldValues ...string) (oldPolicies [][]string, err error) {
	whereCondition, whereArgs := adapter.dao.GenFilteredCondition(ptype, fieldIndex, fieldValues...)

	var oldRules []rule
	oldRules, err = adapter.dao.SelectByCondition(ctx, whereCondition, whereArgs...)
	if err != nil {
		return
	}

	args := make([][]interface{}, 0, len(newRules))
	for _, policy := range newRules {
		arg := adapter.genArgs(ptype, policy)
		args = append(args, arg)
	}

	if err = adapter.dao.UpdateFilteredRows(ctx, whereCondition, whereArgs, args); err != nil {
		return
	}

	oldPolicies = make([][]string, 0, len(oldRules))
	for _, rule := range oldRules {
		oldPolicies = append(oldPolicies, rule.Data())
	}

	return
}
