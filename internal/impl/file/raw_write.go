package file

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func WriteJSONs[T any](enc *json.Encoder, entities []T) error {
	for _, entity := range entities {
		if err := enc.Encode(entity); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
