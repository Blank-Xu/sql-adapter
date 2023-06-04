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
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

var (
	_ persist.Adapter          = new(Adapter)
	_ persist.FilteredAdapter  = new(Adapter)
	_ persist.BatchAdapter     = new(Adapter)
	_ persist.UpdatableAdapter = new(Adapter)
)

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
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	if db == nil {
		return nil, errors.New("db is nil")
	}

	// check db connection
	err := db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	if tableName == "" {
		tableName = defaultTableName
	}

	d, err := getDao(db, driverName, tableName)
	if err != nil {
		return nil, err
	}

	if !d.IsTableExist(ctx) {
		if err = d.CreateTable(ctx); err != nil {
			return nil, err
		}
	}

	adapter := Adapter{ctx: ctx, dao: d}

	return &adapter, nil
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
	lines, err := adapter.dao.SelectAll(adapter.ctx)
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

	return adapter.dao.DeleteAllAndInsertRows(adapter.ctx, args)
}

// AddPolicy  add one policy rule to the storage.
func (adapter Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	args := adapter.genArgs(ptype, rule)

	return adapter.dao.InsertRow(adapter.ctx, args...)
}

// AddPolicies  add multiple policy rules to the storage.
func (adapter Adapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	args := make([][]interface{}, 0, len(rules))

	for _, rule := range rules {
		arg := adapter.genArgs(ptype, rule)
		args = append(args, arg)
	}

	return adapter.dao.InsertRows(adapter.ctx, args)
}

// RemovePolicy  remove policy rules from the storage.
func (adapter Adapter) RemovePolicy(sec, ptype string, rule []string) error {
	return adapter.dao.DeleteByArgs(adapter.ctx, ptype, rule)
}

// RemoveFilteredPolicy  remove policy rules that match the filter from the storage.
func (adapter Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	whereCondition, whereArgs := adapter.dao.GenFilteredCondition(ptype, fieldIndex, fieldValues...)

	return adapter.dao.DeleteByCondition(adapter.ctx, whereCondition, whereArgs...)
}

func (adapter Adapter) RemovePolicies(sec string, ptype string, rules [][]string) (err error) {
	args := make([][]interface{}, len(rules))

	for idx, rule := range rules {
		arg := adapter.genArgs(ptype, rule)
		args[idx] = arg
	}

	return adapter.dao.DeleteRows(adapter.ctx, args)
}

// LoadFilteredPolicy  load policy rules that match the Filter.
// filterPtr must be a pointer.
func (adapter *Adapter) LoadFilteredPolicy(model model.Model, filterPtr interface{}) error {
	if filterPtr == nil {
		return adapter.LoadPolicy(model)
	}

	filter, ok := filterPtr.(*Filter)
	if !ok {
		return errors.New("invalid filter type")
	}

	lines, err := adapter.dao.SelectByFilter(adapter.ctx, filter.genData())
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
	return adapter.filtered != nil
}

// UpdatePolicy update a policy rule from storage.
// This is part of the Auto-Save feature.
func (adapter Adapter) UpdatePolicy(sec, ptype string, oldRule, newPolicy []string) error {
	oldArgs := adapter.genArgs(ptype, oldRule)
	newArgs := adapter.genArgs(ptype, newPolicy)

	return adapter.dao.UpdateRow(adapter.ctx, append(newArgs, oldArgs...)...)
}

// UpdatePolicies updates policy rules to storage.
func (adapter Adapter) UpdatePolicies(sec, ptype string, oldRules, newRules [][]string) (err error) {
	if len(oldRules) != len(newRules) {
		return errors.New("old rules size not equal to new rules size")
	}

	args := make([][]interface{}, 0, len(oldRules)+len(newRules))

	for idx := range oldRules {
		oldArgs := adapter.genArgs(ptype, oldRules[idx])
		newArgs := adapter.genArgs(ptype, newRules[idx])
		args = append(args, append(newArgs, oldArgs...))
	}

	return adapter.dao.UpdateRows(adapter.ctx, args)
}

// UpdateFilteredPolicies deletes old rules and adds new rules.
func (adapter Adapter) UpdateFilteredPolicies(sec, ptype string, newPolicies [][]string, fieldIndex int, fieldValues ...string) (oldPolicies [][]string, err error) {
	whereCondition, whereArgs := adapter.dao.GenFilteredCondition(ptype, fieldIndex, fieldValues...)

	var oldRules []rule
	oldRules, err = adapter.dao.SelectByCondition(adapter.ctx, whereCondition, whereArgs...)
	if err != nil {
		return
	}

	args := make([][]interface{}, 0, len(newPolicies))
	for _, policy := range newPolicies {
		arg := adapter.genArgs(ptype, policy)
		args = append(args, arg)
	}

	if err = adapter.dao.UpdateFilteredRows(adapter.ctx, whereCondition, whereArgs, args); err != nil {
		return
	}

	oldPolicies = make([][]string, 0, len(oldRules))
	for _, rule := range oldRules {
		oldPolicies = append(oldPolicies, rule.Data())
	}

	return
}
