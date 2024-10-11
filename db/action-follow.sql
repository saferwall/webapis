/* N1QL query to prepend a username to the user's following. */
UPDATE `bucket_name`
USE KEYS LOWER($user)
SET
  `DYNAMIC_FIELD` = ARRAY_PREPEND(
    {"username": $username, "ts": $ts},
    `DYNAMIC_FIELD`
  )
WHERE
  NOT ANY item IN `DYNAMIC_FIELD` SATISFIES item.username = $username END;
