CREATE SCHEMA IF NOT EXISTS storage AUTHORIZATION current_user;
   

GRANT USAGE, CREATE ON SCHEMA storage TO current_user;

CREATE TABLE IF NOT EXISTS storage.users (
    login VARCHAR(1000) PRIMARY KEY NOT NULL,
    password VARCHAR(1000) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
 );

CREATE TABLE IF NOT EXISTS storage.auth_data (
    user_login VARCHAR(1000),
    site VARCHAR(1000),
    login VARCHAR(1000) NOT NULL,
    password VARCHAR(1000) NOT NULL,
    meta JSON,
    created_at TIMESTAMP DEFAULT NOW(),
    primary key (user_login, site),
    constraint uq_auth_data unique(user_login, site)
)

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