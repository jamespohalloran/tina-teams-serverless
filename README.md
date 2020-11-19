# tina-teams-serverless

## Requirements
* Serverless `npm install -g serverless`
* aws-cli 

## Deploying

Assuming you have serverless and the aws-cli configured locally, you should be able to deploy functions with 
```
make deploy
```

This will setup a Lambda/API Gateway so that the functions can be be called. You can get the URL for your functions by checking in API Gateway

## Info

* Function configuration lives in `serverless.yml`. Also in this is the permissions that the serverless app will require (currently just * on all cognito resources)
* CORS is setup to allow API requests from `http://localhost:3002`

## createAccount
There is currently a single function that creates a userpool/domain/default application. I've tried to set this up to have the defaults that we want (i.e., email 
address as username, linked to a dashboard app client, email verification, PKCE compatible auth, etc)

```
POST <API GATEWAY URL>/createAccount
{
  "name": string,
  "callbackUrl": string
}
```

If successful it returns
```
{
  UserPoolID: string,
  ClientID: string
}
```

it also creates a domain that looks like `https://tina-auth-<slugified-realm-name>.auth.us-east-1.amazoncognito.com`. 
This means that the hosted UI for authentication _should_ be available at 
`https://tina-auth-<slugified-realm-name>.auth.us-east-1.amazoncognito.com/login?client_id=<CliendID>&response_type=code&scope=email+openid+profile&redirect_uri=http://localhost:3002/callback`
