CREATE TABLE if not exists roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    level int NOT NULL DEFAULT 1,
    description TEXT
);


INSERT INTO
    roles (name, description, level)
VALUES
    (
        'user',
        'A user can create posts and comments',
        1
    );

INSERT INTO
    roles (name, description, level)
VALUES
    (
        'admin',
        'An admin can update and delete other users comments, posts and can delete users',
        3
    );