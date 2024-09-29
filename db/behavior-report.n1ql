
/* N1QL query to retrieve a behavior report for a file. */

SELECT OBJECT_CONCAT( (
        SELECT d.*
        FROM `bucket_name` d
        WHERE META(d).id = $behavior_id)[0],
    (
        SELECT a.api_trace
        FROM `bucket_name` a
        WHERE META(a).id = $behavior_id_apis)[0],
    (
        SELECT s.sys_events
        FROM `bucket_name` s
        WHERE META(s).id = $behavior_id_events)[0]
).*;
