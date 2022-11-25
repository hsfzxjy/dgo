package dgo

type DartStreamCallback DartCallback

func (dcb DartStreamCallback) specialInt() {}

func (dcb DartStreamCallback) Add(value any) bool {
	return DartCallback(dcb).
		Flag(CF_WITHCONTEXT | cf_stream_value).
		Call(value)
}

func (dcb DartStreamCallback) AddError(err error) bool {
	return DartCallback(dcb).
		Flag(CF_WITHCONTEXT | cf_stream_error).
		Call(err.Error())
}

func (dcb DartStreamCallback) Done() bool {
	return DartCallback(dcb).
		Flag(CF_POP | CF_WITHCONTEXT | CF_FAST_NIL).
		Call()
}
