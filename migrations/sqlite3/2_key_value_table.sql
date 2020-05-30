-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
create table key_string_values
(
    key      varchar(255) primary key,
    value    text
);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
drop table key_string_value;