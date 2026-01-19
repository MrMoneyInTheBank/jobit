package store

const insertQuery string = `
		INSERT INTO job_applications (
			company_name,
			position,
			application_date,
			status,
			referral,
			remote,
			location,
			pay_min,
			pay_max,
			pay_currency,
			ranking,
			notes,
			job_positing_link,
			company_website_link
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

const getAllQuery string = `
		SELECT
			id,
			company_name,
			position,
			application_date,
			status,
			referral,
			remote,
			location,
			pay_min,
			pay_max,
			pay_currency,
			ranking,
			notes,
			job_positing_link,
			company_website_link
		FROM job_applications
	`

const getByIDQuery string = `
		SELECT 
				id,
			company_name,
			position,
			application_date,
			status,
			referral,
			remote,
			location,
			pay_min,
			pay_max,
			pay_currency,
			ranking,
			notes,
			job_positing_link,
			company_website_link
		FROM job_applications
		WHERE id = ?
		LIMIT 1
	`

const updateQuery string = `
	UPDATE job_applications
		SET
			company_name = ?,
			position = ?,
			application_date = ?,
			status = ?,
			referral = ?,
			remote = ?,
			location = ?,
			pay_min = ?,
			pay_max = ?,
			pay_currency = ?,
			ranking = ?,
			notes = ?,
			job_positing_link = ?,
			company_website_link = ?
		WHERE id = ?
	`

const deleteByIDQuery string = `
		DELETE FROM job_applications
		WHERE id = ?
	`
