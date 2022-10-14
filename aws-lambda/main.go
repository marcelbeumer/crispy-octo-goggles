package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("Processing request data for request %s.\n", req.RequestContext.RequestID)
	fmt.Printf("Body size = %d.\n", len(req.Body))

	fmt.Println("Headers:")
	for key, value := range req.Headers {
		fmt.Printf("    %s: %s\n", key, value)
	}

	return events.APIGatewayProxyResponse{Body: "ok", StatusCode: 200}, nil
}

func main() {
	lambda.Start(handleRequest)
}
