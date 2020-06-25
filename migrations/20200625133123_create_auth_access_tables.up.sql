BEGIN;

CREATE TABLE auth_tokens (
    id bigserial not null primary key,
    user_id bigint not null,
    access_token varchar null,
    refresh_token varchar null,
    remember BOOLEAN DEFAULT FALSE,
    crtd_at timestamp null,
    constraint fk_users_t
     foreign key (user_id) 
     REFERENCES users (id)
);

CREATE TABLE auth_access (
    id bigserial not null primary key,
    token_id bigint not null,
    secret varchar not null,
    constraint fk_accs_s
     foreign key (token_id) 
     REFERENCES auth_tokens (id)
);

COMMIT;