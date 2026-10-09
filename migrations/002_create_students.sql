CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    nim VARCHAR(20) NOT NULL UNIQUE,
    nama VARCHAR(255) NOT NULL,
    prodi VARCHAR(100) NOT NULL,
    angkatan INT NOT NULL,
    ipk_terakhir NUMERIC(3,2) NOT NULL DEFAULT 0 CHECK (ipk_terakhir >= 0 AND ipk_terakhir <= 4),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);
