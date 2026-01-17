package db

var insertQuery string = `
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

var getAllQuery string = `
		SELECT *
		FROM job_applications
	`

var getByIDQuery string = `
		SELECT *
		FROM job_applications
		WHERE id = ?
		LIMIT 1
	`

var updateQuery string = `
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
