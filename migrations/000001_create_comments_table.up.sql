CREATE TABLE IF NOT EXISTS credentials (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    email_address text NOT NULL,
    name text NOT NULL,
    version integer NOT NULL DEFAULT 1
);