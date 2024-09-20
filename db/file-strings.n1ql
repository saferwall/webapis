SELECT RAW
    ARRAY_FLATTEN (
    [
        ARRAY {"value": s, "encoding": "ascii"} FOR s IN f.strings.ascii WHEN LOWER(s) LIKE LOWER($term) END,
        ARRAY {"value": s, "encoding": "wide"} FOR s IN f.strings.wide WHEN LOWER(s) LIKE LOWER($term) END,
        ARRAY {"value": s,  "encoding": "asm"} FOR s IN f.strings.asm WHEN LOWER(s) LIKE LOWER($term) END
    ], 1 )[$offset:$limit]

FROM `bucket_name` f
USE KEYS $sha256
