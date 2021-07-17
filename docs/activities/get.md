# Get activities

Get the details of the currently Authenticated User along with basic
subscription information.

**URL** : `/activities/`

**Method** : `GET`

**Auth required** : No

**Permissions required** : None

## Success Response

**Code** : `200 OK`

**Content examples**

For an activity with `submit` type:

```json
{
  "type": "submit",
  "author": {
    "username": "MrRobot",
    "member_since": 1626494492
  },
  "follow": true,
  "file": {
    "hash": "01583d41cc2559280eb9a5495d2fb2678f60cb56f88cf32143f5392d499f0b3e",
    "tags": {
      "packer": ["vmprotect"],
      "pe": ["dll"]
    },
    "classification": "Label.MALICIOUS",
    "name": "malware.exe",
    "score": {
      "value": 2,
      "total": 12
    }
  },
  "timestamp": 162649449
}
```

For an activity with `like` type:

```json
{
  "type": "like",
  "author": {
    "username": "MrRobot",
    "member_since": 1626494492
  },
  "follow": true,
  "liked": false,
  "file": {
    "hash": "01583d41cc2559280eb9a5495d2fb2678f60cb56f88cf32143f5392d499f0b3e",
    "tags": {
      "packer": ["vmprotect"],
      "pe": ["dll"]
    },
    "classification": "Label.MALICIOUS",
    "name": "malware.exe",
    "score": {
      "value": 2,
      "total": 12
    }
  },
  "timestamp": 162649449
}
```

For an activity with `comment` type:

```json
{
  "type": "comment",
  "author": {
    "username": "MrRobot",
    "member_since": 1626494492
  },
  "follow": true,
  "comment": "#wannacry #ransomware",
  "timestamp": 162649449
}
```

For an activity with `follow` type:

```json
{
  "type": "follow",
  "author": {
    "username": "MrRobot",
    "member_since": 1626494492
  },
  "target": {
    "username": "40hex",
    "member_since": 4786494493
  },
  "follow": true,
  "timestamp": 162649449
}
```
