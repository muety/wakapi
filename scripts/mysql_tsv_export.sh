DB_USER=wakapi_user
DB_PASS=sshhh
DB_NAME=wakapi

mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -q -e "select * from users" | pigz -c > users.gz.tsv
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -q -e "select * from key_string_values" | pigz -c > key_string_values.gz.tsv
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -q -e "select * from language_mappings" | pigz -c > language_mappings.gz.tsv
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -q -e "select * from project_labels" | pigz -c > project_labels.gz.tsv
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -q -e "select * from aliases" | pigz -c > aliases.gz.tsv
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -q -e "select * from leaderboard_items" | pigz -c > leaderboard_items.gz.tsv
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -q -e "select * from summary_items" | pigz -c > summary_items.gz.tsv
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -q -e "select * from summaries" | pigz -c > summaries.gz.tsv
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -q -e "select * from durations" | pigz -c > durations.gz.tsv
mysql -u $DB_USER -D $DB_NAME -p$DB_PASS -q -e "select * from heartbeats" | pigz -c > heartbeats.gz.tsv
