/* N1QL query to get all existing document of a given type. */
SELECT
  `bucket_name`.*
FROM
  `bucket_name`
WHERE
  type = $docType
OFFSET
  $offset
LIMIT
  $limit
