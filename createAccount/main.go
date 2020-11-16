package main

// The purpose of this lambda is to provision the resources for a user whenever they create an account. This means:
// A) A user pool and entities needs for that user pool (domain, default dashboard app)
// B) An AppSync record
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

type CreateUserPoolRequest struct {
	PoolName    string `json:"name" schema:"name"`
	CallbackURL string `json: "callbackUrl", schema:"callbackUrl"`
}

type CreateUserPoolResponse struct {
	UserPoolID string
	ClientID   string
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var createPoolRequest CreateUserPoolRequest
	var response CreateUserPoolResponse

	err := json.Unmarshal([]byte(request.Body), &createPoolRequest)
	if err != nil {
		return Response{StatusCode: 400}, err
	}

	domainPrefix := fmt.Sprintf("tina-auth-%v", slug.Make(createPoolRequest.PoolName))

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
		AutoVerifiedAttributes: []*string{aws.String("email")},
		// user is allowed to use their email as their username
		UsernameAttributes: []*string{aws.String("email")},

		// TODO: Add in SES Configuration
	}

	createPoolResponse, err := cognitoService.CreateUserPool(newPoolParams)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	dashboardClientParams := &cognitoidentityprovider.CreateUserPoolClientInput{
		ClientName: aws.String("Dashboard"),

		AllowedOAuthFlows:               []*string{aws.String("code")}, // TODO: Confirm this is correct for PKCE
		AllowedOAuthFlowsUserPoolClient: aws.Bool(true),
		AllowedOAuthScopes:              []*string{aws.String("email"), aws.String("profile"), aws.String("openid")},
		CallbackURLs:                    []*string{aws.String(createPoolRequest.CallbackURL)},
		SupportedIdentityProviders:      []*string{aws.String("COGNITO")},
		UserPoolId:                      createPoolResponse.UserPool.Id,
	}
	clientResponse, err := cognitoService.CreateUserPoolClient(dashboardClientParams)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	// create a domain to user for this UserPool, possibly will want to switch this to a custom domain
	poolDomainParams := &cognitoidentityprovider.CreateUserPoolDomainInput{
		Domain:     aws.String(domainPrefix),
		UserPoolId: createPoolResponse.UserPool.Id,
	}

	_, err = cognitoService.CreateUserPoolDomain(poolDomainParams)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	response.UserPoolID = *createPoolResponse.UserPool.Id
	response.ClientID = *clientResponse.UserPoolClient.ClientId

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	return Response{
		StatusCode: 201,
		Headers:    map[string]string{"Access-Control-Allow-Origin": "http://localhost:3002"},
		Body:       string(responseJSON),
	}, nil

}

func main() {
	lambda.Start(Handler)
}
