# Cognito Refresh

## Lambda function to refresh Cognito Auth tokens

This function handles exchanging the Auth flow token for valid IDP tokens from AWS Cognito

The request structure to send to the function looks like this:

```json
{
    "refresh_token": "some token here"
}
```

The response from the services should be either this:

```json
{
    "idToken": "the generated token",
    "accessToken": "the generated access token"
}
```

Or if the request is invalid will return a 400 status code.  This means the user needs to sign in again

## Building
Install Godep for dependency management

```shell
dep ensure
GOOS=linux go build -o main
zip deployment.zip main
```

TODO:
 * add more api gateway types

## Deployment
Upload to Lambda on AWS
Make the Lambda source an API Gateway resource
Enable CORS on the API Gateway path for this function

TODO:
 * Setup the ability for local testing