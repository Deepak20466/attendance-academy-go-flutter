CREATE TABLE batches (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id         UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    name                VARCHAR(120) NOT NULL,
    description         TEXT,
    default_coach_id    UUID REFERENCES coaches(id) ON DELETE SET NULL,
    location_id         UUID REFERENCES locations(id) ON DELETE SET NULL,
    days_of_week        SMALLINT[] NOT NULL DEFAULT '{}', -- 0=Sun .. 6=Sat
    start_time          TIME NOT NULL,
    end_time            TIME NOT NULL,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_batches_activity ON batches(activity_id) WHERE is_active = true;
CREATE INDEX idx_batches_default_coach ON batches(default_coach_id);

CREATE TABLE students (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id        UUID NOT NULL REFERENCES batches(id) ON DELETE RESTRICT,
    activity_id     UUID NOT NULL REFERENCES activities(id) ON DELETE RESTRICT,
    name            VARCHAR(120) NOT NULL,
    phone           VARCHAR(20),
    guardian_name   VARCHAR(120),
    guardian_phone  VARCHAR(20),
    email           VARCHAR(160),
    date_of_birth   DATE,
    joining_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_students_activity ON students(activity_id) WHERE is_active = true;
CREATE INDEX idx_students_batch ON students(batch_id) WHERE is_active = true;

-- A concrete scheduled session of a batch on a specific date.
CREATE TABLE classes (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id                UUID NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    activity_id             UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    coach_id                UUID NOT NULL REFERENCES coaches(id) ON DELETE RESTRICT,
    class_date              DATE NOT NULL,
    start_time              TIME NOT NULL,
    end_time                TIME NOT NULL,
    status                  VARCHAR(20) NOT NULL DEFAULT 'scheduled', -- scheduled/completed/cancelled
    attendance_marked       BOOLEAN NOT NULL DEFAULT false,
    attendance_marked_at    TIMESTAMPTZ,
    reminder_sent_at        TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (batch_id, class_date, start_time)
);

CREATE INDEX idx_classes_activity_date ON classes(activity_id, class_date);
CREATE INDEX idx_classes_coach_date ON classes(coach_id, class_date);
-- Fast lookup for the 15-min reminder & EOD job: classes past end_time with no attendance yet.
CREATE INDEX idx_classes_pending_attendance ON classes(class_date, end_time)
    WHERE attendance_marked = false AND status = 'scheduled';
