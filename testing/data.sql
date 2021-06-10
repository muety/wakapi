BEGIN TRANSACTION;
INSERT INTO "users" ("id", "api_key", "email", "location", "password", "created_at", "last_logged_in_at",
                     "share_data_max_days", "share_editors", "share_languages", "share_projects", "share_oss",
                     "share_machines", "is_admin", "has_data", "wakatime_api_key", "reset_token", "reports_weekly")
VALUES ('readuser', '33e7f538-0dce-4eba-8ffe-53db6814ed42', '', 'Europe/Berlin',
        '$2a$10$RCyfAFdlZdFJVWbxKz4f2uJ/MospiE1EFAIjvRizC4Nop9GfjgKzW', '2021-05-28 12:34:25',
        '2021-05-28 14:34:34.178+02:00', 0, 0, 0, 0, 0, 0, 1, 0, '', '', 0);
INSERT INTO "users" ("id", "api_key", "email", "location", "password", "created_at", "last_logged_in_at",
                     "share_data_max_days", "share_editors", "share_languages", "share_projects", "share_oss",
                     "share_machines", "is_admin", "has_data", "wakatime_api_key", "reset_token", "reports_weekly")
VALUES ('writeuser', 'f7aa255c-8647-4d0b-b90f-621c58fd580f', '', 'Europe/Berlin',
        '$2a$10$vsksPpiXZE9/xG9pRrZP.eKkbe/bGWW4wpPoXqvjiImZqMbN5c4Km', '2021-05-28 12:34:56',
        '2021-05-28 14:35:05.118+02:00', 7, 0, 0, 1, 0, 0, 0, 1, '', '', 0);
COMMIT;
