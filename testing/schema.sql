-- Created with SQLite DB Browser through:
-- File -> Export -> to SQL file
-- with options:
-- Overwrite, Keep original CREATE

BEGIN TRANSACTION;
DROP TABLE IF EXISTS "aliases";
CREATE TABLE `aliases` (`id` integer,`type` integer NOT NULL,`user_id` text NOT NULL,`key` text NOT NULL,`value` text NOT NULL,PRIMARY KEY (`id`),CONSTRAINT `fk_aliases_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE ON UPDATE CASCADE);
DROP TABLE IF EXISTS "diagnostics";
CREATE TABLE `diagnostics` (`id` integer,`user_id` text NOT NULL,`platform` text,`architecture` text,`plugin` text,`cli_version` text,`logs` text,`stack_trace` text,PRIMARY KEY (`id`),CONSTRAINT `fk_diagnostics_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE ON UPDATE CASCADE);
DROP TABLE IF EXISTS "heartbeats";
CREATE TABLE `heartbeats` (`id` integer,`user_id` text NOT NULL,`entity` text NOT NULL,`type` text,`category` text,`project` text,`branch` text,`language` text,`is_write` numeric,`editor` text,`operating_system` text,`machine` text,`user_agent` text,`time` timestamp,`hash` varchar(17),`origin` text,`origin_id` text,`created_at` timestamp,PRIMARY KEY (`id`),CONSTRAINT `fk_heartbeats_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE ON UPDATE CASCADE);
DROP TABLE IF EXISTS "key_string_values";
CREATE TABLE `key_string_values` (`key` text,`value` text,PRIMARY KEY (`key`));
DROP TABLE IF EXISTS "language_mappings";
CREATE TABLE `language_mappings` (`id` integer,`user_id` text NOT NULL,`extension` varchar(16),`language` varchar(64),PRIMARY KEY (`id`),CONSTRAINT `fk_language_mappings_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE ON UPDATE CASCADE);
DROP TABLE IF EXISTS "project_labels";
CREATE TABLE `project_labels` (`id` integer,`user_id` text NOT NULL,`project_key` text,`label` varchar(64),PRIMARY KEY (`id`),CONSTRAINT `fk_project_labels_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE ON UPDATE CASCADE);
DROP TABLE IF EXISTS "summaries";
CREATE TABLE "summaries" (`id` integer,`user_id` text NOT NULL,`from_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,`to_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,PRIMARY KEY (`id`),CONSTRAINT `fk_summaries_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE ON UPDATE CASCADE);
DROP TABLE IF EXISTS "summary_items";
CREATE TABLE `summary_items` (`id` integer,`summary_id` integer,`type` integer,`key` text,`total` integer,PRIMARY KEY (`id`),CONSTRAINT `fk_summaries_editors` FOREIGN KEY (`summary_id`) REFERENCES `summaries`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,CONSTRAINT `fk_summaries_operating_systems` FOREIGN KEY (`summary_id`) REFERENCES `summaries`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,CONSTRAINT `fk_summaries_machines` FOREIGN KEY (`summary_id`) REFERENCES `summaries`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,CONSTRAINT `fk_summaries_projects` FOREIGN KEY (`summary_id`) REFERENCES `summaries`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,CONSTRAINT `fk_summaries_languages` FOREIGN KEY (`summary_id`) REFERENCES `summaries`(`id`) ON DELETE CASCADE ON UPDATE CASCADE);
DROP TABLE IF EXISTS "users";
CREATE TABLE `users` (`id` text,`api_key` text UNIQUE,`email` text,`location` text,`password` text,`created_at` timestamp DEFAULT CURRENT_TIMESTAMP,`last_logged_in_at` timestamp DEFAULT CURRENT_TIMESTAMP,`share_data_max_days` integer DEFAULT 0,`share_editors` numeric DEFAULT false,`share_languages` numeric DEFAULT false,`share_projects` numeric DEFAULT false,`share_oss` numeric DEFAULT false,`share_machines` numeric DEFAULT false,`share_labels` numeric DEFAULT false,`is_admin` numeric DEFAULT false,`has_data` numeric DEFAULT false,`wakatime_api_key` text,`reset_token` text,`reports_weekly` numeric DEFAULT false,PRIMARY KEY (`id`));
DROP INDEX IF EXISTS "idx_alias_type_key";
CREATE INDEX `idx_alias_type_key` ON `aliases`(`type`,`key`);
DROP INDEX IF EXISTS "idx_alias_user";
CREATE INDEX `idx_alias_user` ON `aliases`(`user_id`);
DROP INDEX IF EXISTS "idx_diagnostics_user";
CREATE INDEX `idx_diagnostics_user` ON `diagnostics`(`user_id`);
DROP INDEX IF EXISTS "idx_entity";
CREATE INDEX `idx_entity` ON `heartbeats`(`entity`);
DROP INDEX IF EXISTS "idx_heartbeats_hash";
CREATE UNIQUE INDEX `idx_heartbeats_hash` ON `heartbeats`(`hash`);
DROP INDEX IF EXISTS "idx_language";
CREATE INDEX `idx_language` ON `heartbeats`(`language`);
DROP INDEX IF EXISTS "idx_language_mapping_composite";
CREATE UNIQUE INDEX `idx_language_mapping_composite` ON `language_mappings`(`user_id`,`extension`);
DROP INDEX IF EXISTS "idx_language_mapping_user";
CREATE INDEX `idx_language_mapping_user` ON `language_mappings`(`user_id`);
DROP INDEX IF EXISTS "idx_project_label_user";
CREATE INDEX `idx_project_label_user` ON `project_labels`(`user_id`);
DROP INDEX IF EXISTS "idx_time";
CREATE INDEX `idx_time` ON `heartbeats`(`time`);
DROP INDEX IF EXISTS "idx_time_summary_user";
CREATE INDEX `idx_time_summary_user` ON `summaries`(`user_id`,`from_time`,`to_time`);
DROP INDEX IF EXISTS "idx_time_user";
CREATE INDEX `idx_time_user` ON `heartbeats`(`user_id`);
DROP INDEX IF EXISTS "idx_type";
CREATE INDEX `idx_type` ON `summary_items`(`type`);
DROP INDEX IF EXISTS "idx_user_email";
CREATE INDEX `idx_user_email` ON `users`(`email`);
COMMIT;
