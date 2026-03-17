package repository

var insertUser string = `
		INSERT INTO storage.users (id, login, password)
		VALUES ($1, $2, $3);
	`
var selectUserLogin string = `SELECT id, password FROM storage.users WHERE login = $1`

var insertAuthData string = `
       	INSERT INTO storage.auth_data(
			id, user_id,login,password, meta
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING version`

var selectAuthData string = `
	 SELECT login, password, meta, version, updated_at
     FROM storage.auth_data 
     WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
`

var updateAuthData string = `
	 UPDATE  storage.auth_data 
		SET 
		    login = $1,
		    password = $2,
		    meta = $3,
		    version = version + 1,
		    updated_at = now()
		WHERE id = $4
		  AND user_id = $5
		  AND version = $6
		  AND deleted_at IS NULL
		RETURNING version
`

var deleteAuthData string = `
	 UPDATE storage.auth_data 
		SET deleted_at = now(),
		    version = version + 1,
		    updated_at = now()
		WHERE id = $1
		  AND user_id = $2
		  AND version = $3
		  AND deleted_at IS NULL
 `

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
