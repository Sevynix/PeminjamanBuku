CREATE TABLE IF NOT EXISTS loans (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER     NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    book_id     INTEGER     NOT NULL REFERENCES books(id) ON DELETE RESTRICT,
    borrowed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_at      TIMESTAMPTZ NOT NULL,
    returned_at TIMESTAMPTZ,
    CONSTRAINT loans_due_after_borrowed CHECK (due_at > borrowed_at)
);

CREATE UNIQUE INDEX IF NOT EXISTS loans_one_active_per_book_key
    ON loans (user_id, book_id)
    WHERE returned_at IS NULL;

CREATE INDEX IF NOT EXISTS loans_book_id_idx ON loans (book_id);

INSERT INTO permissions (name, description) VALUES
    ('loan:list', 'Melihat daftar seluruh pinjaman'),
    ('loan:read:any', 'Melihat pinjaman milik siapa pun'),
    ('loan:return', 'Memproses pengembalian buku')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'loan:list'),
    ('admin', 'loan:read:any'),
    ('admin', 'loan:return')
ON CONFLICT DO NOTHING;