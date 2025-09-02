DB_USER=wakapi_user
DB_PASS=sshhh
DB_NAME=wakapi

mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -e "load data infile 'users.tsv' ignore into table users fields terminated by '\t' ignore 1 rows;"
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -e "load data infile 'key_string_values.tsv' ignore into table key_string_values fields terminated by '\t' ignore 1 rows;"
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -e "load data infile 'language_mappings.tsv' ignore into table language_mappings fields terminated by '\t' ignore 1 rows;"
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -e "load data infile 'project_labels.tsv' ignore into table project_labels fields terminated by '\t' ignore 1 rows;"
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -e "load data infile 'aliases.tsv' ignore into table aliases fields terminated by '\t' ignore 1 rows;"
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -e "load data infile 'leaderboard_items.tsv' ignore into table leaderboard_items fields terminated by '\t' ignore 1 rows;"
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -e "load data infile 'summary_items.tsv' ignore into table summary_items fields terminated by '\t' ignore 1 rows;"
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -e "load data infile 'summaries.tsv' ignore into table summaries fields terminated by '\t' ignore 1 rows;"
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -e "load data infile 'durations.tsv' ignore into table durations fields terminated by '\t' ignore 1 rows;"
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -e "load data infile 'heartbeats.tsv' ignore into table heartbeats fields terminated by '\t' ignore 1 rows;"