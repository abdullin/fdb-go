package main

import (
	"bitbucket.org/FoundationDB/fdb-go/fdb"
	"bitbucket.org/FoundationDB/fdb-go/fdb/tuple"
	"log"
	"fmt"
	"os"
	"strings"
	"sync"
	"runtime"
	"reflect"
)

const verbose bool = false

func int64ToBool(i int64) bool {
	switch i {
	case 0:
		return false
	default:
		return true
	}
}

type stackEntry struct {
	item interface{}
	idx int
}

type StackMachine struct {
	prefix []byte
	tr fdb.Transaction
	stack []stackEntry
	lastVersion int64
	threads sync.WaitGroup
	verbose bool
}

func newStackMachine(prefix []byte, verbose bool) *StackMachine {
	sm := StackMachine{verbose: verbose, prefix: prefix}
	return &sm
}

func (sm *StackMachine) waitAndPop() (ret stackEntry) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case fdb.Error:
				p, e := tuple.Pack(tuple.Tuple{[]byte("ERROR"), []byte(fmt.Sprintf("%d", int(r)))})
				if e != nil {
					panic(e)
				}
				ret.item = p
			default:
				panic(r)
			}
		}
	}()

	ret, sm.stack = sm.stack[len(sm.stack) - 1], sm.stack[:len(sm.stack) - 1]
	switch el := ret.item.(type) {
	case int64, []byte, string:
	case fdb.FutureNil:
		el.GetOrPanic()
		ret.item = []byte("RESULT_NOT_PRESENT")
	case fdb.FutureValue:
		v := el.GetOrPanic()
		if v != nil {
			ret.item = v
		} else {
			ret.item = []byte("RESULT_NOT_PRESENT")
		}
	case fdb.FutureKey:
		ret.item = el.GetOrPanic()
	case nil:
	default:
		log.Fatalf("Don't know how to pop stack element %v %T\n", el, el)
	}
	return
}

func (sm *StackMachine) pushRange(idx int, sl []fdb.KeyValue) {
	var t tuple.Tuple = make(tuple.Tuple, 0, len(sl) * 2)

	for _, kv := range(sl) {
		t = append(t, kv.Key)
		t = append(t, kv.Value)
	}

	p, e := tuple.Pack(t)
	if e != nil {
		panic(e)
	}

	sm.store(idx, p)
}

func (sm *StackMachine) store(idx int, item interface{}) {
	sm.stack = append(sm.stack, stackEntry{item, idx})
}

func (sm *StackMachine) dumpStack() {
	for i := len(sm.stack) - 1; i >= 0; i-- {
		el := sm.stack[i].item
		switch el := el.(type) {
		case int64:
			fmt.Printf(" %d", el)
		case fdb.FutureNil:
			fmt.Printf(" FutureNil")
		case fdb.FutureValue:
			fmt.Printf(" FutureValue")
		case fdb.FutureKey:
			fmt.Printf(" FutureKey")
		case []byte:
			fmt.Printf(" %q", string(el))
		case string:
			fmt.Printf(" %s", el)
		case nil:
			fmt.Printf(" nil")
		default:
			log.Fatalf("Don't know how to dump stack element %v %T\n", el, el)
		}
	}
}

