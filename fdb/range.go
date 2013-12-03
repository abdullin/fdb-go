// FoundationDB Go API
// Copyright (c) 2013 FoundationDB, LLC

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package fdb

/*
 #define FDB_API_VERSION 200
 #include <foundationdb/fdb_c.h>
*/
import "C"

import (
	"fmt"
)

// KeyValue represents a single key-value pair in the database.
type KeyValue struct {
	Key Key
	Value []byte
}

// RangeOptions specify how a database range read operation is carried
// out. RangeOptions objects are passed to GetRange() methods of Database,
// Transaction and Snapshot.
//
// The zero value of RangeOptions represents the default range read
// configuration (no limit, lexicographic order, to be used as an iterator).
type RangeOptions struct {
	// Limit restricts the number of key-value pairs returned as part of a range
	// read. A value of 0 indicates no limit.
	Limit int

	// Mode sets the streaming mode of the range read, allowing the database to
	// balance latency and bandwidth for this read.
	Mode StreamingMode

	// Reverse indicates that the read should be performed in lexicographic
	// (false) or reverse lexicographic (true) order. When Reverse is true and
	// Limit is non-zero, the last Limit key-value pairs in the range are
	// returned.
	Reverse bool
}

// A Range describes all keys between a begin (inclusive) and end (exclusive)
// key selector. A Range is provided to a read method of the FoundationDB API,
// and the key selectors are resolved to specific keys in the database while
// satisfying the read.
type Range interface {
	// FDBRangeKeySelectors returns a pair of key selectors that describe the
	// beginning and end of a range.
	FDBRangeKeySelectors() (begin, end Selectable)
}

// An ExactRange describes all keys between a begin (inclusive) and end
// (exclusive) key. An ExactRange is provided to a method of the FoundationDB
// API that does not already incur a read latency. If you need to specify an
// exact range using key selectors, you must first resolve the selectors to keys
// using the GetKey() method.
//
// Any object that implements ExactRange also implements Range, and may be used
// accordingly.
type ExactRange interface {
	// FDBRangeKeys returns a pair of keys that describe the beginning and end
	// of a range.
	FDBRangeKeys() (begin, end KeyConvertible)

	// An object that implements ExactRange must also implement Range
	// (logically, by returning FirstGreaterOrEqual() of the keys returned by
	// FDBRangeKeys).
	Range
}

// KeyRange is an ExactRange constructed from a pair of KeyConvertibles. Note
// that the default zero-value of KeyRange specifies an empty range before all
// keys in the database.
type KeyRange struct {
	Begin, End KeyConvertible
}

// FDBRangeKeys allows KeyRange to satisfy the ExactRange interface.
func (kr KeyRange) FDBRangeKeys() (KeyConvertible, KeyConvertible) {
	return kr.Begin, kr.End
}

// FDBRangeKeySelectors allows KeyRange to satisfy the Range interface.
func (kr KeyRange) FDBRangeKeySelectors() (Selectable, Selectable) {
	return FirstGreaterOrEqual(kr.Begin), FirstGreaterOrEqual(kr.End)
}

// SelectorRange is a Range constructed directly from a pair of Selectable
// objects. Note that the default zero-value of SelectorRange specifies an empty
// range before all keys in the database.
type SelectorRange struct {
	Begin, End Selectable
}

// FDBRangeKeySelectors allows SelectorRange to satisfy the Range interface.
func (sr SelectorRange) FDBRangeKeySelectors() (Selectable, Selectable) {
	return sr.Begin, sr.End
}

// RangeResult is a handle to the asynchronous result of a range
// read. RangeResult is safe for concurrent use by multiple goroutines.
type RangeResult struct {
	t *transaction
	sr SelectorRange
	options RangeOptions
	snapshot bool
	f *futureKeyValueArray
}

// GetSliceWithError returns a slice of KeyValue objects satisfying the range
// specified in the read that returned this RangeResult, or an error if any of
// the asynchronous operations associated with this result did not successfully
// complete. The current goroutine will be blocked until the read has completed.
func (rr RangeResult) GetSliceWithError() ([]KeyValue, error) {
	var ret []KeyValue

	ri := rr.Iterator()

	if rr.options.Limit != 0 {
		ri.options.Mode = StreamingModeExact
	} else {
		ri.options.Mode = StreamingModeWantAll
	}

	for ri.Advance() {
		if ri.err != nil {
			return nil, ri.err
		}
		ret = append(ret, ri.kvs...)
		ri.index = len(ri.kvs)
		ri.fetchNextBatch()
	}

	return ret, nil
}

