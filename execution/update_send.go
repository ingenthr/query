//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package execution

import (
	"fmt"

	"github.com/couchbaselabs/query/catalog"
	"github.com/couchbaselabs/query/errors"
	"github.com/couchbaselabs/query/plan"
	"github.com/couchbaselabs/query/value"
)

// Send to keyspace
type SendUpdate struct {
	base
	plan *plan.SendUpdate
}

func NewSendUpdate(plan *plan.SendUpdate) *SendUpdate {
	rv := &SendUpdate{
		base: newBase(),
		plan: plan,
	}

	rv.output = rv
	return rv
}

func (this *SendUpdate) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitSendUpdate(this)
}

func (this *SendUpdate) Copy() Operator {
	return &SendUpdate{this.base.copy(), this.plan}
}

func (this *SendUpdate) RunOnce(context *Context, parent value.Value) {
}

func (this *SendUpdate) processItem(item value.AnnotatedValue, context *Context) bool {
	return this.enbatch(item, this, context)
}

func (this *SendUpdate) afterItems(context *Context) {
	this.flushBatch(context)
}

func (this *SendUpdate) flushBatch(context *Context) bool {
	if len(this.batch) == 0 {
		return true
	}

	pairs := make([]catalog.Pair, len(this.batch))

	for i, av := range this.batch {
		key, ok := this.requireKey(av, context)
		if !ok {
			return false
		}

		pairs[i].Key = key
		clone := av.GetAttachment("clone")
		switch clone := clone.(type) {
		case value.AnnotatedValue:
			pairs[i].Value = clone
		default:
			context.Error(errors.NewError(nil, fmt.Sprintf(
				"Invalid UPDATE value of type %T.", clone)))
			return false
		}
	}

	pairs, e := this.plan.Keyspace().Update(pairs)
	if e != nil {
		context.Error(e)
		this.batch = nil
		return false
	}

	for _, p := range pairs {
		if !this.sendItem(p.Value.(value.AnnotatedValue)) {
			break
		}
	}

	this.batch = nil
	return true
}
