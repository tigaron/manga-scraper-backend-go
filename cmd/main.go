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
)

var (
	seriesTable   = os.Getenv("SERIES_TABLE")
	chaptersTable = os.Getenv("CHAPTERS_TABLE")
	ddbClient     dynamodbiface.DynamoDBAPI
)

func main() {
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	ddbClient = dynamodb.New(awsSession)
	lambda.Start(handler)
}

func handler(sqsEvent events.SQSEvent) error {
	requestType := sqsEvent.Records[0].MessageAttributes["RequestType"].StringValue
	provider := sqsEvent.Records[0].MessageAttributes["Provider"].StringValue
	sourceUrl := sqsEvent.Records[0].MessageAttributes["SourceUrl"].StringValue
	switch *requestType {
	case "series-list":
		return handlers.SeriesListRequest(provider, sourceUrl, seriesTable, ddbClient)
	case "series-data":
		return handlers.SeriesDataRequest(provider, sourceUrl, seriesTable, ddbClient)
	case "chapters-list":
		return handlers.ChapterListRequest(provider, sourceUrl, chaptersTable, ddbClient)
	case "chapters-data":
		return handlers.ChapterDataRequest(provider, sourceUrl, chaptersTable, ddbClient)
	default:
		log.Printf("Couldn't handle request type of '%v'", *requestType)
		return nil
	}
}

/*
MessageAttributes: {
	"RequestType": {
		DataType: "String",
		StringValue: "series-list"
	},
	"Provider": {
		DataType: "String",
		StringValue: "asura"
	},
	"SourceUrl": {
		DataType: "String",
		StringValue: "https://asura.gg/manga/list-mode/"
	},
}
*/
