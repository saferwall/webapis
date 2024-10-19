/* N1QL query to prepend a file hash to the user's submissions. */
UPDATE `bucket_name`
USE KEYS $user
SET
  submissions = ARRAY_PREPEND({"sha256": $sha256, "ts": $ts}, submissions)
WHERE
  NOT ANY item IN submissions SATISFIES item.sha256 = $sha256 END;
