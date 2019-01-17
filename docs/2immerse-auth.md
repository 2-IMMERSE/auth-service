+++
title = "Authentication in 2Immerse"
weight = 1
+++

## Why does 2Immerse need authentication?

Aside from the security benefits of having secured AP endpoints the 2Immerse
uses authentication to provide user identity.

This use identity is important to the platform as it allows several services to
make intelligent decisions about how to respond to user input and interactions.

## How does 2Immerse use authentication

![Platform Layout](/docs/images/platform-layout.png)

In the first instance the 2Immerse platform uses OAuth2 access tokens to secure
API endpoints. All requests pass through the API gateway which evaluates authentication
credentials for validity. The gateway is responsible for redirecting clients to
the authentication endpoint and users can authenticate with the platform using
the password flow.

Once a user is authenticated and they have unrestricted access to services, their
token is forwarded with requests to backend services allowing the service to
identify the user and lookup additional configuration specific to the currently
logged in user.

Very simply this looks like:

![Traffic Flow](/docs/images/traffic-flow.gif)
