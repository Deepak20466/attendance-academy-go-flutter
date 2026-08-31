CREATE TABLE activities (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(120) NOT NULL UNIQUE,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE coaches (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    employee_code   VARCHAR(40) NOT NULL UNIQUE,
    joining_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    monthly_salary  NUMERIC(10,2) NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Authorization source of truth: which activities a coach may access.
CREATE TABLE coach_activities (
    coach_id     UUID NOT NULL REFERENCES coaches(id) ON DELETE CASCADE,
    activity_id  UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (coach_id, activity_id)
);

CREATE INDEX idx_coach_activities_activity ON coach_activities(activity_id);

CREATE TABLE locations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id     UUID REFERENCES activities(id) ON DELETE CASCADE,
    name            VARCHAR(120) NOT NULL,
    latitude        NUMERIC(9,6) NOT NULL,
    longitude       NUMERIC(9,6) NOT NULL,
    radius_meters   INTEGER NOT NULL DEFAULT 100,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_locations_activity ON locations(activity_id);
