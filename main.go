package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	idp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

type refreshToken struct {
	RefreshToken string `json:"refresh_token"`
}

type authToken struct {
	IDToken     string `json:"idToken"`
	AccessToken string `json:"accessToken"`
}

//Contact Cognito and exchange refresh_token for idtoken and accesstoken
func (refToken *refreshToken) refresh() (*idp.InitiateAuthOutput, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	svc := idp.New(sess)
	return svc.InitiateAuth(
		&idp.InitiateAuthInput{
			AuthFlow:       aws.String("REFRESH_TOKEN"),
			AuthParameters: map[string]*string{"REFRESH_TOKEN": aws.String(refToken.RefreshToken)},
			ClientId:       aws.String(os.Getenv("CLIENTID")),
		},
	)
}

//process the APIGateway proxy request and return either an error or exchanged tokens
func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var myRefreshToken refreshToken
	err := json.Unmarshal([]byte(event.Body), &myRefreshToken)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 400}, nil
	}
	output, err := myRefreshToken.refresh()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 400}, nil
	}
	var returnToken authToken
	//Dereference the token variables
	idToken := *output.AuthenticationResult.IdToken
	accessToken := *output.AuthenticationResult.AccessToken
	returnToken.IDToken = idToken
	returnToken.AccessToken = accessToken
	data, err := json.Marshal(returnToken)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 400}, nil
	}
	return events.APIGatewayProxyResponse{Body: string(data), StatusCode: 200}, nil
}

//Entrypoint lambda to run code
func main() {
	switch os.Getenv("PLATFORM") {
	case "lambda":
		lambda.Start(HandleRequest)
	default:
		log.Println("no platform defined")
	}
}
