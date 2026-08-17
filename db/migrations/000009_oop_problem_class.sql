ALTER TABLE problem_versions
    DROP CONSTRAINT problem_versions_problem_class_check;
ALTER TABLE problem_versions
    ADD CONSTRAINT problem_versions_problem_class_check
        CHECK (problem_class IN ('algorithms', 'concurrency', 'oop'));

ALTER TABLE invitations
    DROP CONSTRAINT invitations_problem_class_check;
ALTER TABLE invitations
    ADD CONSTRAINT invitations_problem_class_check
        CHECK (problem_class IN ('algorithms', 'concurrency', 'oop'));

ALTER TABLE matches
    DROP CONSTRAINT matches_problem_class_check;
ALTER TABLE matches
    ADD CONSTRAINT matches_problem_class_check
        CHECK (problem_class IN ('algorithms', 'concurrency', 'oop'));
