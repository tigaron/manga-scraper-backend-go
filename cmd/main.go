package main

import (
	"log"
	"manga-scraper-be-go/pkg/handlers"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

var (
	seriesTable   = os.Getenv("SERIES_TABLE")
	chaptersTable = os.Getenv("CHAPTERS_TABLE")
	queueUrl      = os.Getenv("QUEUE_URL")
	ddbClient     dynamodbiface.DynamoDBAPI
	sqsClient     sqsiface.SQSAPI
)

func main() {
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	ddbClient = dynamodb.New(awsSession)
	sqsClient = sqs.New(awsSession)
	lambda.Start(handler)
}

func handler(sqsEvent events.SQSEvent) error {
	requestType := sqsEvent.Records[0].MessageAttributes["RequestType"].StringValue
	provider := sqsEvent.Records[0].MessageAttributes["Provider"].StringValue
	sourceUrl := sqsEvent.Records[0].MessageAttributes["SourceUrl"].StringValue
	switch *requestType {
	case "series-list":
		return handlers.SeriesListRequest(provider, sourceUrl, seriesTable, ddbClient, queueUrl, sqsClient)
	case "series-data":
		return handlers.SeriesDataRequest(provider, sourceUrl, seriesTable, ddbClient)
	case "chapter-list":
		return handlers.ChapterListRequest(provider, sourceUrl, chaptersTable, ddbClient, queueUrl, sqsClient)
	case "chapter-data":
		return handlers.ChapterDataRequest(provider, sourceUrl, chaptersTable, ddbClient)
	case "chapter-update": // TODO
		return nil
	default:
		log.Fatalf("Couldn't handle request type of '%v'", *requestType)
		return nil
	}
}
