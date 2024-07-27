package chu

import "context"

func NewChan[T any]() WriteChan[T] {
	return WriteChan[T]{
		ch:     make(chan Either[T]),
		cancel: make(chan struct{}),
	}
}

type WriteChan[T any] struct {
	ch     chan Either[T]
	cancel chan struct{}
}

func (c WriteChan[T]) Done() <-chan struct{} {
	return c.cancel
}

func (c WriteChan[T]) RequestClose() {
	select {
	case <-c.cancel:
		return
	default:
		close(c.cancel)
	}
}

func (c WriteChan[T]) Chan() chan<- Either[T] {
	return c.ch
}

func (c WriteChan[T]) WithContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-c.Done():
			cancel()
		}
	}()
	return ctx
}

func (c WriteChan[T]) IsDoneContext(ctx context.Context) bool {
	select {
	default:
		return false
	case <-ctx.Done():
	case <-c.Done():
	}
	return true
}
func (c WriteChan[T]) IsDone() bool {
	return c.IsDoneContext(context.Background())
}

func (c WriteChan[T]) PushValue(v T) {
	c.ch <- ValueOf(v)
}
func (c WriteChan[T]) PushError(err error) {
	c.ch <- ErrorOf[T](err)
}

func (c WriteChan[T]) Reader() ReadChan[T] {
	return ReadChan[T]{
		ch:     c.ch,
		cancel: c.cancel,
	}
}

type ReadChan[T any] struct {
	ch     <-chan Either[T]
	cancel chan struct{}
}

func (c ReadChan[T]) Chan() <-chan Either[T] {
	return c.ch
}

func (c ReadChan[T]) RequestClose() {
	select {
	case <-c.cancel:
		return
	default:
		close(c.cancel)
	}
}
