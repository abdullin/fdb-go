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
)

type Database struct {
	d *C.FDBDatabase
}

func (d *Database) destroy() {
	C.fdb_database_destroy(d.d)
}

func (d *Database) CreateTransaction() (*Transaction, error) {
	outt := &C.FDBTransaction{}
	if err := C.fdb_database_create_transaction(d.d, &outt); err != 0 {
		return nil, FDBError{Code: err}
	}
	t := &Transaction{outt}
	runtime.SetFinalizer(t, (*Transaction).destroy)
	return t, nil
}

func (d *Database) Transact(f func(tr *Transaction) (interface{}, error)) (ret interface{}, e error) {
	tr, e := d.CreateTransaction()
	/* Any error here is non-retryable */
	if e != nil {
		return
	}

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fdberror, ok := r.(FDBError)
					if ok {
						e = fdberror
					} else {
						panic(r)
					}
				}
			}()

			ret, e = f(tr)

			if e != nil {
				return
			}

			e = tr.Commit().GetWithError()
		}()

		/* No error means success! */
		if e == nil {
			return
		}

		fdberr, ok := e.(FDBError)
		if ok {
			e = tr.OnError(fdberr).GetWithError()
		}

		/* If OnError returns an error, then it's not
		/* retryable; otherwise take another pass at things */
		if e != nil {
			return
		}
	}
}