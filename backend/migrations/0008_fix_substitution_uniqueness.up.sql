-- The original unconditional UNIQUE(class_id) permanently blocked ever
-- reassigning a substitute for a class again after the first substitution
-- was cancelled, since the cancelled row still occupied the constraint.
-- Only one ACTIVE substitution per class should ever be enforced.
ALTER TABLE substitutions DROP CONSTRAINT substitutions_class_id_key;

CREATE UNIQUE INDEX idx_substitutions_active_class ON substitutions(class_id) WHERE status = 'active';
