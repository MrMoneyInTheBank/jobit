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
