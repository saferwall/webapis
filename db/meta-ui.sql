/* N1QL query to retrieve UI metadata. */
SELECT
  pe.meta as pe,
  ARRAY_CONCAT(
    ARRAY_INTERSECT(
      OBJECT_NAMES(d),
      [
        "pe",
        "elf",
        "strings",
        "multiav",
        "behavior_scans"
      ]
    ),
    ["summary", "comments"]
  ) AS tabs
FROM
  `bucket_name` AS d
USE KEYS $sha256
