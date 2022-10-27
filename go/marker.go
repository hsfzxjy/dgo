package dgo

type Optional[T any] struct {
	Value                 T
	isSome                bool
	_dgo_optional_marker_ struct{} //lint:ignore U1000 A special marker
}

func Some[T any](value T) Optional[T] { return Optional[T]{Value: value, isSome: true} }
func None[T any]() Optional[T]        { return Optional[T]{} }

func (o *Optional[T]) IsSome() bool { return o.isSome }
func (o *Optional[T]) IsNone() bool { return !o.isSome }
func (o *Optional[T]) Clear() {
	var zero T
	o.Value = zero
	o.isSome = false
}
func (o *Optional[T]) Set(value T) {
	o.Value = value
	o.isSome = true
}
func (o *Optional[T]) SetZero() {
	var zero T
	o.Value = zero
	o.isSome = true
}
