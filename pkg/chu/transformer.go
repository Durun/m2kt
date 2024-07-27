package chu

func NewPipe[OUT, IN any](upstream ReadChan[IN]) WriteChan[OUT] {
	return WriteChan[OUT]{
		ch:     make(chan Either[OUT]),
		cancel: upstream.cancel,
	}
}

func Transform[OUT, IN any](in ReadChan[IN], f func(out WriteChan[OUT])) ReadChan[OUT] {
	out := NewPipe[OUT](in)
	go func() {
		defer close(out.ch)

		f(out)
	}()

	return out.Reader()
}

func MapError[OUT, IN any](in ReadChan[IN], f func(IN, error) (OUT, error)) ReadChan[OUT] {
	return Transform(in, func(out WriteChan[OUT]) {
		for v := range in.ch {
			out.ch <- EitherOf(f(v.Get()))
		}
	})
}

func Map[OUT, IN any](in ReadChan[IN], f func(IN) (OUT, error)) ReadChan[OUT] {
	return MapError(in, func(v IN, err error) (OUT, error) {
		if err != nil {
			var zero OUT
			return zero, err
		}
		return f(v)
	})
}

func (c ReadChan[T]) FilterError(f func(T, error) bool) ReadChan[T] {
	return Transform(c, func(out WriteChan[T]) {
		for v := range c.ch {
			value, err := v.Get()
			if f(value, err) {
				out.ch <- v
				continue
			}
		}
	})
}

func (c ReadChan[T]) Filter(f func(T) bool) ReadChan[T] {
	return c.FilterError(func(v T, err error) bool {
		return err != nil || f(v)
	})
}

func (c ReadChan[T]) Chunked(size uint) <-chan Either[[]T] { // stop T instantiation cycle
	out := make(chan Either[[]T])
	go func() {
		defer close(out)
		sendChunked[T](out, c.ch, size)
	}()
	return out
}

func Chunked[T any](in ReadChan[T], size uint) ReadChan[[]T] {
	return Transform(in, func(out WriteChan[[]T]) {
		sendChunked[T](out.ch, in.ch, size)
	})
}

func sendChunked[T any](out chan<- Either[[]T], in <-chan Either[T], size uint) {
	iSize := int(size)

	var chunk []T
	for v := range in {
		value, err := v.Get()
		if err != nil {
			out <- ErrorOf[[]T](err)
			continue
		}

		if chunk == nil {
			chunk = make([]T, 0, size)
		}
		chunk = append(chunk, value)
		if len(chunk) >= iSize {
			out <- ValueOf(chunk)
			chunk = nil
		}
	}
	if len(chunk) > 0 {
		out <- ValueOf(chunk)
	}
}

func Flatten[T any, S []T](in ReadChan[S]) ReadChan[T] {
	return Transform(in, func(out WriteChan[T]) {
		for v := range in.ch {
			value, err := v.Get()
			if err != nil {
				out.ch <- ErrorOf[T](err)
				continue
			}

			for _, item := range value {
				out.ch <- ValueOf(item)
			}
		}
	})
}
