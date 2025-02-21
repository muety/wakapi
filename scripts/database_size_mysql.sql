SELECT table_schema                                            name,
       ROUND(SUM(data_length + index_length) / 1024 / 1024, 1) size
FROM information_schema.tables
GROUP BY table_schema
ORDER BY size DESC;