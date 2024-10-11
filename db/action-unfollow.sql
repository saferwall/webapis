/* N1QL query to remove a user from a user's following. */

UPDATE `bucket_name`
USE KEYS LOWER($user)
SET `DYNAMIC_FIELD` = ARRAY item FOR item IN `DYNAMIC_FIELD`
             WHEN item.username != $username END
WHERE ANY item IN `DYNAMIC_FIELD` SATISFIES item.username = $username END;
