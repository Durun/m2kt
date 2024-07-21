package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/wroge/superbasic"

	"github.com/Durun/m2kt/internal/entity"
	"github.com/Durun/m2kt/internal/util"
	"github.com/Durun/m2kt/internal/util/either"
)

func NewIndexedStore(db *sql.DB) *IndexedStore {
	s := &IndexedStore{
		db: db,
	}
	return s
}

type IndexedStore struct {
	prepared bool
	db       *sql.DB
}

func (s *IndexedStore) Prepare(ctx context.Context) error {
	if s.prepared {
		return nil
	}

	sqls := []string{
		`CREATE TABLE IF NOT EXISTS videos (
			videoId TEXT PRIMARY KEY,
			channelId TEXT NOT NULL,
			weekday INT NOT NULL,
			date TEXT NOT NULL,
			time TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			thumbnailURL TEXT NOT NULL,
			eTag TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS videos_channelId ON videos (channelId)`,
		`CREATE INDEX IF NOT EXISTS videos_datetime ON videos (date, time)`,
		`CREATE INDEX IF NOT EXISTS videos_weekday ON videos (weekday, time)`,
		`CREATE INDEX IF NOT EXISTS videos_time ON videos (time)`,
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

func (s *IndexedStore) WriteVideos(ctx context.Context, videos []entity.Video) error {
	if err := s.Prepare(ctx); err != nil {
		return err
	}

	_, err := util.DoExpr(ctx, s.db.ExecContext, superbasic.Compile(`
		INSERT INTO videos (videoId, channelId, weekday, date, time, title, description, thumbnailURL, eTag)
		VALUES ?
		ON CONFLICT(videoId) DO UPDATE SET
			channelId=excluded.channelId,
			weekday=excluded.weekday,
			date=excluded.date,
			time=excluded.time,
			title=excluded.title,
			description=excluded.description,
			thumbnailURL=excluded.thumbnailURL,
			eTag=excluded.eTag`,
		superbasic.Join(`,`, superbasic.Map(videos, func(_ int, video entity.Video) superbasic.Expression {
			t := video.PublishedAt.UTC()
			return superbasic.SQL(`(?,?,?,?,?,?,?,?,?)`,
				video.VideoID,
				video.ChannelID,
				t.Weekday(),
				t.Format(time.DateOnly),
				t.Format(time.TimeOnly),
				video.Title,
				video.Description,
				video.ThumbnailURL,
				video.ETag,
			)
		})...),
	))
	return err
}

func (s *IndexedStore) DumpVideos(ctx context.Context, where, orderby string) <-chan either.Either[entity.Video] {
	ch := make(chan either.Either[entity.Video])

	go func() {
		rows, err := util.DoExpr(ctx, s.db.QueryContext, superbasic.Compile(`
			SELECT videoId, channelId, date, time, title, description, thumbnailURL, eTag
			FROM videos
			? ?`,
			superbasic.If(where != "", superbasic.SQL(`WHERE `+where)),
			superbasic.If(orderby != "", superbasic.SQL(`ORDER BY `+orderby)),
		))
		if err != nil {
			ch <- either.ErrorOf[entity.Video](errors.WithStack(err))
			close(ch)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var dateStr, timeStr string
			var video entity.Video
			err := rows.Scan(
				&video.VideoID,
				&video.ChannelID,
				&dateStr,
				&timeStr,
				&video.Title,
				&video.Description,
				&video.ThumbnailURL,
				&video.ETag,
			)
			if err != nil {
				ch <- either.ErrorOf[entity.Video](errors.WithStack(err))
				break
			}

			video.PublishedAt, err = time.Parse(time.DateTime, dateStr+" "+timeStr)
			if err != nil {
				ch <- either.ErrorOf[entity.Video](errors.WithStack(err))
				break
			}

			ch <- either.Of(video)
		}

		close(ch)
	}()

	return ch
}
