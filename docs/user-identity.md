+++
title = "User Identity"
weight = 3
+++
When getting additional information about a user the identity endpoint will return
a JSON data struct containing several predetermined keys:

```
{
  "ROLES": [],
  "PREFERENCES": {},
  "profile": {
    ...
  }
}
```

A users own identity can be retrieved by a simple request:

```
GET /identify/{access_token}
```

## Roles and permissions

### The ROLES object

The ```ROLES``` object contains a list of "roles" assigned to a user. Applications
can use these roles to determine what actions a user can take and what element should be made visible in the UI.

### Permissions (*WIP*)

Each user can have multiple permissions linked to their account. Permissions need to be created and assigned to a user before being requested.

Permissions are designed for fine grained access control and the auth service provides an endpoint to allow external services to determine if a user has the required permissions. E.G.

**Can the user create new users**
```
GET /permissions/{permission_name}
```

## Preferences (*WIP*)

Preferences are used for storing global, top level user settings.

## Profile

Each user has a single profile which contains multiple "buckets" of data. The profile is split into several buckets to allow preferences to be targeted to specific devices and experiences.

Data stored under the specified buckets are simple key -> value pairs and there is no restriction on data stored.

### Communal bucket

```
{
  "profile": {
    "communal": {
      "prefs": {},
    }
  }
}
```

The communal bucket contains at least a single key of ```prefs```. This key contains user settings for all communal device experiences.

### Companion bucket

```
{
  "profile": {
    "companion": {
      "prefs": {},
    }
  }
}
```

The companion bucket contains at least a single key of ```prefs```. This key contains user settings for all companion device experiences.

### Trial buckets

Each user can be signed up for several trial experiences. For example, in the case of a developer, they may be running several instances of the trial which requires slightly different settings for testing. Alongside the ```prefs``` key under both communal and companion buckets, these trial keys can be used to store preferences specific to each trial and device type.

E.G. a user signed up to trial MotoGP:

```
{
  "profile": {
    "companion": {
      "prefs": {
        "show_advanced_stats": true
      },
      "motogb_1.0.0": {
        "fav_rider": "karel_abraham"
      }
    }
  }
}
```

## Profile API (*WIP*)

There is an API that clients can use to read / edit these preferences.

There is currently no way to make preferences read-only.
