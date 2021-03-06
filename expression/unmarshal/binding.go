//  Copyright (c) 2015 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package unmarshal

import (
	"encoding/json"

	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/expression/parser"
)

func UnmarshalBinding(body []byte) (*expression.Binding, error) {
	var _unmarshalled struct {
		NameVar string `json:"name_var"`
		Var     string `json:"var"`
		Expr    string `json:"expr"`
		Desc    bool   `json:"desc"`
	}

	err := json.Unmarshal(body, &_unmarshalled)
	if err != nil {
		return nil, err
	}

	expr, err := parser.Parse(_unmarshalled.Expr)
	if err != nil {
		return nil, err
	}

	return expression.NewBinding(_unmarshalled.NameVar, _unmarshalled.Var, expr, _unmarshalled.Desc), nil
}

func UnmarshalBindings(body []byte) (expression.Bindings, error) {
	var _unmarshalled []struct {
		NameVar string `json:"name_var"`
		Var     string `json:"var"`
		Expr    string `json:"expr"`
		Desc    bool   `json:"desc"`
	}

	err := json.Unmarshal(body, &_unmarshalled)
	if err != nil {
		return nil, err
	}

	bindings := make(expression.Bindings, len(_unmarshalled))
	for i, binding := range _unmarshalled {
		expr, err := parser.Parse(binding.Expr)
		if err != nil {
			return nil, err
		}

		bindings[i] = expression.NewBinding(binding.NameVar, binding.Var, expr, binding.Desc)
	}

	return bindings, nil
}
