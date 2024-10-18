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
      "class": f.ml.pe.predicted_class,
      "default_behavior_report": f.default_behavior_report,
      "multiav": {
        "value": ARRAY_COUNT(
          ARRAY_FLATTEN(
            ARRAY i.infected FOR i IN OBJECT_VALUES(f.multiav.last_scan) WHEN i.infected = TRUE END,
            1
          )
        ),
        "count": OBJECT_LENGTH(f.multiav.last_scan)
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
