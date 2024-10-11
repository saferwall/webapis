/* N1QL query to get file comments. */

SELECT
{
  "comment": c.body,
  "date": c.timestamp,
  "id": META(c).id,
  "author": {
    "follow":  ARRAY_BINARY_SEARCH(ARRAY_SORT((
              SELECT
              RAW `following`
              FROM
              `bucket_name`
              USE KEYS $loggedInUser
              ) [0]), c.username) >= 0,
    "username":  c.username,
    "member_since": (
      SELECT
        RAW u.member_since
      FROM
        `bucket_name` u
      USE KEYS
        LOWER(c.username)
    ) [0]
  }
}.*
FROM
  `bucket_name` c
WHERE
  c.`sha256` = $sha256 AND c.`type` = 'comment'
ORDER BY
  c.timestamp DESC
OFFSET $offset
LIMIT
  $limit
