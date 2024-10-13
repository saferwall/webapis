/* N1QL query to retrieve UI metadata. */
SELECT
  pe.meta as pe,
  default_behavior_report.id as default_behavior_id,
  ARRAY_CONCAT(
    ARRAY_INTERSECT(
      OBJECT_NAMES(d),
      [
        "pe",
        "elf",
        "strings",
        "behavior_scans"
      ]
    ),
    ["summary", "comments", "antivirus"]
  ) AS tabs
FROM
  `bucket_name` AS d
USE KEYS $sha256
