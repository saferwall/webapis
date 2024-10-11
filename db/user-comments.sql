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
              `bucket_name` USE KEYS $loggedInUser
          ) [0]
        ),
        c.sha256
      ) >= 0,
      "date": c.timestamp,
      "file": {
        "hash": f.sha256,
        "tags": f.tags,
        "filename": f.submissions [0].filename,
        "class": f.ml.pe.predicted_class,
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
  `bucket_name` c
  LEFT JOIN `bucket_name` f ON KEYS c.sha256
WHERE
  c.`type` = 'comment' AND c.`username` = $user
ORDER BY
  c.timestamp DESC
OFFSET $offset
LIMIT
  $limit
