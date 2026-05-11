-- +goose Up
create table refresh_tokens(
    token TEXT primary key,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    user_id UUID  NOT NULL REFERENCES users ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP
);

-- +goose Down
DROP table refresh_tokens;