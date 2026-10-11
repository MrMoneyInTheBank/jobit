CREATE TABLE IF NOT EXISTS job_applications (
  id INTEGER PRIMARY KEY AUTOINCREMENT,

  company_name TEXT NOT NULL,
  position TEXT NOT NULL,
  application_date TEXT NOT NULL,
  status TEXT NOT NULL CHECK (
    status IN ('submitted', 'interviewing', 'offer', 'rejected', 'accepted')
  ),
  referral INTEGER NOT NULL CHECK (referral in (0, 1)),
  remote TEXT NOT NULL DEFAULT '' CHECK (remote in ('', 'remote', 'hybrid', 'onsite')),

  location TEXT NOT NULL DEFAULT '',
  pay_min INTEGER,
  pay_max INTEGER,
  pay_currency TEXT,
  ranking INTEGER NOT NULL CHECK (ranking BETWEEN 1 AND 5),
  notes TEXT NOT NULL DEFAULT '',
  job_posting_link TEXT NOT NULL DEFAULT '',
  company_website_link TEXT NOT NULL DEFAULT ''
);
