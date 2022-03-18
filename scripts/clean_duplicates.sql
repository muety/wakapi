DELETE t1
FROM heartbeats t1
         INNER JOIN heartbeats t2
WHERE t1.id < t2.id
  AND t1.time = t2.time
  AND t1.entity = t2.entity
  AND t1.is_write = t2.is_write
  AND t1.branch = t2.branch
  AND t1.editor = t2.editor
  AND t1.machine = t2.machine
  AND t1.operating_system = t2.operating_system
  AND t1.user_id = t2.user_id;