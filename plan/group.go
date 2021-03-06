//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package plan

import (
	"encoding/json"

	"github.com/couchbase/query/algebra"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/expression/parser"
)

// Grouping of input data. Parallelizable.
type InitialGroup struct {
	readonly
	keys       expression.Expressions
	aggregates algebra.Aggregates
}

func NewInitialGroup(keys expression.Expressions, aggregates algebra.Aggregates) *InitialGroup {
	return &InitialGroup{
		keys:       keys,
		aggregates: aggregates,
	}
}

func (this *InitialGroup) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitInitialGroup(this)
}

func (this *InitialGroup) New() Operator {
	return &InitialGroup{}
}

func (this *InitialGroup) Keys() expression.Expressions {
	return this.keys
}

func (this *InitialGroup) Aggregates() algebra.Aggregates {
	return this.aggregates
}

func (this *InitialGroup) MarshalJSON() ([]byte, error) {
	r := map[string]interface{}{"#operator": "InitialGroup"}
	keylist := make([]string, 0, len(this.keys))
	for _, key := range this.keys {
		keylist = append(keylist, expression.NewStringer().Visit(key))
	}
	r["group_keys"] = keylist
	s := make([]interface{}, 0, len(this.aggregates))
	for _, agg := range this.aggregates {
		s = append(s, expression.NewStringer().Visit(agg))
	}
	r["aggregates"] = s
	return json.Marshal(r)
}

func (this *InitialGroup) UnmarshalJSON(body []byte) error {
	var _unmarshalled struct {
		_    string   `json:"#operator"`
		Keys []string `json:"group_keys"`
		Aggs []string `json:"aggregates"`
	}

	err := json.Unmarshal(body, &_unmarshalled)
	if err != nil {
		return err
	}

	this.keys = make(expression.Expressions, len(_unmarshalled.Keys))
	for i, key := range _unmarshalled.Keys {
		key_expr, err := parser.Parse(key)
		if err != nil {
			return err
		}
		this.keys[i] = key_expr
	}

	this.aggregates = make(algebra.Aggregates, len(_unmarshalled.Aggs))
	for i, agg := range _unmarshalled.Aggs {
		agg_expr, err := parser.Parse(agg)
		if err != nil {
			return err
		}
		this.aggregates[i], _ = agg_expr.(algebra.Aggregate)
	}

	return nil
}

// Grouping of groups. Recursable and parallelizable.
type IntermediateGroup struct {
	readonly
	keys       expression.Expressions
	aggregates algebra.Aggregates
}

func NewIntermediateGroup(keys expression.Expressions, aggregates algebra.Aggregates) *IntermediateGroup {
	return &IntermediateGroup{
		keys:       keys,
		aggregates: aggregates,
	}
}

func (this *IntermediateGroup) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitIntermediateGroup(this)
}

func (this *IntermediateGroup) New() Operator {
	return &IntermediateGroup{}
}

func (this *IntermediateGroup) Keys() expression.Expressions {
	return this.keys
}

func (this *IntermediateGroup) Aggregates() algebra.Aggregates {
	return this.aggregates
}

func (this *IntermediateGroup) MarshalJSON() ([]byte, error) {
	r := map[string]interface{}{"#operator": "IntermediateGroup"}
	keylist := make([]string, 0, len(this.keys))
	for _, key := range this.keys {
		keylist = append(keylist, expression.NewStringer().Visit(key))
	}
	r["group_keys"] = keylist
	s := make([]interface{}, 0, len(this.aggregates))
	for _, agg := range this.aggregates {
		s = append(s, expression.NewStringer().Visit(agg))
	}
	r["aggregates"] = s
	return json.Marshal(r)
}

func (this *IntermediateGroup) UnmarshalJSON(body []byte) error {
	var _unmarshalled struct {
		_    string   `json:"#operator"`
		Keys []string `json:"group_keys"`
		Aggs []string `json:"aggregates"`
	}

	err := json.Unmarshal(body, &_unmarshalled)
	if err != nil {
		return err
	}

	this.keys = make(expression.Expressions, len(_unmarshalled.Keys))
	for i, key := range _unmarshalled.Keys {
		key_expr, err := parser.Parse(key)
		if err != nil {
			return err
		}
		this.keys[i] = key_expr
	}

	this.aggregates = make(algebra.Aggregates, len(_unmarshalled.Aggs))
	for i, agg := range _unmarshalled.Aggs {
		agg_expr, err := parser.Parse(agg)
		if err != nil {
			return err
		}
		this.aggregates[i], _ = agg_expr.(algebra.Aggregate)
	}

	return nil
}

// Final grouping and aggregation.
type FinalGroup struct {
	readonly
	keys       expression.Expressions
	aggregates algebra.Aggregates
}

func NewFinalGroup(keys expression.Expressions, aggregates algebra.Aggregates) *FinalGroup {
	return &FinalGroup{
		keys:       keys,
		aggregates: aggregates,
	}
}

func (this *FinalGroup) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFinalGroup(this)
}

func (this *FinalGroup) New() Operator {
	return &FinalGroup{}
}

func (this *FinalGroup) Keys() expression.Expressions {
	return this.keys
}

func (this *FinalGroup) Aggregates() algebra.Aggregates {
	return this.aggregates
}

func (this *FinalGroup) MarshalJSON() ([]byte, error) {
	r := map[string]interface{}{"#operator": "FinalGroup"}
	keylist := make([]string, 0, len(this.keys))
	for _, key := range this.keys {
		keylist = append(keylist, expression.NewStringer().Visit(key))
	}
	r["group_keys"] = keylist
	s := make([]interface{}, 0, len(this.aggregates))
	for _, agg := range this.aggregates {
		s = append(s, expression.NewStringer().Visit(agg))
	}
	r["aggregates"] = s
	return json.Marshal(r)
}

func (this *FinalGroup) UnmarshalJSON(body []byte) error {
	var _unmarshalled struct {
		_    string   `json:"#operator"`
		Keys []string `json:"group_keys"`
		Aggs []string `json:"aggregates"`
	}

	err := json.Unmarshal(body, &_unmarshalled)
	if err != nil {
		return err
	}

	this.keys = make(expression.Expressions, len(_unmarshalled.Keys))
	for i, key := range _unmarshalled.Keys {
		key_expr, err := parser.Parse(key)
		if err != nil {
			return err
		}
		this.keys[i] = key_expr
	}

	this.aggregates = make(algebra.Aggregates, len(_unmarshalled.Aggs))
	for i, agg := range _unmarshalled.Aggs {
		agg_expr, err := parser.Parse(agg)
		if err != nil {
			return err
		}
		this.aggregates[i], _ = agg_expr.(algebra.Aggregate)
	}

	return nil
}
