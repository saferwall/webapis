/* N1QL query to retrieve file summary of a scan. */
SELECT
  {
    "properties": {
      "Size": f.size,
      "CRC32": f.crc32,
      "MD5": f.md5,
      "SHA1": f.sha1,
      "SHA256": f.sha256,
      "SHA512": f.sha512,
      "SSDeep": f.ssdeep,
      "TLSH": f.tlsh,
      "Packer": f.packer,
      "Magic": f.magic,
      "Tags": f.tags,
      "TrID": f.trid
    },
    "sha256": f.sha256,
    "first_seen": f.first_seen,
    "last_scanned": f.last_scanned,
    "exif": f.exif,
    "submissions": f.submissions,
    "class": f.ml.pe.predicted_class,
    "file_format": f.file_format,
    "file_extension": f.file_extension,
    "signature": f.pe.signature,
    "default_behavior_report": f.default_behavior_report,
    "multiav": {
      "value": ARRAY_COUNT(
        ARRAY_FLATTEN(
          ARRAY i.infected FOR i IN OBJECT_VALUES(f.multiav.last_scan) WHEN i.infected = TRUE END,
          1
        )
      ),
      "count": OBJECT_LENGTH(f.multiav.last_scan)
    }
  }.*
FROM
  `bucket_name` f
USE KEYS $sha256
