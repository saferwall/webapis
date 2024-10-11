/* N1QL query to retrieve file summary of a scan. */
WITH
  user_likes AS (
    SELECT
      RAW ARRAY like_item.sha256 FOR like_item IN u.`likes` END AS sha256_array
    FROM
      `bucket_name` u
    USE KEYS $loggedInUser
  )
SELECT
  {
    "status": f.status,
    "default_behavior_report": f.default_behavior_report,
    "comments_count": f.comments_count,
    "liked": CASE
      WHEN ARRAY_LENGTH(user_likes) = 0 THEN false
      ELSE ARRAY_BINARY_SEARCH(ARRAY_SORT((user_likes) [0]), f.sha256) >= 0
    END
  }.*
FROM
  `bucket_name` f
USE KEYS $sha256
