package repository

var insertUser string = `
		INSERT INTO storage.users (id, login, password)
		VALUES ($1, $2, $3);
	`
var selectUserLogin string = `SELECT id, password FROM storage.users WHERE login = $1`

var insertTextData string = `
       	INSERT INTO storage.text_data(
			id, user_id, data, meta
		)
		VALUES ($1, $2, $3, $4)
		RETURNING version`

var selectTextData string = `
	 SELECT data, meta, version, updated_at
     FROM storage.text_data 
     WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
`

var updateTextData string = `
	 UPDATE  storage.text_data 
		SET 
		    data = $1
		    meta = $2,
		    version = version + 1,
		    updated_at = now()
		WHERE id = $3
		  AND user_id = $4
		  AND version = $5
		  AND deleted_at IS NULL
		RETURNING version
`

var deleteTextData string = `
	 UPDATE storage.text_data 
		SET deleted_at = now(),
		    version = version + 1,
		    updated_at = now()
		WHERE id = $1
		  AND user_id = $2
		  AND version = $3
		  AND deleted_at IS NULL
 `

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

var insertBankCardData string = `
       	INSERT INTO storage.bank_card_data(
			id, user_id, last4, number_card, exp_month, exp_year, owner, meta 
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING version`

var selectBankCardDataData string = `
	 SELECT last4, number_card, exp_month, exp_year, owner, meta  version, updated_at
     FROM storage.bank_card_data
     WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
`

var updateBankCardData string = `
	 UPDATE  storage.auth_data 
		SET 
		    last4 = $1, 
            number_card = $2, 
            exp_month = $3, 
            exp_year = $4, 
            owner = $5,
		    meta = $6,
		    version = version + 1,
		    updated_at = now()
		WHERE id = $7
		  AND user_id = $8
		  AND version = $9
		  AND deleted_at IS NULL
		RETURNING version
`

var deleteBankCardData string = `
	 UPDATE storage.bank_card_data 
		SET deleted_at = now(),
		    version = version + 1,
		    updated_at = now()
		WHERE id = $1
		  AND user_id = $2
		  AND version = $3
		  AND deleted_at IS NULL
 `
