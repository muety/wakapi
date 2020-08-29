-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

alter table heartbeats
    add column `machine` varchar(255);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back

alter table heartbeats
    drop column `machine`;