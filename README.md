fdb-go
======

[Go language](http://golang.org) bindings for [FoundationDB](https://foundationdb.com), a distributed key-value store with ACID transactions.

This package requires:

- Go 1.1+ with CGO enabled
- FoundationDB C API 2.0.x or 3.0.x (part of the [FoundationDB clients package](https://foundationdb.com/get))

Use of this package requires the selection of a FoundationDB API version at runtime. This package currently supports FoundationDB API versions 200 and 300 (although version 300 requires a 3.0.x FoundationDB C library to be installed).

To install this package, run:

    go get github.com/FoundationDB/fdb-go/fdb

Documentation
-------------

* [API documentation](http://godoc.org/github.com/FoundationDB/fdb-go/fdb)
* [Tutorial](https://foundationdb.com/documentation/class-scheduling-go.html)
