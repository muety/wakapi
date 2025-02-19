SELECT project, language, editor, operating_system, machine, branch, SUM(diff) as 'sum'
FROM (SELECT project,
             language,
             editor,
             operating_system,
             machine,
             branch,
             TIME_TO_SEC(LEAST(TIMEDIFF(time, LAG(time) over w), '00:00:00')) as 'diff'  -- time constant ~ heartbeats padding (none by default, formerly 2 mins)
      FROM heartbeats
      WHERE user_id = 'n1try'
      WINDOW w AS (ORDER BY time)) s2
WHERE diff IS NOT NULL
GROUP BY project, language, editor, operating_system, machine, branch;