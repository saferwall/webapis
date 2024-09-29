/* N1QL query to count activities for a logged-in user. */

WITH user_following AS (
(SELECT RAW
        s.`following`
    FROM
        `bucket_name` AS s
    USE KEYS $user
)[0]),
activities AS (
    SELECT COUNT(*) AS activity_count
    FROM
        user_following AS d
    INNER JOIN `bucket_name` AS activity ON activity.username = d.username
    WHERE
        activity.`type` = 'activity'
        AND activity.src = 'web'
        AND activity.kind != 'comment'
)

SELECT RAW activity_count
FROM activities;
