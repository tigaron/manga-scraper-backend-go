package handlers

import (
	"log"
	"manga-scraper-be-go/pkg/scraper"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

func SeriesListRequest(message map[string]events.SQSMessageAttribute, tableName string, ddbClient dynamodbiface.DynamoDBAPI) error {
	data, err := scraper.ScrapeSeriesList(message, tableName)
	if err != nil {
		return err
	}

	for _, entry := range *data {
		item, err := dynamodbattribute.MarshalMap(entry)
		if err != nil {
			log.Printf("Failed to marshal entry of %v\n", entry)
			continue
		}

		cond := expression.AttributeNotExists(expression.Name("SeriesId"))
		expr, err := expression.NewBuilder().WithCondition(cond).Build()
		if err != nil {
			log.Printf("Failed to build expression for %v\n", entry)
			continue
		}

		input := &dynamodb.PutItemInput{
			TableName:                 aws.String(tableName),
			Item:                      item,
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			ConditionExpression:       expr.Condition(),
		}

		_, err = ddbClient.PutItem(input)
		if err != nil {
			log.Printf("Failed to store input of %v\n. Here's why: %v\n", input, err)
		}
	}

	return nil
}

func SeriesDataRequest(message map[string]events.SQSMessageAttribute, tableName string, ddbClient dynamodbiface.DynamoDBAPI) error {
	err := scraper.ScrapeSeriesData(message, tableName)
	return err
}

func ChapterListRequest(message map[string]events.SQSMessageAttribute, tableName string, ddbClient dynamodbiface.DynamoDBAPI) error {
	err := scraper.ScrapeChaptersList(message, tableName)
	return err
}

func ChapterDataRequest(message map[string]events.SQSMessageAttribute, tableName string, ddbClient dynamodbiface.DynamoDBAPI) error {
	err := scraper.ScrapeChaptersData(message, tableName)
	return err
}
