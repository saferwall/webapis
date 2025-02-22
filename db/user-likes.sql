/* N1QL query to retrieve likes for an logged-in user. */
SELECT
  {
    "id": UUID(),
    "date": l.ts,
    "liked": ARRAY_BINARY_SEARCH(
      ARRAY_SORT(
        (
          SELECT
            RAW ARRAY like_item.sha256 FOR like_item IN u.`likes` END AS sha256_array
          FROM
            `bucket_name` u
          USE KEYS $loggedInUser
        ) [0]
      ),
      f.sha256
    ) >= 0,
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
  (
    SELECT
      userLikes.*
    FROM
      `bucket_name` s
    USE KEYS $user
    UNNEST
      s.likes AS userLikes
  ) AS l
  LEFT JOIN `bucket_name` f ON f.sha256 = l.sha256
WHERE
  f.`type` = "file"
OFFSET
  $offset
LIMIT
  $limit
