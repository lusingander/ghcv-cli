# ghcv-cli

## Setup

First, Authorization must be granted according to the [Device flow](https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#device-flow). You will need to enter the code that will be displayed in console when you start the application.

> The application requires only minimal scope (access to public information).

Or, you can set the [personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token) as the environment variable.

```sh
export GHCV_GITHUB_ACCESS_TOKEN=<token>
```

> In this case as well, you don't need to specify anything in the scope (only public information will be accessed).

