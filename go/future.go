package dgo

type DartFutureCallback DartCallback

func (dcb DartCallback) AsFut() DartFutureCallback {
	return DartFutureCallback(dcb)
}

func (dcb DartFutureCallback) Complete(val any, err error) bool {
	if err == nil {
		return DartCallback(dcb).
			Flag(CF_POP | CF_WITHCODE | cf_fut_resolve).
			Call(val)
	} else {
		return DartCallback(dcb).
			Flag(CF_POP | CF_WITHCODE | cf_fut_reject).
			Call(err.Error())
	}
}

func (dcb DartFutureCallback) Resolve(val any) bool {
	return dcb.Complete(val, nil)
}

func (dcb DartFutureCallback) Reject(err error) bool {
	if err == nil {
		panic("dgo:go expect non-nil error argument for DartFutureCallback.Reject()")
	}
	return dcb.Complete(nil, err)
}
