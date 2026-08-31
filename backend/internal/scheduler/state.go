package scheduler

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// stateStore persists "have we already run job X for period Y" markers in
// the database rather than in process memory, so a restart (deploy, crash,
// horizontal scale-out) can never cause a once-per-day/month job to fire
// again for a period it already handled.
type stateStore struct {
	pool *pgxpool.Pool
}

func newStateStore(pool *pgxpool.Pool) *stateStore {
	return &stateStore{pool: pool}
}

// tryClaim atomically checks whether jobKey has already run for period,
// and if not, records that it now has. Returns true only for the caller
// that should actually run the job. Safe under concurrent callers (e.g. an
// API process and a standalone worker process both running the scheduler)
// because the conditional UPDATE only succeeds for one racer.
func (s *stateStore) tryClaim(ctx context.Context, jobKey, period string) (bool, error) {
	var claimed string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO scheduler_state (job_key, last_period) VALUES ($1, $2)
		ON CONFLICT (job_key) DO UPDATE
			SET last_period = EXCLUDED.last_period, updated_at = now()
			WHERE scheduler_state.last_period IS DISTINCT FROM EXCLUDED.last_period
		RETURNING job_key
	`, jobKey, period).Scan(&claimed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// release removes a claim so a later tick can retry the same period —
// used when the claimed job actually failed, so a transient error (DB
// hiccup, FCM outage) doesn't permanently skip that day/month's report.
func (s *stateStore) release(ctx context.Context, jobKey string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM scheduler_state WHERE job_key = $1`, jobKey)
	return err
}
