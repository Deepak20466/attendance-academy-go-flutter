DROP INDEX IF EXISTS idx_substitutions_active_class;
ALTER TABLE substitutions ADD CONSTRAINT substitutions_class_id_key UNIQUE (class_id);
