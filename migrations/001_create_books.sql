CREATE TABLE IF NOT EXISTS books (
    id         SERIAL PRIMARY KEY,
    title      VARCHAR(200) NOT NULL,
    stock      INTEGER      NOT NULL CHECK (stock >= 0),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS books_title_lower_key
    ON books (LOWER(title));