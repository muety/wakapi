-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

create table custom_rules
(
    id               integer primary key autoincrement,
    user_id          varchar(255) not null REFERENCES users (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    extension        varchar(255),
    language         varchar(255)
);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE custom_rules;
