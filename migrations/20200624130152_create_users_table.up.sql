BEGIN;

CREATE TABLE users (
    id bigserial not null primary key,
    name varchar null,
    email varchar not null unique,
    password varchar null,
    token varchar null,
    role smallint DEFAULT 0,
    email_verified timestamp null,
    crtd_at timestamp null,
    chng_at timestamp null
);

COMMIT;