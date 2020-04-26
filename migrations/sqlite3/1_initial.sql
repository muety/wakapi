-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
create table aliases
(
    id      integer primary key autoincrement,
    type    integer      not null,
    user_id varchar(255) not null,
    key     varchar(255) not null,
    value   varchar(255) not null
);

create index idx_alias_type_key
    on aliases (type, key);

create index idx_alias_user
    on aliases (user_id);

create table summaries
(
    id        integer primary key autoincrement,
    user_id   varchar(255) not null,
    from_time timestamp default CURRENT_TIMESTAMP not null,
    to_time   timestamp default CURRENT_TIMESTAMP not null
);

create index idx_time_summary_user
    on summaries (user_id, from_time, to_time);

create table summary_items
(
    id         integer primary key autoincrement,
    summary_id integer REFERENCES summaries (id) ON DELETE CASCADE ON UPDATE CASCADE,
    type       integer,
    key        varchar(255),
    total      bigint
);

create table users
(
    id       varchar(255) primary key,
    api_key  varchar(255) unique,
    password varchar(255)
);

create table heartbeats
(
    id               integer primary key autoincrement,
    user_id          varchar(255) not null REFERENCES users (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    entity           varchar(255) not null,
    type             varchar(255),
    category         varchar(255),
    project          varchar(255),
    branch           varchar(255),
    language         varchar(255),
    is_write         bool,
    editor           varchar(255),
    operating_system varchar(255),
    time             timestamp default CURRENT_TIMESTAMP
);

create index idx_entity
    on heartbeats (entity);

create index idx_language
    on heartbeats (language);

create index idx_time
    on heartbeats (time);

create index idx_time_user
    on heartbeats (user_id, time);



-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP INDEX idx_alias_user;
DROP INDEX idx_alias_type_key;
DROP TABLE aliases;
DROP INDEX idx_time_summary_user;
DROP TABLE summaries;
DROP TABLE summary_items;
DROP TABLE heartbeats;
DROP INDEX idx_entity;
DROP INDEX idx_language;
DROP INDEX idx_time;
DROP INDEX idx_time_user;