package chu

type Either[T any] interface {
	Get() (T, error)
}

func ErrorOf[T any](err error) Error[T] {
	return Error[T]{err}
}

type Error[T any] struct {
	err error
}

func (v Error[T]) Get() (T, error) {
	var zero T
	return zero, v.err
}

func (v Error[T]) String() string {
	return v.err.Error()
}

func ValueOf[T any](value T) Ok[T] {
	return Ok[T]{value}
}

type Ok[T any] struct {
	value T
}

func (v Ok[T]) Get() (T, error) {
	return v.value, nil
}

func EitherOf[T any](value T, err error) Either[T] {
	if err != nil {
		return ErrorOf[T](err)
	}
	return ValueOf[T](value)
}
