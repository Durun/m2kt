package chu

import "context"

func (c ReadChan[T]) OnEach(f func(T) error) <-chan error {
	out := Map(c, func(v T) (struct{}, error) {
		return struct{}{}, f(v)
	})
	out = out.FilterError(func(_ struct{}, err error) bool {
		return err != nil
	})
	_, errs := out.Unwrap()
	return errs
}

func (c ReadChan[T]) ForEachCloseOnError(f func(T) error) []error {
	var out []error
	for err := range c.OnEach(f) {
		c.RequestClose()
		out = append(out, err)
	}
	return out
}

func (c ReadChan[T]) Unwrap() (<-chan T, <-chan error) {
	values := make(chan T)
	errs := make(chan error)
	go func() {
		defer close(values)
		defer close(errs)

		for v := range c.ch {
			value, err := v.Get()
			if err != nil {
				errs <- err
				continue
			}

			values <- value
		}
	}()
	return values, errs
}

func (c ReadChan[T]) WaitClose(ctx context.Context) []Either[T] {
	c.RequestClose()

	var out []Either[T]
	for {
		select {
		case v := <-c.ch:
			out = append(out, v)
		case <-ctx.Done():
			return out
		default:
			return out
		}
	}
}