func (sm *StackMachine) processInst(idx int, inst tuple.Tuple) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case fdb.Error:
				p, e := tuple.Pack(tuple.Tuple{[]byte("ERROR"), []byte(fmt.Sprintf("%d", int(r)))})
				if e != nil {
					panic(e)
				}
				sm.store(idx, p)
			default:
				panic(r)
			}
		}
	}()

	var e error

	op := string(inst[0].([]byte))
	if sm.verbose {
		fmt.Printf("Instruction is %s (%v)\n", op, sm.prefix)
		fmt.Printf("Stack from [")
		sm.dumpStack()
		fmt.Printf(" ]\n")
	}

	var obj interface{}
	switch {
	case strings.HasSuffix(op, "_SNAPSHOT"):
		obj = sm.tr.Snapshot()
		op = op[:len(op)-9]
	case strings.HasSuffix(op, "_DATABASE"):
		obj = db
		op = op[:len(op)-9]
	default:
		obj = sm.tr
	}

	switch string(op) {
	case "PUSH":
		sm.store(idx, inst[1])
	case "DUP":
		entry := sm.stack[len(sm.stack) - 1]
		sm.store(entry.idx, entry.item)
	case "EMPTY_STACK":
		sm.stack = []stackEntry{}
		sm.stack = make([]stackEntry, 0)
	case "SWAP":
		idx := sm.waitAndPop().item.(int64)
		sm.stack[len(sm.stack) - 1], sm.stack[len(sm.stack) - 1 - int(idx)] = sm.stack[len(sm.stack) - 1 - int(idx)], sm.stack[len(sm.stack) - 1]
	case "POP":
		sm.stack = sm.stack[:len(sm.stack) - 1]
	case "SUB":
		sm.store(idx, sm.waitAndPop().item.(int64) - sm.waitAndPop().item.(int64))
	case "NEW_TRANSACTION":
		sm.tr, e = db.CreateTransaction()
		if e != nil {
			panic(e)
		}
	case "ON_ERROR":
		sm.store(idx, sm.tr.OnError(fdb.Error(int(sm.waitAndPop().item.(int64)))))
	case "GET_READ_VERSION":
		sm.lastVersion = obj.(fdb.ReadTransaction).GetReadVersion().GetOrPanic()
		sm.store(idx, []byte("GOT_READ_VERSION"))
	case "SET":
		switch o := obj.(type) {
		case fdb.Database:
			e = o.Set(sm.waitAndPop().item.([]byte), sm.waitAndPop().item.([]byte))
			if e != nil {
				panic(e)
			}
			sm.store(idx, []byte("RESULT_NOT_PRESENT"))
		case fdb.Transaction:
			o.Set(sm.waitAndPop().item.([]byte), sm.waitAndPop().item.([]byte))
		}
	case "LOG_STACK":
		prefix := sm.waitAndPop().item.([]byte)
		for i := len(sm.stack)-1; i >= 0; i-- {
			if i % 100 == 0 {
				sm.tr.Commit().GetOrPanic()
			}

			el := sm.waitAndPop()

			var keyt tuple.Tuple
			keyt = append(keyt, int64(i))
			keyt = append(keyt, int64(el.idx))
			pk, e := tuple.Pack(keyt)
			if e != nil {
				panic(e)
			}
			pk = append(prefix, pk...)

			var valt tuple.Tuple
			valt = append(valt, el.item)
			pv, e := tuple.Pack(valt)
			if e != nil {
				panic(e)
			}

			vl := 40000
			if len(pv) < vl {
				vl = len(pv)
			}

			sm.tr.Set(pk, pv[:vl])
		}
		sm.tr.Commit().GetOrPanic()
	case "GET":
		switch o := obj.(type) {
		case fdb.Database:
			v, e := db.Get(sm.waitAndPop().item.([]byte))
			if e != nil {
				panic(e)
			}
			if v != nil {
				sm.store(idx, v)
			} else {
				sm.store(idx, []byte("RESULT_NOT_PRESENT"))
			}
		case fdb.ReadTransaction:
			sm.store(idx, o.Get(sm.waitAndPop().item.([]byte)))
		}
	case "COMMIT":
		sm.store(idx, sm.tr.Commit())
	case "RESET":
		sm.tr.Reset()
	case "CLEAR":
		switch o := obj.(type) {
		case fdb.Database:
			e := db.Clear(sm.waitAndPop().item.([]byte))
			if e != nil {
				panic(e)
			}
			sm.store(idx, []byte("RESULT_NOT_PRESENT"))
		case fdb.Transaction:
			o.Clear(sm.waitAndPop().item.([]byte))
		}
	case "SET_READ_VERSION":
		sm.tr.SetReadVersion(sm.lastVersion)
	case "WAIT_FUTURE":
		entry := sm.waitAndPop()
		sm.store(entry.idx, entry.item)
	case "GET_COMMITTED_VERSION":
		sm.lastVersion, e = sm.tr.GetCommittedVersion()
		if e != nil {
			panic(e)
		}
		sm.store(idx, []byte("GOT_COMMITTED_VERSION"))
	case "GET_KEY":
		sel := fdb.KeySelector{sm.waitAndPop().item.([]byte), int64ToBool(sm.waitAndPop().item.(int64)), int(sm.waitAndPop().item.(int64))}
		switch o := obj.(type) {
		case fdb.Database:
			v, e := o.GetKey(sel)
			if e != nil {
				panic(e)
			}
			sm.store(idx, v)
		case fdb.ReadTransaction:
			sm.store(idx, o.GetKey(sel))
		}
	case "GET_RANGE":
		begin := sm.waitAndPop().item.([]byte)
		end := sm.waitAndPop().item.([]byte)
		var limit int
		switch l := sm.waitAndPop().item.(type) {
		case int64:
			limit = int(l)
		}
		reverse := int64ToBool(sm.waitAndPop().item.(int64))
		mode := sm.waitAndPop().item.(int64)
		switch o := obj.(type) {
		case fdb.Database:
			v, e := db.GetRange(begin, end, fdb.RangeOptions{Limit: int(limit), Reverse: reverse, Mode: fdb.StreamingMode(mode+1)})
			if e != nil {
				panic(e)
			}
			sm.pushRange(idx, v)
		case fdb.ReadTransaction:
			sm.pushRange(idx, o.GetRange(begin, end, fdb.RangeOptions{Limit: int(limit), Reverse: reverse, Mode: fdb.StreamingMode(mode+1)}).GetSliceOrPanic())
		}
	case "CLEAR_RANGE":
		switch o := obj.(type) {
		case fdb.Database:
			e := o.ClearRange(sm.waitAndPop().item.([]byte), sm.waitAndPop().item.([]byte))
			if e != nil {
				panic(e)
			}
			sm.store(idx, []byte("RESULT_NOT_PRESENT"))
		case fdb.Transaction:
			o.ClearRange(sm.waitAndPop().item.([]byte), sm.waitAndPop().item.([]byte))
		}
	case "GET_RANGE_STARTS_WITH":
		prefix := sm.waitAndPop().item.([]byte)
		var limit int
		switch l := sm.waitAndPop().item.(type) {
		case int64:
			limit = int(l)
		}
		reverse := int64ToBool(sm.waitAndPop().item.(int64))
		mode := sm.waitAndPop().item.(int64)
		begin, end, e := fdb.PrefixRange(prefix)
		if e != nil {
			panic(e)
		}
		switch o := obj.(type) {
		case fdb.Database:
			v, e := db.GetRange(begin, end, fdb.RangeOptions{Limit: int(limit), Reverse: reverse, Mode: fdb.StreamingMode(mode+1)})
			if e != nil {
				panic(e)
			}
			sm.pushRange(idx, v)
		case fdb.ReadTransaction:
			sm.pushRange(idx, o.GetRange(begin, end, fdb.RangeOptions{Limit: int(limit), Reverse: reverse, Mode: fdb.StreamingMode(mode+1)}).GetSliceOrPanic())
		}
	case "GET_RANGE_SELECTOR":
		begin := fdb.KeySelector{Key: sm.waitAndPop().item.([]byte), OrEqual: int64ToBool(sm.waitAndPop().item.(int64)), Offset: int(sm.waitAndPop().item.(int64))}
		end := fdb.KeySelector{Key: sm.waitAndPop().item.([]byte), OrEqual: int64ToBool(sm.waitAndPop().item.(int64)), Offset: int(sm.waitAndPop().item.(int64))}
		var limit int
		switch l := sm.waitAndPop().item.(type) {
		case int64:
			limit = int(l)
		}
		reverse := int64ToBool(sm.waitAndPop().item.(int64))
		mode := sm.waitAndPop().item.(int64)
		switch o := obj.(type) {
		case fdb.Database:
			v, e := db.GetRangeSelector(begin, end, fdb.RangeOptions{Limit: int(limit), Reverse: reverse, Mode: fdb.StreamingMode(mode+1)})
			if e != nil {
				panic(e)
			}
			sm.pushRange(idx, v)
		case fdb.ReadTransaction:
			sm.pushRange(idx, o.GetRangeSelector(begin, end, fdb.RangeOptions{Limit: int(limit), Reverse: reverse, Mode: fdb.StreamingMode(mode+1)}).GetSliceOrPanic())
		}
	case "CLEAR_RANGE_STARTS_WITH":
		prefix := sm.waitAndPop().item.([]byte)
		begin, end, e := fdb.PrefixRange(prefix)
		if e != nil {
			panic(e)
		}
		switch o := obj.(type) {
		case fdb.Database:
			e := o.ClearRange(begin, end)
			if e != nil {
				panic(e)
			}
			sm.store(idx, []byte("RESULT_NOT_PRESENT"))
		case fdb.Transaction:
			o.ClearRange(begin, end)
		}
	case "TUPLE_PACK":
		var t tuple.Tuple
		count := sm.waitAndPop().item.(int64)
		for i := 0; i < int(count); i++ {
			t = append(t, sm.waitAndPop().item)
		}
		p, e := tuple.Pack(t)
		if e != nil {
			panic(e)
		}
		sm.store(idx, p)
	case "TUPLE_UNPACK":
		t, e := tuple.Unpack(sm.waitAndPop().item.([]byte))
		if e != nil {
			panic(e)
		}
		for _, el := range(t) {
			p, e := tuple.Pack(tuple.Tuple{el})
			if e != nil {
				panic(e)
			}
			sm.store(idx, p)
		}
	case "TUPLE_RANGE":
		var t tuple.Tuple
		count := sm.waitAndPop().item.(int64)
		for i := 0; i < int(count); i++ {
			t = append(t, sm.waitAndPop().item)
		}
		begin, end, e := tuple.Range(t)
		if e != nil {
			panic(e)
		}
		sm.store(idx, begin)
		sm.store(idx, end)
	case "START_THREAD":
		newsm := newStackMachine(sm.waitAndPop().item.([]byte), verbose)
		sm.threads.Add(1)
		go func() {
			newsm.Run()
			sm.threads.Done()
		}()
	case "WAIT_EMPTY":
		prefix := sm.waitAndPop().item.([]byte)
		begin, end, e := fdb.PrefixRange(prefix)
		if e != nil {
			panic(e)
		}
		db.Transact(func (tr fdb.Transaction) (interface{}, error) {
			v := tr.GetRange(begin, end, fdb.RangeOptions{}).GetSliceOrPanic()
			if len(v) != 0 {
				panic(fdb.Error(1020))
			}
			return nil, nil
		})
		sm.store(idx, []byte("WAITED_FOR_EMPTY"))
	case "READ_CONFLICT_RANGE":
		e = sm.tr.AddReadConflictRange(sm.waitAndPop().item.([]byte), sm.waitAndPop().item.([]byte))
		if e != nil {
			panic(e)
		}
		sm.store(idx, []byte("SET_CONFLICT_RANGE"))
	case "WRITE_CONFLICT_RANGE":
		e = sm.tr.AddWriteConflictRange(sm.waitAndPop().item.([]byte), sm.waitAndPop().item.([]byte))
		if e != nil {
			panic(e)
		}
		sm.store(idx, []byte("SET_CONFLICT_RANGE"))
	case "READ_CONFLICT_KEY":
		e = sm.tr.AddReadConflictKey(sm.waitAndPop().item.([]byte))
		if e != nil {
			panic(e)
		}
		sm.store(idx, []byte("SET_CONFLICT_KEY"))
	case "WRITE_CONFLICT_KEY":
		e = sm.tr.AddWriteConflictKey(sm.waitAndPop().item.([]byte))
		if e != nil {
			panic(e)
		}
		sm.store(idx, []byte("SET_CONFLICT_KEY"))
	case "ATOMIC_OP":
		opname := strings.Title(string(sm.waitAndPop().item.([]byte)))
		key := sm.waitAndPop().item.([]byte)
		value := sm.waitAndPop().item.([]byte)
		reflect.ValueOf(&obj).MethodByName(opname).Call([]reflect.Value{reflect.ValueOf(&key), reflect.ValueOf(&value)})
		switch obj.(type) {
		case fdb.Database:
			sm.store(idx, []byte("RESULT_NOT_PRESENT"))
		}
	case "DISABLE_WRITE_CONFLICT":
		sm.tr.Options().SetNextWriteNoWriteConflictRange()
	case "CANCEL":
		sm.tr.Cancel()
	case "UNIT_TESTS":
	default:
		log.Fatalf("Unhandled operation %s\n", string(inst[0].([]byte)))
	}

	if sm.verbose {
		fmt.Printf("        to [")
		sm.dumpStack()
		fmt.Printf(" ]\n\n")
	}

	runtime.Gosched()
}

func (sm *StackMachine) Run() {
	begin, end, e := tuple.Range(tuple.Tuple{sm.prefix})
	if e != nil {
		panic(e)
	}

	r, e := db.Transact(func (tr fdb.Transaction) (interface{}, error) {
		return tr.GetRange(begin, end, fdb.RangeOptions{}).GetSliceOrPanic(), nil
	})
	if e != nil {
		panic(e)
	}

	instructions := r.([]fdb.KeyValue)

	for i, kv := range(instructions) {
		inst, _ := tuple.Unpack(kv.Value)

		if sm.verbose {
			fmt.Printf("Instruction %d\n", i)
		}
		sm.processInst(i, inst)
	}

	sm.threads.Wait()
}

var db fdb.Database

var clusterFile string

func main() {
	prefix := []byte(os.Args[1])
	if len(os.Args) > 2 {
		clusterFile = os.Args[2]
	}

	var e error

	e = fdb.APIVersion(100)
	if e != nil {
		log.Fatal(e)
	}

	db, e = fdb.Open(clusterFile, "DB")
	if e != nil {
		log.Fatal(e)
	}

	sm := newStackMachine(prefix, verbose)

	sm.Run()
}
