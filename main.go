package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	idp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

//Input struct
type refreshToken struct {
	RefreshToken string `json:"refresh_token"`
}

//Output struct
type authToken struct {
	IDToken      string `json:"idToken"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
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

//HandleRequest the APIGateway proxy request and return either an error or exchanged tokens
func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var myRefreshToken refreshToken
	log.Println("BODY: ", event.Body)
	err := json.Unmarshal([]byte(event.Body), &myRefreshToken)
	headers := make(map[string]string)
	headers["Access-Control-Allow-Origin"] = "*"
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 400, Headers: headers}, nil
	}
	output, err := myRefreshToken.refresh()
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 400, Headers: headers}, nil
	}
	var returnToken authToken
	//Dereference the token variables
	idToken := *output.AuthenticationResult.IdToken
	accessToken := *output.AuthenticationResult.AccessToken
	refreshToken := *output.AuthenticationResult.RefreshToken
	returnToken.IDToken = idToken
	returnToken.AccessToken = accessToken
	returnToken.RefreshToken = refreshToken
	data, err := json.Marshal(returnToken)
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 400, Headers: headers}, nil
	}
	response := events.APIGatewayProxyResponse{
		Body:       string(data),
		StatusCode: 200,
		Headers:    headers,
	}
	fmt.Printf("%+v\n", response)
	return response, nil
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
