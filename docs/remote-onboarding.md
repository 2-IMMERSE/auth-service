+++
title = "Remote Onboarding"
weight = 4
+++
In the case where the communal device and companion device are not on the same network, the device connection API can be used to link communal devices to a user account. This is useful for new user onboarding as well as temporary log on to guest devices.

This is especially useful for communal devices that have no direct input device, meaning entering authentication credentials is prohibitive.


### Connection code (*WIP*)

This remote system works by the communal device first requesting a connection code from the auth service by sending it's own name to the devices endpoint.

```
POST /devices
{ "name": "Home-TV" }
```

The API will return a time limited code to the device:

```
{ "code": "12345" }
```

This code will expire in 30 minutes.

At this point the communal device should either poll the details endpoint or connect to the websocket service and wait for a connection notification.

### User authentication (*WIP*)

Next the user will follow the standard on-boarding / authentication flow on their companion device and when asked for the device code the should enter the code displayed on the communal device.

The auth service will then send the access token to the communal device so it can now interact with services as the logged in user.

The auth service will also automatically expire any connection codes that have successfully being linked to accounts.
