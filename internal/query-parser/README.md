# Query Parser

This package implements a parsing and transpiling of _search terms or modifiers_ into **Couchbase SQL++** (formerly *N1QL*).

## Specification

- Casing does not matter.
- Search modifiers have to respect the case sensitivity.
- Search values does not need to respect the case sensitivity.
- Dates are stored as `int64` (unix timestamps).
- Dates are in *ISO* format: `2023-09-12T14:30:00`.

You can apply conditionals on the different modifiers.
  - `AND` : usual boolean AND operation, both modifiers must be satisfied in the query.
  - `OR` : usual boolean OR operation, only a single modifier needs to be satisfied.
  - Use mathematical operators to indicate `>`(bigger), `<=`(smaller or equal).
  - Use `!=` to indicate non-equality.

## Search Modifiers

| Modifier  | Description                                                                                                                                                                                                                                          |
| --------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| size      | Represents the file size. The size can be specified in bytes (default), kilobytes or megabytes.                                                                                                                                                      |
| name      | Represents the file name. Saferwall keeps track of all file names submitted for a unique file.                                                                                                                                                       |
| type      | Represents the file format. This is the full list of available file type literals: `pe`, `elf` `macho`, `pdf`, `doc`                                                                                                                                 |
| extension | Represents the file extension. Examples of file extensions are `exe`, `ps1`, `sys`, ..                                                                                                                                                               |
| fs        | Stands for First Seen, it allows you to search files according to the first submission date.                                                                                                                                                         |
| ls        | Stands for Last Seen, it allows you to search files according to the last scan date.                                                                                                                                                                 |
| positives | Represents the count of antivirus vendors that flags the file as malicious. It allows you to specify larger than or smaller than values (max = 14).                                                                                                  |
| engines   | Allows you to search for a detection in any anti-virus vendor.                                                                                                                                                                                       |
| <av>      | This modifier allows you to target a specific anti-virus vendor. The full list of allowed vendors is: `avast`, `avira`, `bitdefender`, `clamav`, `comodo`, `drweb`, `eset`, `kaspersky`, `mcafee`, `sophos`, `symantec`, `trendmicro`, `windefender` |
| imphash   | Returns all PE files that are similar to the Import Hash given.                                                                                                                                                                                      |
| ssdeep    | Returns all files that are similar to the ssdeep hash provided.                                                                                                                                                                                      |
| tlsh      | Returns all files that are similar to the TLSH hash provided.                                                                                                                                                                                        |
| crc32     | Returns all files that are similar to the CRC32 hash provided.                                                                                                                                                                                       |
| trid      | Allows you to search a substring inside [TRiD](https://mark0.net/soft-trid-e.html) File Identifier                                                                                                                                                   |
| packer    | Allows you to search a substring inside [DiE](https://horsicq.github.io/) file type identification and packer detection tool.                                                                                                                        |
| magic     | Allows you to search a substring inside the linux utility `file`.                                                                                                                                                                                    |
| tag | Returns files that are tagged with a specific tag. The full list of supported tags: `packed`, `signed`, `benign`, ... |

## Examples:

Search all files of type PE and tagged as UPX:
```c
type=pe and tag=upx // AND is optional, by default, we always AND 2 search sub-expressions.
type=pe or tag=upx // Usage of OR,
( type=pe or tag=upx ) and avast!=locky // Parenthesis.
size >= 1000000 // by default, it's bytes.
size < 1000KB // use KB to imply kilo-bytes.
size > 1MB
fs >= 2009 // First seen after year 2009.
fs <= 2020-12 // First seen before December 2020.
fs <= 2020-01-30 // Example of full date
fs < 2012-08-21T1
fs < 2012-08-21T16:59
fs < 2012-08-21T16:59:20 // UTC
fs < 2012-08-21T16:59:20Z // UTC explicit
fs < 2012-08-21T16:59:20+02:00 // 2 hours ahead of UTC
fs < 3d // `ls` has the same syntax as `fs`.
```
