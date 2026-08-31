-- A substitution row is the authorization scope for the substitute coach:
-- they may only act on this exact class while status = 'active'.
CREATE TABLE substitutions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    class_id            UUID NOT NULL UNIQUE REFERENCES classes(id) ON DELETE CASCADE,
    original_coach_id   UUID NOT NULL REFERENCES coaches(id),
    substitute_coach_id UUID NOT NULL REFERENCES coaches(id),
    authorized_by       UUID NOT NULL REFERENCES users(id),
    reason              TEXT,
    status              VARCHAR(20) NOT NULL DEFAULT 'active', -- active/cancelled
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (original_coach_id <> substitute_coach_id)
);

CREATE INDEX idx_substitutions_substitute ON substitutions(substitute_coach_id) WHERE status = 'active';
CREATE INDEX idx_substitutions_class ON substitutions(class_id);

CREATE TABLE leaves (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id      UUID NOT NULL REFERENCES coaches(id) ON DELETE CASCADE,
    start_date    DATE NOT NULL,
    end_date      DATE NOT NULL,
    reason        TEXT NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending/approved/rejected/cancelled
    reviewed_by   UUID REFERENCES users(id),
    reviewed_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date >= start_date)
);

CREATE INDEX idx_leaves_coach ON leaves(coach_id, status);
CREATE INDEX idx_leaves_status ON leaves(status) WHERE status = 'pending';
