# Get activities

Create a new user.

**URL** : `/v1/users`

**Method** : `POST`

**Auth required** : No

**Permissions required** : None

## Success Response

**Code** : `201 Created`

**Content examples**

Request body:

```json
{
    "username": "mrrobot", "password": "allsafe", "email": "mrrobot@ecorp.com"
}
```

Response when invalid data is sent:

```json
{
    "status": 400,
    "message": "There is some problem with the data you submitted.",
    "details": [
        "Key: 'CreateUserRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag"
    ]
}
```

Response when valid data is sent:

```json
{
    "email": "mrrobot@ecorp.com",
    "username": "mrrobot",
    "name": "",
    "location": "",
    "url": "",
    "bio": "",
    "confirmed": false,
    "member_since": 1626687896,
    "last_seen": 1626687896,
    "admin": false,
    "has_avatar": false,
    "following": null,
    "following_count": 0,
    "followers": null,
    "followers_count": 0,
    "likes": null,
    "likes_count": 0,
    "submissions_count": 0,
    "comments_count": 0
}
```
