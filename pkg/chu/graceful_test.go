package chu

import (
	"sync"
	"testing"
)

func TestGracefulReadChan_RequestClose(t *testing.T) {
	t.Parallel()

	t.Run("parallel", func(t *testing.T) {
		t.Parallel()
		w := NewChan[string]()
		r := w.Reader()

		wg := sync.WaitGroup{}
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func() {
				for i := 0; i < 1000; i++ {
					r.RequestClose()
				}
				wg.Done()
			}()
		}

		<-w.Done()
		wg.Wait()
	})
}
