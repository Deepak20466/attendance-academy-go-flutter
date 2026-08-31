DO $$
DECLARE
    t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY[
        'users','activities','coaches','batches','students','classes',
        'student_attendance','coach_attendance','locations','substitutions',
        'leaves','fees','salary_acknowledgements'
    ] LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS trg_%I_updated_at ON %I;', t, t);
    END LOOP;
END $$;

DROP FUNCTION IF EXISTS set_updated_at();
