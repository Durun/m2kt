package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

func NewSQLiteDB(file string) (*sql.DB, error) {
	db, err := sql.Open("sqlite",
		fmt.Sprintf(`file:%s`, file),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return db, nil
}
