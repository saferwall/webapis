/* N1QL query to retrieve users' following for a logged-in user. */

SELECT RAW {
    "id": META(p).id,
    "username": p.`username`,
    "member_since": p.`member_since`,
    "follow": ARRAY_BINARY_SEARCH(ARRAY_SORT((
            SELECT RAW ARRAY LOWER(u.username) FOR u IN n.`following` END
            FROM `bucket_name` n USE KEYS $loggedInUser ) [0]),
    META(p).id) >= 0 }
FROM `bucket_name` p USE KEYS [(
    SELECT RAW ARRAY LOWER(u.username) FOR u IN s.`following` END
    FROM `bucket_name` s USE KEYS $user )[0]]
OFFSET $offset
LIMIT
  $limit
