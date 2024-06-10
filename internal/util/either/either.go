package either

type Either[T any] struct {
	Value T
	Err   error
}

func Of[T any](value T) Either[T] {
	return Either[T]{Value: value}
}

func ErrorOf[T any](err error) Either[T] {
	if err == nil {
		panic("err is nil")
	}
	return Either[T]{Err: err}
}

func Chunked[T any](ch <-chan Either[T], size int) <-chan Either[[]T] {
	out := make(chan Either[[]T])
	go func() {
		defer close(out)

		chunk := make([]T, 0, size)
		for e := range ch {
			if e.Err != nil {
				out <- ErrorOf[[]T](e.Err)
				return
			}

			chunk = append(chunk, e.Value)
			if size <= len(chunk) {
				out <- Of(chunk)
				chunk = make([]T, 0, size)
			}
		}

		if 0 < len(chunk) {
			out <- Of(chunk)
		}
	}()
	return out
}
