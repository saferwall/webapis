WITH
  user_following AS (
    (
      SELECT
        RAW s.`following`
      FROM
        `bucket_name` AS s
      USE KEYS $user
    ) [0]
  ),
  activities AS (
    SELECT
      activity.*,
      d.username,
      d.member_since
    FROM
      user_following AS d
      INNER JOIN `bucket_name` AS activity ON activity.username = d.username
    WHERE
      activity.`type` = 'activity'
      AND activity.src = 'web'
      AND activity.kind != 'comment'
      AND activity.kind != 'follow'
  )
SELECT
  {
    "type": activity.kind,
    "id": META(activity).id,
    "author": {
      "username": activity.username,
      "member_since": (
        SELECT
          RAW u.member_since
        FROM
          `bucket_name` AS u
        USE KEYS LOWER(activity.username)
      ) [0]
    },
    "follow": TRUE,
    "comment": f.body,
    "date": activity.timestamp
  }.*,
  (
    CASE
      WHEN activity.kind = "follow" THEN {"target": activity.target}
      ELSE {
        "file": {
          "hash": f.sha256,
          "tags": f.tags,
          "filename": f.submissions[0].filename,
          "class": f.classification,
          "default_behavior_id": f.default_behavior_report.id,
          "multiav": {
            "value": f.multiav.last_scan.stats.positives,
            "count": f.multiav.last_scan.stats.engines_count
          }
        }
      }
    END
  ).*
FROM
  activities AS activity
  INNER JOIN `bucket_name` AS f ON activity.target = META(f).id
WHERE
  f.`type` = 'file'
ORDER BY
  activity.timestamp DESC
OFFSET
  $offset
LIMIT
  $limit;
