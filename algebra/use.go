//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package algebra

import (
	"github.com/couchbase/query/expression"
)

var EMPTY_USE = NewUse(nil, nil)

type Use struct {
	keys    expression.Expression
	indexes IndexRefs
}

func NewUse(keys expression.Expression, indexes IndexRefs) *Use {
	return &Use{keys, indexes}
}

func (this *Use) Keys() expression.Expression {
	return this.keys
}

func (this *Use) Indexes() IndexRefs {
	return this.indexes
}
