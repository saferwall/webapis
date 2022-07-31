## Indexes

- users acitivities:

```
CREATE INDEX adv_username_type ON `sfw`(`username`) WHERE `type` = 'activity'
CREATE INDEX adv_meta_self_id_type ON `sfw`(meta(self).`id`) WHERE `type` = 'file'
CREATE INDEX act_timestamp_desc ON `sfw`(`type`,`timestamp` DESC)
```

## Statistics

- total number of docs: 3964485
- total number of `file` docs: 3643550; 320935
- total number of `activity` docs: 320741
- total number of `comment` docs: 10
- total number of `user` docs: 163

## Queries

- magic data
```sql
SELECT DISTINCT magic
FROM sfw d
WHERE d.`type` = "file"
```

- update
```sql
MERGE INTO sfw AS d
USING  (SELECT META(act).id, c.sha256
        FROM sfw AS act
        JOIN sfw AS c
        ON act.`type`="activity" AND act.`kind`="comment" AND act.target = META(c).id) AS s
ON KEY s.id
WHEN MATCHED THEN
UPDATE set d.target = s.sha256;
```
