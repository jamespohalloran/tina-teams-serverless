package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/gosimple/slug"
)

type Response events.APIGatewayProxyResponse

type CreatePoolRequest struct {
	PoolName string `json:"name"`
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var createPoolRequest CreatePoolRequest

	err := json.Unmarshal([]byte(request.Body), &createPoolRequest)
	if err != nil {
		return Response{StatusCode: 400}, err
	}

	fmt.Println("THIS IS THE CREATE POOL REQUEST")
	fmt.Println(createPoolRequest)

	awsSession, err := session.NewSession()

	if err != nil {
		return Response{StatusCode: 500}, err
	}

	cognitoService := cognitoidentityprovider.New(awsSession)

	newPoolParams := &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: &createPoolRequest.PoolName,

		AdminCreateUserConfig: &cognitoidentityprovider.AdminCreateUserConfigType{
			AllowAdminCreateUserOnly: aws.Bool(false), // users can create their own accounts
		},
		// user is allowed to use their email as their username
		UsernameAttributes: []*string{aws.String("email")},

		// TODO: Add in SES Configuration
	}

	createPoolResponse, err := cognitoService.CreateUserPool(newPoolParams)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	// create a domain to user for this UserPool, possibly will want to switch this to a custom domain
	poolDomainParams := &cognitoidentityprovider.CreateUserPoolDomainInput{
		Domain:     aws.String(fmt.Sprintf("tina-auth-%v", slug.Make(createPoolRequest.PoolName))),
		UserPoolId: createPoolResponse.UserPool.Id,
	}

	_, err = cognitoService.CreateUserPoolDomain(poolDomainParams)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	dashboardClientParams := &cognitoidentityprovider.CreateUserPoolClientInput{
		ClientName: aws.String("Dashboard"),

		AllowedOAuthFlows:               []*string{aws.String("code")}, // TODO: Confirm this is correct for PKCE
		AllowedOAuthFlowsUserPoolClient: aws.Bool(true),
		AllowedOAuthScopes:              []*string{aws.String("email"), aws.String("profile"), aws.String("openid")},
		CallbackURLs:                    []*string{aws.String("http://localhost:3000")},
		SupportedIdentityProviders:      []*string{aws.String("COGNITO")},
		UserPoolId:                      createPoolResponse.UserPool.Id,
	}

	createClientResponse, err := cognitoService.CreateUserPoolClient(dashboardClientParams)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	fmt.Println(createClientResponse)
	return Response{StatusCode: 201}, nil

}

func main() {
	lambda.Start(Handler)
}
