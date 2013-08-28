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
 #define FDB_API_VERSION 100
 #include <foundationdb/fdb_c.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

type Cluster struct {
	c *C.FDBCluster
}

func (c *Cluster) destroy() {
	C.fdb_cluster_destroy(c.c)
}

func CreateCluster() (*Cluster, error) {
	f := C.fdb_create_cluster(nil)
	fdb_future_block_until_ready(f)
	outc := &C.FDBCluster{}
	if err := C.fdb_future_get_cluster(f, &outc); err != 0 {
		return nil, FDBError{Code: err}
	}
	C.fdb_future_destroy(f)
	c := &Cluster{c: outc}
	runtime.SetFinalizer(c, (*Cluster).destroy)
	return c, nil
}

func (c *Cluster) OpenDatabase(dbname []byte) (*Database, error) {
	f := C.fdb_cluster_create_database(c.c, (*C.uint8_t)(unsafe.Pointer(&dbname[0])), C.int(len(dbname)))
	fdb_future_block_until_ready(f)
	outd := &C.FDBDatabase{}
	if err := C.fdb_future_get_database(f, &outd); err != 0 {
		return nil, FDBError{Code: err}
	}
	C.fdb_future_destroy(f)
	d := &Database{d: outd}
	runtime.SetFinalizer(d, (*Database).destroy)
	return d, nil
}