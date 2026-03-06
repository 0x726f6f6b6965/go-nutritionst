CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    line_id TEXT UNIQUE,
    height NUMERIC(10, 2),
    weight NUMERIC(10, 2),
    age INTEGER,
    gender SMALLINT,
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
    meals INT,
    total_calories_kcal NUMERIC(10, 2),
    total_protein_g NUMERIC(10, 2),
    total_carbs_g NUMERIC(10, 2),
    total_fat_g NUMERIC(10, 2),
    total_sodium_mg NUMERIC(10, 2),
    compliance_calories VARCHAR(10),
    compliance_protein VARCHAR(10),
    compliance_sodium VARCHAR(10),
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
    status SMALLINT,
    fail_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);