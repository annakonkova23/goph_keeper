CREATE SCHEMA IF NOT EXISTS storage AUTHORIZATION current_user;
   

GRANT USAGE, CREATE ON SCHEMA storage TO current_user;

CREATE TABLE IF NOT EXISTS storage.users (
    id UUID PRIMARY KEY,
    login VARCHAR(1000) NOT NULL UNIQUE,
    password VARCHAR(1000) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
 );


CREATE TABLE IF NOT EXISTS storage.auth_data (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    login VARCHAR(1000) NOT NULL,
    password BYTEA NOT NULL,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    version BIGINT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS  idx_auth_data_user_id ON storage.auth_data(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_data_alive ON storage.auth_data(user_id, id) WHERE deleted_at IS NULL;




CREATE TABLE IF NOT EXISTS storage.text_data (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    data text NOT NULL,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    version BIGINT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_text_data_user_id ON storage.text_data(user_id);
CREATE INDEX IF NOT EXISTS idx_text_data_alive ON storage.text_data(user_id, id) WHERE deleted_at IS NULL;


CREATE TABLE IF NOT EXISTS storage.bank_card_data (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    last4 int,
    number_card BYTEA NOT NULL,
    exp_month int NOT NULL,
    exp_year int NOT NULL,
    owner varchar(1000),
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    version BIGINT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_bank_card_data_user_id ON storage.bank_card_data(user_id);
CREATE INDEX IF NOT EXISTS idx_bank_card_data_alive ON storage.bank_card_data(user_id, id) WHERE deleted_at IS NULL;


CREATE TABLE IF NOT EXISTS storage.files (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    current_version BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS storage.uploads (
    id UUID PRIMARY KEY,
    file_id UUID NOT NULL,
    user_id UUID NOT NULL,
    expected_version BIGINT NOT NULL,
    filename TEXT NOT NULL,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL CHECK (status IN ('pending', 'committed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS storage.upload_chunks (
    upload_id UUID NOT NULL,
    chunk_no BIGINT NOT NULL,
    data BYTEA NOT NULL,
    size_bytes BIGINT NOT NULL,
    PRIMARY KEY (upload_id, chunk_no)
);

CREATE TABLE IF NOT EXISTS storage.file_versions (
    file_id UUID NOT NULL,
    version BIGINT NOT NULL,
    filename TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    checksum TEXT NOT NULL DEFAULT '',
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (file_id, version)
);

CREATE TABLE IF NOT EXISTS storage.file_chunks (
    file_id UUID NOT NULL,
    version BIGINT NOT NULL,
    chunk_no BIGINT NOT NULL,
    data BYTEA NOT NULL,
    size_bytes BIGINT NOT NULL,
    PRIMARY KEY (file_id, version, chunk_no)
);