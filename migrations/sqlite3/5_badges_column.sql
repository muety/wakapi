-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

alter table users
    add column `badges_enabled` tinyint(1) default 0 not null;

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back

alter table users
    drop column `badges_enabled`;