// GetSliceOrPanic returns a slice of KeyValue objects satisfying the range
// specified in the read that returned this RangeResult, or panics if any of the
// asynchronous operations associated with this result did not successfully
// complete. The current goroutine will be blocked until the read has completed.
func (rr RangeResult) GetSliceOrPanic() []KeyValue {
	kvs, e := rr.GetSliceWithError()
	if e != nil {
		panic(e)
	}
	return kvs
}

// Iterator returns a RangeIterator over the key-value pairs satisfying the
// range specified in the read that returned this RangeResult.
func (rr RangeResult) Iterator() *RangeIterator {
	return &RangeIterator{
		t: rr.t,
		f: rr.f,
		sr: rr.sr,
		options: rr.options,
		iteration: 1,
		snapshot: rr.snapshot,
	}
}

// RangeIterator returns the key-value pairs in the database (as KeyValue
// objects) satisfying the range specified in a range read. RangeIterator is
// constructed with the (RangeResult).Iterator() method.
//
// RangeIterator should not be copied or used concurrently from multiple
// goroutines.
type RangeIterator struct {
	t *transaction
	f *futureKeyValueArray
	sr SelectorRange
	options RangeOptions
	iteration int
	done bool
	more bool
	kvs []KeyValue
	index int
	err error
	snapshot bool
}

// Advance attempts to advance the iterator to the next key-value pair. Advance
// returns true if there are more key-value pairs satisfying the range, or false
// if the range has been exhausted.
func (ri *RangeIterator) Advance() bool {
	if ri.done {
		return false
	}

	if ri.f == nil {
		return true
	}

	ri.kvs, ri.more, ri.err = ri.f.GetWithError()
	ri.index = 0
	ri.f = nil
	
	if ri.err != nil || len(ri.kvs) > 0 {
		return true
	}

	return false
}

func (ri *RangeIterator) fetchNextBatch() {
	if !ri.more || ri.index == ri.options.Limit {
		ri.done = true
		return
	}

	if ri.options.Limit > 0 {
		// Not worried about this being zero, checked equality above
		ri.options.Limit -= ri.index
	}

	if ri.options.Reverse {
		ri.sr.End = FirstGreaterOrEqual(ri.kvs[ri.index-1].Key)
	} else {
		ri.sr.Begin = FirstGreaterThan(ri.kvs[ri.index-1].Key)
	}

	ri.iteration += 1

	f := ri.t.doGetRange(ri.sr, ri.options, ri.snapshot, ri.iteration)
	ri.f = &f
}

// GetNextWithError returns the next KeyValue in a range read, or an error if
// one of the asynchronous operations associated with this range did not
// successfully complete. The Advance method of this RangeIterator must have
// returned true prior to calling GetNextWithError.
func (ri *RangeIterator) GetNextWithError() (kv KeyValue, e error) {
	if ri.err != nil {
		e = ri.err
		return
	}

	kv = ri.kvs[ri.index]

	ri.index += 1

	if ri.index == len(ri.kvs) {
		ri.fetchNextBatch()
	}

	return
}

// GetNextOrPanic returns the next KeyValue in a range read, or panics if one of
// the asynchronous operations associated with this range did not successfully
// complete. The Advance method of this RangeIterator must have returned true
// prior to calling GetNextWithError.
func (ri *RangeIterator) GetNextOrPanic() KeyValue {
	kv, e := ri.GetNextWithError()
	if e != nil {
		panic(e)
	}
	return kv
}

func strinc(prefix []byte) ([]byte, error) {
	for i := len(prefix) - 1; i >= 0; i-- {
		if prefix[i] != 0xFF {
			ret := make([]byte, i+1)
			copy(ret, prefix[:i+1])
			ret[i] += 1
			return ret, nil
		}
	}

	return nil, fmt.Errorf("Key must contain at least one byte not equal to 0xFF")
}

// PrefixRange returns the KeyRange describing the range of keys k such that
// bytes.HasPrefix(k, prefix) is true. PrefixRange returns an error if prefix
// consists entirely of zero of more 0xFF bytes.
func PrefixRange(prefix []byte) (KeyRange, error) {
	begin := make([]byte, len(prefix))
	copy(begin, prefix)
	end, e := strinc(begin)
	if e != nil {
		return KeyRange{}, nil
	}
	return KeyRange{Key(begin), Key(end)}, nil
}
