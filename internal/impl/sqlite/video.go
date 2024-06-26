package sqlite

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/wroge/superbasic"
	"google.golang.org/api/youtube/v3"

	"github.com/Durun/m2kt/internal/util"
)

func NewVideoStore(db *sql.DB) *VideoStore {
	s := &VideoStore{
		db: db,
	}
	return s
}

type VideoStore struct {
	prepared bool
	db       *sql.DB
}

func (s *VideoStore) Prepare(ctx context.Context) error {
	if s.prepared {
		return nil
	}

	_, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS videos_raw (
		videoId TEXT PRIMARY KEY,
		json TEXT NOT NULL
	)`)
	if err != nil {
		return errors.WithStack(err)
	}

	s.prepared = true
	return nil
}

func (s *VideoStore) CountVideos(ctx context.Context, ids []string) (int, error) {
	if err := s.Prepare(ctx); err != nil {
		return 0, err
	}

	values := make([]superbasic.Expression, 0, len(ids))
	for _, id := range ids {
		values = append(values, superbasic.Value(id))
	}

	rows, err := util.DoExpr(ctx, s.db.QueryContext, superbasic.Compile(`
		SELECT COUNT(*) FROM videos_raw WHERE videoId IN (?)`,
		superbasic.Join(`,`, values...),
	))
	defer rows.Close()
	if err != nil {
		return 0, err
	}

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
	}
	return count, err
}

func (s *VideoStore) WriteVideos(ctx context.Context, videos []*youtube.SearchResult) error {
	if err := s.Prepare(ctx); err != nil {
		return err
	}

	values := make([]superbasic.Expression, 0, len(videos))
	for _, video := range videos {
		json, err := video.MarshalJSON()
		if err != nil {
			return errors.WithStack(err)
		}
		values = append(values, superbasic.SQL(`(?,?)`,
			video.Id.VideoId,
			string(json),
		))
	}

	_, err := util.DoExpr(ctx, s.db.ExecContext, superbasic.Compile(`
		INSERT INTO videos_raw (videoId, json)
		VALUES ?
		ON CONFLICT(videoId) DO UPDATE SET json=excluded.json`,
		superbasic.Join(`,`, values...),
	))
	return err
}
