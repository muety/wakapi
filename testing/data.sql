BEGIN TRANSACTION;
INSERT INTO "key_string_values" ("key","value") VALUES ('20210213-add_has_data_field','done');
INSERT INTO "key_string_values" ("key","value") VALUES ('20210221-add_created_date_column','done');
INSERT INTO "key_string_values" ("key","value") VALUES ('imprint','no content here');
INSERT INTO "key_string_values" ("key","value") VALUES ('20210411-add_imprint_content','done');
INSERT INTO "key_string_values" ("key","value") VALUES ('20210806-remove_persisted_project_labels','done');
INSERT INTO "key_string_values" ("key","value") VALUES ('20211215-migrate_id_to_bigint-add_has_data_field','done');
INSERT INTO "key_string_values" ("key","value") VALUES ('20212212-total_summary_heartbeats','done');
INSERT INTO "key_string_values" ("key","value") VALUES ('20220317-align_num_heartbeats','done');
INSERT INTO "key_string_values" ("key","value") VALUES ('20220318-mysql_timestamp_precision','done');
INSERT INTO "key_string_values" ("key","value") VALUES ('202203191-drop_diagnostics_user','done');
COMMIT;

BEGIN TRANSACTION;
INSERT INTO "users" ("id", "api_key", "email", "location", "password", "created_at", "last_logged_in_at",
                     "share_data_max_days", "share_editors", "share_languages", "share_projects", "share_oss",
                     "share_machines", "is_admin", "has_data", "wakatime_api_key", "reset_token", "reports_weekly")
VALUES ('readuser', '33e7f538-0dce-4eba-8ffe-53db6814ed42', '', 'Europe/Berlin',
        '$2a$10$93CAptdjLGRtc1D3xrZJcu8B/YBAPSjCZOHZRId.xpyrsLAeHOoA.', '2021-05-28 12:34:25',
        '2021-05-28 14:34:34.178+02:00', 0, 0, 0, 0, 0, 0, 1, 0, '', '', 0);
INSERT INTO "users" ("id", "api_key", "email", "location", "password", "created_at", "last_logged_in_at",
                     "share_data_max_days", "share_editors", "share_languages", "share_projects", "share_oss",
                     "share_machines", "is_admin", "has_data", "wakatime_api_key", "reset_token", "reports_weekly")
VALUES ('writeuser', 'f7aa255c-8647-4d0b-b90f-621c58fd580f', '', 'Europe/Berlin',
        '$2a$10$93CAptdjLGRtc1D3xrZJcu8B/YBAPSjCZOHZRId.xpyrsLAeHOoA.', '2021-05-28 12:34:56',
        '2021-05-28 14:35:05.118+02:00', 7, 0, 0, 1, 0, 0, 0, 1, '', '', 0);

INSERT INTO "api_keys" ("api_key", "user_id", "label", "read_only")
VALUES
    ('1c91f670-2309-45fb-9d7e-738c766e85a6', 'writeuser', 'Full Access Key', false),
    ('774f7e16-b9a3-433e-ac68-7e28a82a50ca', 'writeuser', 'Read Only Key', true);
COMMIT;