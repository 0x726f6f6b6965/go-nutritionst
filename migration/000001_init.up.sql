CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    line_id TEXT UNIQUE,
    height NUMERIC(10, 2),
    weight NUMERIC(10, 2),
    target_weight NUMERIC(10, 2),
    target_timeframe INTEGER,
    age INTEGER,
    gender SMALLINT,
    max_daily_token BIGINT DEFAULT 0,
    morning_msg_sent BOOLEAN DEFAULT FALSE,
    evening_msg_sent BOOLEAN DEFAULT FALSE,
    breakfast_msg_sent BOOLEAN DEFAULT FALSE,
    lunch_msg_sent BOOLEAN DEFAULT FALSE,
    dinner_msg_sent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS meal_history (
    id BIGSERIAL PRIMARY KEY,
    request_id UUID,
    line_id TEXT,
    meal SMALLINT,
    description TEXT,
    photo BYTEA,
    calories_kcal NUMERIC(10, 2),
    protein_g NUMERIC(10, 2),
    carbs_g NUMERIC(10, 2),
    fat_g NUMERIC(10, 2),
    sodium_mg NUMERIC(10, 2),
    ai_description TEXT,
    ai_warnings TEXT,
    ai_suggest TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_histories_line_id FOREIGN KEY (line_id) REFERENCES users (line_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS meal_daily (
    id BIGSERIAL PRIMARY KEY,
    request_id UUID,
    line_id TEXT,
    date TIMESTAMP,
    breakfast_meals INT DEFAULT 0,
    lunch_meals INT DEFAULT 0,
    dinner_meals INT DEFAULT 0,
    snack_meals INT DEFAULT 0,
    total_calories_kcal NUMERIC(10, 2),
    total_protein_g NUMERIC(10, 2),
    total_carbs_g NUMERIC(10, 2),
    total_fat_g NUMERIC(10, 2),
    total_sodium_mg NUMERIC(10, 2),
    compliance_calories TEXT,
    compliance_protein TEXT,
    compliance_sodium TEXT,
    compliance_deltas_calories_kcal NUMERIC(10, 2),
    compliance_deltas_protein_g NUMERIC(10, 2),
    compliance_deltas_sodium_mg NUMERIC(10, 2),
    insights TEXT [],
    today_coaching TEXT [],
    data_quality_issues TEXT [],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_dailies_line_id FOREIGN KEY (line_id) REFERENCES users (line_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS send_requests (
    id BIGSERIAL PRIMARY KEY,
    request_id UUID,
    request_type SMALLINT,
    line_id TEXT,
    status SMALLINT,
    data TEXT,
    fail_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS used_tokens (
    id BIGSERIAL PRIMARY KEY,
    line_id TEXT UNIQUE,
    used_token BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_used_tokens_line_id FOREIGN KEY (line_id) REFERENCES users (line_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS daily_record (
    id BIGSERIAL PRIMARY KEY,
    request_id UUID,
    line_id TEXT,
    date TEXT,
    total_water_ml NUMERIC(10, 2) DEFAULT NULL,
    total_sleep_hour NUMERIC(10, 2) DEFAULT NULL,
    breakfast_meals INT DEFAULT 0,
    lunch_meals INT DEFAULT 0,
    dinner_meals INT DEFAULT 0,
    snack_meals INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE (line_id, date),
    CONSTRAINT fk_daily_record_line_id FOREIGN KEY (line_id) REFERENCES users (line_id) ON DELETE CASCADE
);

-- create a function to update updated_at on used_tokens once the row is updated
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_used_tokens_updated_at
BEFORE UPDATE ON used_tokens
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_send_requests_updated_at
BEFORE UPDATE ON send_requests
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_daily_record_updated_at
BEFORE UPDATE ON daily_record
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();