BEGIN;

CREATE TABLE profiles (
    id bigserial not null primary key,
    user_id bigint null,
    firstname varchar null,
    lastname varchar null,
    photo varchar null,
    phone varchar null,
    crtd_at timestamp null,
    chng_at timestamp null,
    constraint fk_users_profiles
     foreign key (user_id) 
     REFERENCES users (id)
);

COMMIT;