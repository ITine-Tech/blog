ALTER TABLE
    IF EXISTS users
ADD column role_id INT REFERENCES roles(id) DEFAULT 1;

UPDATE
    users
SET
    role_id = (
    SELECT
        id
    FROM
        roles
    WHERE
        name = 'user'
    );

ALTER TABLE users
ALTER COLUMN
    role_id DROP DEFAULT;

ALTER TABLE users
    ALTER column role_id
SET NOT NULL;