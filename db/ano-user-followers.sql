/* N1QL query to retrieve user followers for an anonymous user. */
SELECT
  RAW {
    "id": META(p).id,
    "username": p.`username`,
    "member_since": p.`member_since`,
    "follow": FALSE
  }
FROM
  `bucket_name` p
USE KEYS [
  (
    SELECT
      RAW ARRAY LOWER(u.username) FOR u IN s.`followers` END
    FROM
      `bucket_name` s
    USE KEYS $user
  ) [0]
]
OFFSET
  $offset
LIMIT
  $limit
