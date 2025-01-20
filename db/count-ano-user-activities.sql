SELECT
  RAW COUNT(*)
FROM
  `bucket_name` activity
WHERE
  activity.type = "activity"
  AND activity.src = "web"
  AND activity.kind != "comment"
