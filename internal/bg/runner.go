package bg

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/robfig/cron/v3"
)

type Runner struct {
	cron        *cron.Cron
	logger      *slog.Logger
	bucketStore namemyserver.BucketStore
}

func NewRunner(logger *slog.Logger, bucketStore namemyserver.BucketStore) *Runner {
	r := &Runner{
		cron:        cron.New(cron.WithLogger(&cronLogger{Logger: logger.With(slog.String("service", "cron"))})),
		logger:      logger,
		bucketStore: bucketStore,
	}
	r.setup()

	return r
}

func (r *Runner) setup() {
	r.cron.AddFunc("0 * * * *", r.task("remove_archived_buckets", removeArchivedBucketsTask(r.logger, r.bucketStore)))
}

func (r *Runner) task(name string, f func(context.Context) error) func() {
	if !isSnakeCase(name) {
		panic(fmt.Sprintf("invalid cron task name '%s', task names must be snake_case.", name))
	}

	return func() {
		r.logger.Info("starting task", slog.String("name", name))
		now := time.Now()
		defer func() {
			r.logger.Info("ending task",
				slog.String("name", name),
				slog.Duration("elapsed", time.Since(now)),
			)
		}()

		ctx := context.Background()
		if err := f(ctx); err != nil {
			r.logger.Error("failure running task", slog.Any("err", err), slog.String("task", name))
		}
	}
}

func (r *Runner) Start() {
	r.cron.Start()
}

type cronLogger struct {
	*slog.Logger
}

func (l *cronLogger) Error(err error, msg string, keysAndValues ...any) {
	kvs := make([]any, 0, len(keysAndValues)+1)
	kvs = append(kvs, slog.Any("err", err))
	kvs = append(kvs, keysAndValues...)

	l.Logger.Error(msg, kvs...)
}

var snakeCaseRegexp = regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*$`)

func isSnakeCase(s string) bool {
	return snakeCaseRegexp.MatchString(s)
}
