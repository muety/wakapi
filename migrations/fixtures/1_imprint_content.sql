-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
insert into key_string_values ("key", "value") values ('imprint', 'no content here');

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
SET SQL_MODE=ANSI_QUOTES;
delete from key_string_values where key = 'imprint';