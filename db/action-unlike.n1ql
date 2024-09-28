/* N1QL query to remove a liked file from a user's likes. */

UPDATE `bucket_name`
USE KEYS $user
SET likes = ARRAY like_item FOR like_item IN likes
             WHEN like_item.sha256 != $sha256 END
WHERE ANY like_item IN likes SATISFIES like_item.sha256 = $sha256 END;
