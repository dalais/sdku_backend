BEGIN;

CREATE TABLE profiles (
    id bigserial not null primary key,
    user_id bigint null,
    firstname varchar null,
    lastname varchar null,
    photo varchar null,
    phone varchar null,
    crtd_at timestamp with time zone NULL DEFAULT (current_timestamp AT TIME ZONE 'UTC'),
    chng_at timestamp with time zone NULL,
    constraint fk_users_profiles
     foreign key (user_id) 
     REFERENCES users (id)
);

COMMIT;