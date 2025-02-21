/* N1QL query to retrieve activities for an anonymous user. */
WITH
  activity_data AS (
    SELECT
      activity.kind,
      META(activity).id AS activity_id,
      activity.username,
      activity.timestamp,
      activity.target
    FROM
      `bucket_name` AS activity
    WHERE
      activity.type = 'activity'
      AND activity.src = "web"
      AND activity.kind != "comment"
      AND activity.kind != 'follow'
    ORDER BY
      activity.timestamp DESC
    OFFSET
      $offset
    LIMIT
      $limit
  )
SELECT
  {
    "type": activity.kind,
    "id": activity.activity_id,
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
    "follow": false,
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
          "class": f.ml.pe.predicted_class,
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
  activity_data AS activity
  LEFT JOIN `bucket_name` AS f ON KEYS activity.target;
