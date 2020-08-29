-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

alter table users
    add `machine` varchar(255);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back

alter table users
    drop column `machine`;