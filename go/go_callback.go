package dgo

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

type GoCallback uint64

var (
	goCallbackCounter uint32 = 0
	goCallbackMap     sync.Map
)

// cb must be a function
func PendGo(cb any) GoCallback {
	cbv := reflect.ValueOf(cb)
	if cbv.Kind() != reflect.Func {
		panic("dgo:go cb must be a function")
	}
	nextId := atomic.AddUint32(&goCallbackCounter, 1)
	if _, loaded := goCallbackMap.LoadOrStore(nextId, cbv); loaded {
		panic("dgo:go too many go callbacks pending")
	}
	return GoCallback(nextId)
}

func (gcb GoCallback) Exists() bool {
	_, loaded := goCallbackMap.Load(uint32(gcb))
	return loaded
}

func (gcb GoCallback) decompose() (id uint32, cf CallbackFlag) {
	return uint32(gcb & ((1 << 32) - 1)), CallbackFlag(gcb)
}

func (gcb GoCallback) handle(objs []any) {
	id, cf := gcb.decompose()
	var (
		v      any
		loaded bool
	)
	if cf.HasPop() {
		v, loaded = goCallbackMap.LoadAndDelete(id)
	} else {
		v, loaded = goCallbackMap.Load(id)
	}
	if !loaded {
		panic(fmt.Sprintf("dgo:go go callback not exist, id=%d", id))
	}
	fn := v.(reflect.Value)

	hasPackArray := cf.HasPackArray()

	var values []reflect.Value
	var args []any

	if hasPackArray {
		args = make([]any, 0, len(objs)+2)
	} else {
		values = make([]reflect.Value, 0, len(objs)+2)
	}

	if cf.HasWithCode() {
		if hasPackArray {
			args = append(args, cf)
		} else {
			values = append(values, reflect.ValueOf(cf))
		}
	}
	if cf.HasFast() {
		if len(objs) != 0 {
			panic("dgo:go go callback with flag CF_FAST should have no arguments")
		}
		var arg any
		switch cf.FastKind() {
		case CFFK_VOID:
			goto SKIP
		case CFFK_NIL:
			if hasPackArray {
				args = append(args, nil)
			} else {
				values = append(values, reflect.ValueOf(&arg).Elem())
			}
			goto SKIP
		case CFFK_NO:
			arg = false
		case CFFK_YES:
			arg = true
		}
		if hasPackArray {
			args = append(args, arg)
		} else {
			values = append(values, reflect.ValueOf(arg))
		}
	SKIP:
	} else {
		if hasPackArray {
			args = append(args, objs...)
		} else {
			for i, arg := range objs {
				if arg == nil {
					values = append(values, reflect.ValueOf(&objs[i]).Elem())
				} else {
					values = append(values, reflect.ValueOf(arg))
				}

			}
		}
	}
	if hasPackArray {
		values = []reflect.Value{reflect.ValueOf(args)}
	}
	fn.Call(values)
}
