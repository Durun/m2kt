package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/wroge/superbasic"
	"google.golang.org/api/youtube/v3"

	"github.com/Durun/m2kt/internal/util"
	"github.com/Durun/m2kt/pkg/chu"
)

func NewRawStore(db *sql.DB) *RawStore {
	s := &RawStore{
		db: db,
	}
	return s
}

type RawStore struct {
	prepared bool
	db       *sql.DB
}

func (s *RawStore) Prepare(ctx context.Context) error {
	if s.prepared {
		return nil
	}

	sqls := []string{
		`CREATE TABLE IF NOT EXISTS videos_raw (
			videoId TEXT PRIMARY KEY,
			json TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS channels_raw (
			channelId TEXT PRIMARY KEY,
			fetchedAt TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			json TEXT NOT NULL
		)`,
	}

	for _, query := range sqls {
		_, err := s.db.ExecContext(ctx, query)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	s.prepared = true
	return nil
}

func (s *RawStore) CountVideos(ctx context.Context, ids []string) (int, error) {
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

func (s *RawStore) WriteVideos(ctx context.Context, videos []*youtube.SearchResult) error {
	if err := s.Prepare(ctx); err != nil {
		return err
	}

	values := make([]superbasic.Expression, 0, len(videos))
	for _, video := range videos {
		videoJson, err := video.MarshalJSON()
		if err != nil {
			return errors.WithStack(err)
		}
		values = append(values, superbasic.SQL(`(?,?)`,
			video.Id.VideoId,
			string(videoJson),
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

func (s *RawStore) DumpVideos(ctx context.Context) chu.ReadChan[*youtube.SearchResult] {
	return chu.GenerateContext(ctx, func(ctx context.Context, out chu.WriteChan[*youtube.SearchResult]) {
		rows, err := util.DoExpr(ctx, s.db.QueryContext, superbasic.Compile(`
			SELECT json FROM videos_raw`,
		))
		if err != nil {
			out.PushError(err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var videoJson string
			if err := rows.Scan(&videoJson); err != nil {
				out.PushError(errors.WithStack(err))
				break
			}

			video := new(youtube.SearchResult)
			if err := json.Unmarshal([]byte(videoJson), video); err != nil {
				out.PushError(errors.WithStack(err))
				break
			}

			out.PushValue(video)
		}
	})
}

func (s *RawStore) WriteChannels(ctx context.Context, fetchedAt time.Time, channels []*youtube.Channel) error {
	if err := s.Prepare(ctx); err != nil {
		return err
	}

	values := make([]superbasic.Expression, 0, len(channels))
	for _, channel := range channels {
		bytes, err := channel.MarshalJSON()
		if err != nil {
			return errors.WithStack(err)
		}

		fetchedAtStr := fetchedAt.UTC().Format(time.DateTime)
		values = append(values, superbasic.SQL(`(?,?,?)`,
			channel.Id,
			fetchedAtStr,
			string(bytes),
		))
	}

	_, err := util.DoExpr(ctx, s.db.ExecContext, superbasic.Compile(`
		INSERT INTO channels_raw (channelId, fetchedAt, json)
		VALUES ?
		ON CONFLICT(channelId) DO UPDATE SET json=excluded.json, fetchedAt=excluded.fetchedAt`,
		superbasic.Join(`,`, values...),
	))
	return err
}

func (s *RawStore) DumpChannels(ctx context.Context) chu.ReadChan[*youtube.Channel] {
	return chu.GenerateContext(ctx, func(ctx context.Context, out chu.WriteChan[*youtube.Channel]) {
		rows, err := util.DoExpr(ctx, s.db.QueryContext, superbasic.Compile(`
		SELECT json FROM channels_raw`,
		))
		if err != nil {
			out.PushError(errors.WithStack(err))
			return
		}
		defer rows.Close()

		for rows.Next() {
			var jsonStr string
			if err := rows.Scan(&jsonStr); err != nil {
				out.PushError(errors.WithStack(err))
				break
			}

			value := new(youtube.Channel)
			if err := json.Unmarshal([]byte(jsonStr), value); err != nil {
				out.PushError(errors.WithStack(err))
				break
			}

			out.PushValue(value)
		}
	})
}

func (s *RawStore) ListChannelIDs(ctx context.Context, channelIDs []string) ([]string, error) {
	rows, err := util.DoExpr(ctx, s.db.QueryContext, superbasic.Compile(`
		SELECT channelId
		FROM channels_raw
		WHERE channelId IN (?)`,
		superbasic.Join(`,`,
			superbasic.Map(channelIDs,
				func(_ int, id string) superbasic.Expression { return superbasic.Value(id) },
			)...,
		),
	))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0, len(channelIDs))
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, errors.WithStack(err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
