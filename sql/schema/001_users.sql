-- +goose Up
CREATE TABLE users (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  email TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE users;

--Start postgres: sudo systemctl start postgresql
--goose postgres postgres://postgres:postgres@localhost:5432/chirpy up
--goose postgres postgres://postgres:postgres@localhost:5432/chirpy down
--Log into postgres: sudo -u postgres psql
--Connect to chirpy: \c chirpy
--List data tables: \dt
