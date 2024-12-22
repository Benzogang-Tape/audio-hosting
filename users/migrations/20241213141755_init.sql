-- +goose Up
-- +goose StatementBegin

CREATE TABLE artists
(
  label   text NOT NULL DEFAULT '',
  user_id uuid NOT NULL
);

CREATE TABLE listeners
(
  user_id uuid NOT NULL
);

CREATE TABLE notifications_settings
(
  email_notifications bool NOT NULL DEFAULT true,
  user_id             uuid NOT NULL,
  PRIMARY KEY (user_id)
);

CREATE TABLE refresh_sessions
(
  token      uuid      NOT NULL,
  expires_at timestamp NOT NULL,
  user_id    uuid      NOT NULL,
  PRIMARY KEY (token)
);

CREATE TABLE users
(
  id            uuid NOT NULL,
  name          text NOT NULL,
  email         text NOT NULL UNIQUE,
  avatar_url    text,
  password_hash text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE users_users
(
  follower_id uuid      NOT NULL,
  followed_id uuid      NOT NULL,
  id          bigserial NOT NULL,
  PRIMARY KEY (id)
);

ALTER TABLE artists
  ADD CONSTRAINT FK_users_TO_artists
    FOREIGN KEY (user_id)
    REFERENCES users (id);

ALTER TABLE listeners
  ADD CONSTRAINT FK_users_TO_listeners
    FOREIGN KEY (user_id)
    REFERENCES users (id);

ALTER TABLE notifications_settings
  ADD CONSTRAINT FK_users_TO_notifications_settings
    FOREIGN KEY (user_id)
    REFERENCES users (id);

ALTER TABLE users_users
  ADD CONSTRAINT FK_users_TO_users_users
    FOREIGN KEY (follower_id)
    REFERENCES users (id);

ALTER TABLE users_users
  ADD CONSTRAINT FK_users_TO_users_users1
    FOREIGN KEY (followed_id)
    REFERENCES users (id);

ALTER TABLE refresh_sessions
  ADD CONSTRAINT FK_users_TO_refresh_sessions
    FOREIGN KEY (user_id)
    REFERENCES users (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE refresh_sessions;
DROP TABLE notifications_settings;
DROP TABLE users_users;
DROP TABLE listeners;
DROP TABLE artists;
DROP TABLE users;
-- +goose StatementEnd
