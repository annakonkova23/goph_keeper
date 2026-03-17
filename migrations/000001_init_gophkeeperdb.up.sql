CREATE SCHEMA IF NOT EXISTS storage AUTHORIZATION current_user;
   

GRANT USAGE, CREATE ON SCHEMA storage TO current_user;

CREATE TABLE IF NOT EXISTS storage.users (
    id UUID PRIMARY KEY,
    login VARCHAR(1000) NOT NULL UNIQUE,,
    password VARCHAR(1000) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
 );


CREATE TABLE IF NOT EXISTS storage.auth_data (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    login VARCHAR(1000) NOT NULL,
    password BYTEA NOT NULL,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    version BIGINT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)

CREATE INDEX idx_auth_data_user_id ON storage.auth_data(user_id);
CREATE INDEX idx_auth_data_alive ON storage.auth_data(user_id, id) WHERE deleted_at IS NULL;


CREATE TABLE storage.file_data (
    user_login VARCHAR(1000),
    file_name varchar(1000),
    chunk_num INT NOT NULL,
    data BYTEA NOT NULL,
    meta JSON,
    created_at TIMESTAMP DEFAULT NOW(),
    primary key (user_login, file_name, chunk_num),
    constraint uq_file_data unique(user_login, file_name,chunk_num)
);


CREATE TABLE storage.text_data (
    user_login VARCHAR(1000),
    title varchar(1000),
    data text NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    meta JSON
    primary key (user,  title),
    constraint uq_file_data unique(user_login, file_name)
);

CREATE TABLE storage.bank_card_data (
    user_login VARCHAR(1000),
    last4 int,
    number_card varchar(1000),
    exp_month int NOT NULL,
    exp_year int NOT NULL,
    owner varchar(1000),
    meta JSON,
    primary key (user_login, number_card),
    constraint uq_bank_card_data unique(user_login, number_card)
);