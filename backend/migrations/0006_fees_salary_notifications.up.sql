CREATE TABLE fees (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    activity_id     UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    amount          NUMERIC(10,2) NOT NULL,
    due_date        DATE NOT NULL,
    paid_date       DATE,
    payment_method  VARCHAR(40),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending/paid/overdue/waived
    period_month    SMALLINT NOT NULL,
    period_year     SMALLINT NOT NULL,
    remarks         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (student_id, period_month, period_year)
);

CREATE INDEX idx_fees_status_due ON fees(status, due_date);
CREATE INDEX idx_fees_activity_period ON fees(activity_id, period_year, period_month);

CREATE TABLE salary_acknowledgements (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id          UUID NOT NULL REFERENCES coaches(id) ON DELETE CASCADE,
    period_month      SMALLINT NOT NULL,
    period_year       SMALLINT NOT NULL,
    amount            NUMERIC(10,2),
    status            VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending/acknowledged
    acknowledged_at   TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (coach_id, period_month, period_year)
);

CREATE INDEX idx_salary_ack_status ON salary_acknowledgements(status);

CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(160) NOT NULL,
    body        TEXT NOT NULL,
    type        VARCHAR(40) NOT NULL,
    data        JSONB,
    is_read     BOOLEAN NOT NULL DEFAULT false,
    sent_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read, created_at DESC);
