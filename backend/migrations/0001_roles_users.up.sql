CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE roles (
    id          SMALLSERIAL PRIMARY KEY,
    name        VARCHAR(20) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO roles (name) VALUES ('admin'), ('coach');

CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id             SMALLINT NOT NULL REFERENCES roles(id),
    name                VARCHAR(120) NOT NULL,
    email               VARCHAR(160) NOT NULL UNIQUE,
    phone               VARCHAR(20),
    password_hash       VARCHAR(200) NOT NULL,
    fcm_token           TEXT,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_role_id ON users(role_id) WHERE is_active = true;

CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(200) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked     BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id) WHERE revoked = false;

CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id   UUID REFERENCES users(id) ON DELETE SET NULL,
    action          VARCHAR(60) NOT NULL,
    entity_type     VARCHAR(60) NOT NULL,
    entity_id       UUID,
    before_data     JSONB,
    after_data      JSONB,
    ip_address      VARCHAR(64),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_actor_created ON audit_logs(actor_user_id, created_at DESC);
