-- +goose Up
-- +goose StatementBegin
CREATE TABLE notification_setting
(
  id                  serial NOT NULL,
  email_notifications bool NOT NULL DEFAULT true,
  PRIMARY KEY (id)
);

CREATE TABLE refresh_sessions
(
  token      uuid      NOT NULL,
  expires_at timestamp NOT NULL,
  user_id    serial      NOT NULL,
  PRIMARY KEY (token)
);

CREATE TABLE users
(
  id            serial NOT NULL,
  name          text NOT NULL,
  email         text NOT NULL,
  avatar_url    text,
  password_hash text NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE artists
(
  label text
) INHERITS (users);

CREATE TABLE users_users
(
  followed_id serial NOT NULL,
  follower_id serial NOT NULL
);

ALTER TABLE users_users
  ADD CONSTRAINT FK_users_TO_users_users
    FOREIGN KEY (followed_id)
    REFERENCES users (id);

ALTER TABLE users_users
  ADD CONSTRAINT FK_users_TO_users_users1
    FOREIGN KEY (follower_id)
    REFERENCES users (id);

ALTER TABLE refresh_sessions
  ADD CONSTRAINT FK_users_TO_refresh_sessions
    FOREIGN KEY (user_id)
    REFERENCES users (id);

ALTER TABLE notification_setting
  ADD CONSTRAINT FK_users_TO_notification_setting
    FOREIGN KEY (id)
    REFERENCES users (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE refresh_sessions;
DROP TABLE notification_setting;
DROP TABLE users_users;
DROP TABLE artists;
DROP TABLE users;
-- +goose StatementEnd
