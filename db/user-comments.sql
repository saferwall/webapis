/* N1QL query to retrieve user's comments for a logged-in user. */
SELECT
  {
    "id": META(c).id,
    "comment": c.body,
    "liked": ARRAY_BINARY_SEARCH(
      ARRAY_SORT(
        (
          SELECT
            RAW likes
          FROM
            `bucket_name`
          USE KEYS $loggedInUser
        ) [0]
      ),
      c.sha256
    ) >= 0,
    "date": c.timestamp,
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
  }.*
FROM
  `bucket_name` c
  LEFT JOIN `bucket_name` f ON KEYS c.sha256
WHERE
  c.`type` = 'comment'
  AND c.`username` = $user
ORDER BY
  c.timestamp DESC
OFFSET
  $offset
LIMIT
  $limit
