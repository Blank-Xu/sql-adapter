// Copyright 2023 by Blank-Xu. All Rights Reserved.
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

// rule define the casbin rule model.
// It used for save or load policy lines from connected database.
type rule struct {
	PType string
	V0    string
	V1    string
	V2    string
	V3    string
	V4    string
	V5    string
}

func (rule rule) Data() []string {
	s := []string{rule.PType, rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5}
	data := make([]string, 0, maxParameterCount)

	for _, val := range s {
		if val == "" {
			break
		}
		data = append(data, val)
	}

	return data
}

// Filter define the filtering rules for a FilteredAdapter's policy.
// Empty values are ignored, but all others must match the Filter.
type Filter struct {
	PType []string
	V0    []string
	V1    []string
	V2    []string
	V3    []string
	V4    []string
	V5    []string
}

type filterData struct {
	fieldName string
	arg       []string
}

func (filter Filter) genData() [maxParameterCount]filterData {
	return [maxParameterCount]filterData{
		{"p_type", filter.PType},
		{"v0", filter.V0},
		{"v1", filter.V1},
		{"v2", filter.V2},
		{"v3", filter.V3},
		{"v4", filter.V4},
		{"v5", filter.V5},
	}
}
