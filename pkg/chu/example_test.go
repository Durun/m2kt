package chu

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func ExampleReadChan() {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	r := exampleGenerate(ctx)
	r = exampleMap(r)
	defer r.RequestClose()

	for v := range r.Chan() {
		value, err := v.Get()
		if err != nil {
			fmt.Println("received", err, ": request to close")
			r.RequestClose()
			continue
		}

		fmt.Println("received", value)
	}

	// Output:
	// sending {value1}
	// received VALUE1
	// sending {value2}
	// received VALUE2
	// sending ERROR3
	// received ERROR3 : request to close
	// close request detected
	// received context canceled : request to close
}

func exampleGenerate(ctx context.Context) ReadChan[string] {
	i := 0
	generate := func() Either[string] {
		i++
		if i%3 == 0 {
			return ErrorOf[string](fmt.Errorf("ERROR%d", i))
		}

		return ValueOf(fmt.Sprintf("value%d", i))
	}

	return GenerateContext(ctx, func(ctx context.Context, out WriteChan[string]) {
		for {
			select {
			case <-ctx.Done():
				out.PushError(errors.WithStack(ctx.Err()))
				fmt.Println("close request detected")
				return
			default:
			}

			item := generate()
			fmt.Println("sending", item)
			out.Chan() <- item
			time.Sleep(time.Millisecond)
		}
	})
}

func exampleMap(ch ReadChan[string]) ReadChan[string] {
	return Map(ch, func(in string) (string, error) {
		return strings.ToUpper(in), nil
	})
}
