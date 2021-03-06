//  Copyright (c) 2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package system

import (
	"fmt"
	"strings"

	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/timestamp"
	"github.com/couchbase/query/value"
)

type storeKeyspace struct {
	namespace *namespace
	name      string
	si        datastore.Indexer
}

func (b *storeKeyspace) Release() {
}

func (b *storeKeyspace) NamespaceId() string {
	return b.namespace.Id()
}

func (b *storeKeyspace) Id() string {
	return b.Name()
}

func (b *storeKeyspace) Name() string {
	return b.name
}

func (b *storeKeyspace) Count() (int64, errors.Error) {
	return 1, nil
}

func (b *storeKeyspace) Indexer(name datastore.IndexType) (datastore.Indexer, errors.Error) {
	return b.si, nil
}

func (b *storeKeyspace) Indexers() ([]datastore.Indexer, errors.Error) {
	return []datastore.Indexer{b.si}, nil
}

func (b *storeKeyspace) Fetch(keys []string) ([]value.AnnotatedPair, []errors.Error) {
	var errs []errors.Error
	rv := make([]value.AnnotatedPair, 0, len(keys))
	for _, k := range keys {
		item, e := b.fetchOne(k)
		if e != nil {
			if errs == nil {
				errs = make([]errors.Error, 0, 1)
			}
			errs = append(errs, e)
			continue
		}

		if item != nil {
			item.SetAttachment("meta", map[string]interface{}{
				"id": k,
			})
		}

		rv = append(rv, value.AnnotatedPair{
			Name:  k,
			Value: item,
		})
	}

	return rv, errs
}

func (b *storeKeyspace) fetchOne(key string) (value.AnnotatedValue, errors.Error) {
	if key == b.namespace.store.actualStore.Id() {
		doc := value.NewAnnotatedValue(map[string]interface{}{
			"id":  b.namespace.store.actualStore.Id(),
			"url": b.namespace.store.actualStore.URL(),
		})
		return doc, nil
	}
	return nil, errors.NewSystemDatastoreError(nil, "Key Not Found "+key)
}

func (b *storeKeyspace) Insert(inserts []value.Pair) ([]value.Pair, errors.Error) {
	// FIXME
	return nil, errors.NewSystemNotImplementedError(nil, "")
}

func (b *storeKeyspace) Update(updates []value.Pair) ([]value.Pair, errors.Error) {
	// FIXME
	return nil, errors.NewSystemNotImplementedError(nil, "")
}

func (b *storeKeyspace) Upsert(upserts []value.Pair) ([]value.Pair, errors.Error) {
	// FIXME
	return nil, errors.NewSystemNotImplementedError(nil, "")
}

func (b *storeKeyspace) Delete(deletes []string) ([]string, errors.Error) {
	// FIXME
	return nil, errors.NewSystemNotImplementedError(nil, "")
}

func newStoresKeyspace(p *namespace) (*storeKeyspace, errors.Error) {
	b := new(storeKeyspace)
	b.namespace = p
	b.name = KEYSPACE_NAME_DATASTORES

	primary := &storeIndex{name: "#primary", keyspace: b}
	b.si = newSystemIndexer(b, primary)

	return b, nil
}

type storeIndex struct {
	name     string
	keyspace *storeKeyspace
}

func (pi *storeIndex) KeyspaceId() string {
	return pi.name
}

func (pi *storeIndex) Id() string {
	return pi.Name()
}

func (pi *storeIndex) Name() string {
	return pi.name
}

func (pi *storeIndex) Type() datastore.IndexType {
	return datastore.DEFAULT
}

func (pi *storeIndex) SeekKey() expression.Expressions {
	return nil
}

func (pi *storeIndex) RangeKey() expression.Expressions {
	return nil
}

func (pi *storeIndex) Condition() expression.Expression {
	return nil
}

func (pi *storeIndex) IsPrimary() bool {
	return true
}

func (pi *storeIndex) State() (state datastore.IndexState, msg string, err errors.Error) {
	return datastore.ONLINE, "", nil
}

func (pi *storeIndex) Statistics(requestId string, span *datastore.Span) (
	datastore.Statistics, errors.Error) {
	return nil, nil
}

func (pi *storeIndex) Drop(requestId string) errors.Error {
	return errors.NewSystemIdxNoDropError(nil, pi.Name())
}

func (pi *storeIndex) Scan(requestId string, span *datastore.Span, distinct bool, limit int64,
	cons datastore.ScanConsistency, vector timestamp.Vector, conn *datastore.IndexConnection) {
	defer close(conn.EntryChannel())

	val := ""

	a := span.Seek[0].Actual()
	switch a := a.(type) {
	case string:
		val = a
	default:
		conn.Error(errors.NewSystemDatastoreError(nil, fmt.Sprintf("Invalid seek value %v of type %T.", a, a)))
		return
	}

	if strings.EqualFold(val, pi.keyspace.namespace.store.actualStore.Id()) {
		entry := datastore.IndexEntry{PrimaryKey: pi.keyspace.namespace.store.actualStore.Id()}
		conn.EntryChannel() <- &entry
	}
}

func (pi *storeIndex) ScanEntries(requestId string, limit int64, cons datastore.ScanConsistency,
	vector timestamp.Vector, conn *datastore.IndexConnection) {
	defer close(conn.EntryChannel())

	entry := datastore.IndexEntry{PrimaryKey: pi.keyspace.namespace.store.actualStore.Id()}
	conn.EntryChannel() <- &entry
}
