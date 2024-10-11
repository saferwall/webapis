SELECT
  RAW COUNT(*)
FROM
  `bucket_name` s
WHERE
  s.type = "activity"
  AND activity.src = "web"
  AND activity.kind != "comment"
