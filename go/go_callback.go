package dgo

import (
	"fmt"
	"log"
	"reflect"
)

func (p *Port) Pend(callback any) GoCallback {
	callbackV := reflect.ValueOf(callback)
	if callbackV.Kind() != reflect.Func {
		panic("dgo:go: callback must be a function")
	}

	var callbackId uint64
	for n := 0; n < 10; n++ {
		callbackId = uint64(p.nextCallbackId.Add(1))
		_, loaded := p.goCallbacks.LoadOrStore(callbackId, callbackV)
		if !loaded {
			goto STORE_SUCCESS
		}
	}
	panic(fmt.Sprintf("dgo:go: too many callbacks pending on %s", p))

STORE_SUCCESS:
	return GoCallback{callbackId, p}
}

func Pend(callback any, port *Port) GoCallback {
	return portMap.ResolvePort(port).Pend(callback)
}

type GoCallback struct {
	id   uint64 // Id (32 bits)
	port *Port
}

var _ _Serializable = GoCallback{}

func (cb GoCallback) specialInt()              {}
func (cb GoCallback) getKind() _SpecialIntKind { return sikGoCallback }
func (cb GoCallback) getPayload() uint64       { return cb.id }

func (gcb *GoCallback) Remove() {
	gcb.port.goCallbacks.Delete(gcb.id)
}

func (gcb *GoCallback) Exists() bool {
	_, loaded := gcb.port.goCallbacks.Load(gcb.id)
	return loaded
}

type invokingGoCallback struct {
	payload uint64 // Flag (16 bits) | Id (32 bits)
	port    *Port
}

func (cb invokingGoCallback) String() string {
	return fmt.Sprintf("invokingGoCallback[payload=%016X, port=%s]", cb.payload, cb.port)
}

func (cb invokingGoCallback) specialInt() {}

func (cb invokingGoCallback) handleCObjects(objs []*Dart_CObject) {
	id := cb.payload & callbackIdMask
	cf := CallbackFlag(cb.payload)

	var (
		callbackV reflect.Value
		loaded    bool
	)

	if cf.HasFallible() {
		defer func() {
			if p := recover(); p != nil {
				log.Printf("%+v\n", p)
			}
		}()
	}

	if cf.HasPop() {
		callbackV, loaded = cb.port.goCallbacks.LoadAndDelete(id)
	} else {
		callbackV, loaded = cb.port.goCallbacks.Load(id)
	}
	if !loaded {
		panic(fmt.Sprintf("dgo:go: go callback not exist, %s", cb))
	}

	hasPackArray := cf.HasPackArray()

	var values []reflect.Value
	var args []any

	if hasPackArray {
		args = make([]any, 0, len(objs)+2)
	} else {
		values = make([]reflect.Value, 0, len(objs)+2)
	}

	if cf.HasWithContext() {
		context := &InvokeContext{cf, cb.port}
		if hasPackArray {
			args = append(args, context)
		} else {
			values = append(values, reflect.ValueOf(context))
		}
	}
	if cf.HasFast() {
		if len(objs) != 0 {
			panic(fmt.Sprintf("dgo:go: expect zero argument when called with FAST flag, %s", cb))
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
		for _, obj := range objs {
			arg := cobjectParse(cb.port, obj)
			if hasPackArray {
				args = append(args, arg)
			} else if arg == nil {
				values = append(values, reflect.ValueOf(&arg).Elem())
			} else {
				values = append(values, reflect.ValueOf(arg))
			}
		}
	}
	if hasPackArray {
		values = []reflect.Value{reflect.ValueOf(args)}
	}
	callbackV.Call(values)
}
