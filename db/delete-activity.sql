/* N1QL query to delete user's activity. */

DELETE
 FROM `bucket_name`
WHERE
 `type` = "activity"
AND kind = $kind
AND username = $username
AND target = $target
