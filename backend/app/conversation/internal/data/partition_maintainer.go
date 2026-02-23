package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultPartitionMonthsAhead    = 12
	defaultPartitionEnsureInterval = 24 * time.Hour
	partitionEnsureTimeout         = 10 * time.Second
)

func ensureFuturePartitions(ctx context.Context, db *pgxpool.Pool, monthsAhead int) error {
	if monthsAhead <= 0 {
		monthsAhead = defaultPartitionMonthsAhead
	}
	_, err := db.Exec(ctx, `SELECT ensure_next_month_partitions($1)`, monthsAhead)
	return err
}

func startPartitionMaintainer(db *pgxpool.Pool, logger *log.Helper, interval time.Duration, monthsAhead int) func() {
	if interval <= 0 {
		interval = defaultPartitionEnsureInterval
	}
	if monthsAhead <= 0 {
		monthsAhead = defaultPartitionMonthsAhead
	}

	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runCtx, runCancel := context.WithTimeout(context.Background(), partitionEnsureTimeout)
				if err := ensureFuturePartitions(runCtx, db, monthsAhead); err != nil {
					logger.Warnf("ensure future partitions failed: %v", err)
				}
				runCancel()
			}
		}
	}()

	return cancel
}
