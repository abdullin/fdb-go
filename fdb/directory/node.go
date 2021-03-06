// FoundationDB Go Directory Layer
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

package directory

import (
	"github.com/abdullin/fdb-go/fdb"
	"github.com/abdullin/fdb-go/fdb/subspace"
	"bytes"
)

type node struct {
	subspace subspace.Subspace
	path []string
	targetPath []string
	_layer fdb.FutureByteSlice
}

func (n *node) exists() bool {
	if n.subspace == nil {
		return false
	}
	return true
}

func (n *node) prefetchMetadata(rtr fdb.ReadTransaction) *node {
	if n.exists() {
		n.layer(rtr)
	}
	return n
}

func (n *node) layer(rtr fdb.ReadTransaction) fdb.FutureByteSlice {
	if n._layer == nil {
		fv := rtr.Get(n.subspace.Sub([]byte("layer")))
		n._layer = fv
	}

	return n._layer
}

func (n *node) isInPartition(tr *fdb.Transaction, includeEmptySubpath bool) bool {
	return n.exists() && bytes.Compare(n._layer.MustGet(), []byte("partition")) == 0 && (includeEmptySubpath || len(n.targetPath) > len(n.path))
}

func (n *node) getPartitionSubpath() []string {
	return n.targetPath[len(n.path):]
}

func (n *node) getContents(dl directoryLayer, tr *fdb.Transaction) (DirectorySubspace, error) {
	return dl.contentsOfNode(n.subspace, n.path, n._layer.MustGet())
}
