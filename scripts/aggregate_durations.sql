SELECT project, language, editor, operating_system, machine, branch, SUM(GREATEST(1, diff)) as 'sum'
FROM (
         SELECT project, language, editor, operating_system, machine, branch, TIME_TO_SEC(LEAST(TIMEDIFF(time, LAG(time) over w), '00:02:00')) as 'diff'
         FROM heartbeats
         WHERE user_id = 'n1try'
             WINDOW w AS (ORDER BY time)
     ) s2
WHERE diff IS NOT NULL
GROUP BY project, language, editor, operating_system, machine, branch;