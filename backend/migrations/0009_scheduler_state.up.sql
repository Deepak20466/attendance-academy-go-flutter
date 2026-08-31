-- Persists "have we already run job X for period Y" markers so the
-- scheduler's once-per-day/once-per-month jobs survive process restarts.
-- Previously tracked only in an in-memory struct field, which meant every
-- redeploy on the last day of the month (or the 10th, for salary) re-sent
-- duplicate admin notifications.
CREATE TABLE scheduler_state (
    job_key      VARCHAR(50) PRIMARY KEY,
    last_period  VARCHAR(20) NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
