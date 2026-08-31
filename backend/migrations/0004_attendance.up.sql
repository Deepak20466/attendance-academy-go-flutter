CREATE TABLE student_attendance (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    class_id     UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    student_id   UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    activity_id  UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    status       VARCHAR(20) NOT NULL, -- present/absent/late/excused
    marked_by    UUID NOT NULL REFERENCES users(id),
    marked_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    remarks      TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (class_id, student_id)
);

-- Monthly percentage / history queries filter by student + date range.
CREATE INDEX idx_student_attendance_student ON student_attendance(student_id, marked_at DESC);
CREATE INDEX idx_student_attendance_activity_class ON student_attendance(activity_id, class_id);
CREATE INDEX idx_student_attendance_class ON student_attendance(class_id);

CREATE TABLE coach_attendance (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id            UUID NOT NULL REFERENCES coaches(id) ON DELETE CASCADE,
    class_id            UUID REFERENCES classes(id) ON DELETE SET NULL,
    attendance_date     DATE NOT NULL,
    check_in_time       TIMESTAMPTZ,
    check_in_lat        NUMERIC(9,6),
    check_in_lng        NUMERIC(9,6),
    check_in_verified   BOOLEAN NOT NULL DEFAULT false,
    check_out_time      TIMESTAMPTZ,
    check_out_lat       NUMERIC(9,6),
    check_out_lng       NUMERIC(9,6),
    check_out_verified  BOOLEAN NOT NULL DEFAULT false,
    status              VARCHAR(20) NOT NULL DEFAULT 'present',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (coach_id, class_id)
);

CREATE INDEX idx_coach_attendance_coach_date ON coach_attendance(coach_id, attendance_date DESC);
CREATE INDEX idx_coach_attendance_date ON coach_attendance(attendance_date);
