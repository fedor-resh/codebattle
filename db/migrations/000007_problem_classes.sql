ALTER TABLE problem_versions
    ADD COLUMN problem_class text NOT NULL DEFAULT 'algorithms',
    ADD COLUMN requirements jsonb NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE problem_versions
    ADD CONSTRAINT problem_versions_problem_class_check
        CHECK (problem_class IN ('algorithms', 'concurrency')),
    ADD CONSTRAINT problem_versions_requirements_object_check
        CHECK (jsonb_typeof(requirements) = 'object');

ALTER TABLE invitations
    ADD COLUMN problem_class text NOT NULL DEFAULT 'algorithms';

ALTER TABLE invitations
    ADD CONSTRAINT invitations_problem_class_check
        CHECK (problem_class IN ('algorithms', 'concurrency'));

ALTER TABLE matches
    ADD COLUMN problem_class text NOT NULL DEFAULT 'algorithms';

UPDATE matches AS m
SET problem_class = problem.problem_class
FROM problem_versions AS problem
WHERE problem.id = m.problem_version_id;

ALTER TABLE matches
    ADD CONSTRAINT matches_problem_class_check
        CHECK (problem_class IN ('algorithms', 'concurrency'));
