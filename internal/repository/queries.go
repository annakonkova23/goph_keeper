package repository

var insertUser string = `
		INSERT INTO storage.users (login, password)
		VALUES ($1, $2);
	`
var selectUserLogin string = `SELECT login, password FROM accum_system.users WHERE login = $1`
var insertAuthData string = `
        INSERT INTO storage.auth_data (user_login, site, login, password, meta)
        VALUES ($1, $2, $3, $4, $5::jsonb)
        ON CONFLICT (user_login, site) DO UPDATE
        SET login = EXCLUDED.login,
            password = EXCLUDED.password,
            meta = EXCLUDED.meta`

var insertFileChunk string = `
        INSERT INTO storage.file_data (user_login, file_name, chunk_num, data, meta)
        VALUES ($1, $2, $3, $4, $5::jsonb)
        ON CONFLICT (user_login, file_name, chunk_num) DO UPDATE
        SET data = EXCLUDED.data,
            meta = EXCLUDED.meta`

var insertTextData string = `
        INSERT INTO storage.text_data (user_login, title, data, meta)
        VALUES ($1, $2, $3, $4::jsonb)
        ON CONFLICT (user_login, title) DO UPDATE
        SET data = EXCLUDED.data,
            meta = EXCLUDED.meta`

var insertBankCardData string = `
        INSERT INTO storage.bank_card_data (user_login, last4, numbercard, expmonth, expyear, owner, meta)
        VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
        ON CONFLICT (user_login, numbercard) DO UPDATE
        SET last4 = EXCLUDED.last4,
            expmonth = EXCLUDED.expmonth,
            expyear = EXCLUDED.expyear,
            owner = EXCLUDED.owner,
            meta = EXCLUDED.meta`

var selectAuthData string = `
	 SELECT site, login, password, meta 
     FROM storage.auth_data 
     WHERE user_login = $1 AND (site = $2 OR $2='')
`
var selectFileData string = `
	SELECT file_name, chunk_num, data, meta 
    FROM storage.file_data 
    WHERE user_login = $1 AND file_name = $2 AND chunk_num = $3
`
var selectTextData string = `
	SELECT title, data, meta 
    FROM storage.text_data 
    WHERE user_login = $1 AND (title = $2 OR $2='')
`

var selectBankCardData string = `
	SELECT last4, number_card, exp_month, exp_year, owner, meta 
    FROM storage.bank_card_data 
    WHERE user_login = $1 AND (last4 = $2 OR $2=0)
`
