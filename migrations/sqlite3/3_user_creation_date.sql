-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

-- SQLite does not allow altering a table to add a new column with default of CURRENT_TIMESTAMP
-- See https://www.sqlite.org/lang_altertable.html

alter table users
    add `created_at` timestamp default '2020-01-01T00:00:00.000' not null;

alter table users
    add `last_logged_in_at` timestamp default '2020-01-01T00:00:00.000' not null;

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back

alter table users
    drop column `created_at`;

alter table users
    drop column `last_logged_in_at`;