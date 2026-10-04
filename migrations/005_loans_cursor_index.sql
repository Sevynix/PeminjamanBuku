CREATE INDEX IF NOT EXISTS loans_borrowed_at_id_desc_idx
    ON loans (borrowed_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS loans_user_borrowed_at_id_desc_idx
    ON loans (user_id, borrowed_at DESC, id DESC);