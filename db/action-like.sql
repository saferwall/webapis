/* N1QL query to prepend a liked file from a user's likes. */

UPDATE `bucket_name`
USE KEYS $user
SET likes = ARRAY_PREPEND({"sha256": $sha256, "ts": $ts}, likes)
WHERE NOT ANY item IN likes SATISFIES item.sha256 = $sha256 END;
