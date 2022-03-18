SELECT s2.user_id, sum(c) as count, total, (sum(c) / total) as ratio
FROM (
         SELECT time, user_id, entity, is_write, branch, editor, machine, operating_system, COUNT(time) as c
         FROM heartbeats
         GROUP BY time, user_id, entity, is_write, branch, editor, machine, operating_system
         HAVING COUNT(time) > 1
     ) s2
         LEFT JOIN (SELECT user_id, count(id) AS total FROM heartbeats GROUP BY user_id) s3 ON s2.user_id = s3.user_id
GROUP BY user_id
ORDER BY count DESC;