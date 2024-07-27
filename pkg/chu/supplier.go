package chu

import "context"

func GenerateContext[T any](ctx context.Context, f func(ctx context.Context, out WriteChan[T])) ReadChan[T] {
	out := NewChan[T]()
	ctx = out.WithContext(ctx)
	go func() {
		defer close(out.ch)
		f(ctx, out)
	}()
	return out.Reader()
}

func Generate[T any](f func(ctx context.Context, out WriteChan[T])) ReadChan[T] {
	return GenerateContext(context.Background(), f)
}

func FromSlice[T any](s []T) ReadChan[T] {
	return Generate(func(ctx context.Context, out WriteChan[T]) {
		for _, v := range s {
			select {
			case <-ctx.Done():
				out.PushError(ctx.Err())
				return
			default:
				select {
				case <-ctx.Done():
				case out.ch <- ValueOf(v):
				}
			}
		}
	})
}
