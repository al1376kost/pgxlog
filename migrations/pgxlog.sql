CREATE SCHEMA adm;

CREATE TABLE adm.logs (
	add_date_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	level_id INTEGER NOT NULL,
	message TEXT NULL,
	message_data JSONB
);

-- for timescaledb uncomment
-- SELECT create_hypertable('adm.logs', 'add_date_time', chunk_time_interval => interval '1 hour');

-- indexes
CREATE INDEX ON adm.logs (add_date_time);
CREATE INDEX ON adm.logs (level_id);

-- for tests
CREATE TABLE adm.logs_test (
	add_date_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	level_id INTEGER NOT NULL,
	message TEXT NULL,
	message_data JSONB
);
