package handlers

import (
	"log"
	"manga-scraper-be-go/pkg/scraper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

func SeriesListRequest(provider *string, sourceUrl *string, tableName string, ddbClient dynamodbiface.DynamoDBAPI) error {
	data, err := scraper.ScrapeSeriesList(provider, sourceUrl, tableName)
	if err != nil {
		return err
	}

	for _, entry := range *data {
		item, err := dynamodbattribute.MarshalMap(entry)
		if err != nil {
			log.Printf("Failed to marshal entry of %v. Here's why: %v\n", entry.SeriesId, err)
			continue
		}

		cond := expression.AttributeNotExists(expression.Name("SeriesId"))
		expr, err := expression.NewBuilder().WithCondition(cond).Build()
		if err != nil {
			log.Printf("Failed to build expression for %v. Here's why: %v\n", entry.SeriesId, err)
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
			log.Printf("Failed to store input of %v. Here's why: %v\n", entry.SeriesId, err)
		} else {
			log.Printf("Finished storing input of %v.", entry.SeriesId)
			// _ = SeriesDataRequest(provider, &entry.SeriesUrl, tableName, ddbClient)
			// TODO add new queue to update series data
		}
	}

	return nil
}

func SeriesDataRequest(provider *string, sourceUrl *string, tableName string, ddbClient dynamodbiface.DynamoDBAPI) error {
	key, data, err := scraper.ScrapeSeriesData(provider, sourceUrl, tableName)
	if err != nil {
		return err
	}

	updateKey, err := dynamodbattribute.MarshalMap(key)
	if err != nil {
		log.Printf("Failed to marshal update key of %v. Here's why: %v\n", key.SeriesId, err)
		return err
	}

	input := &dynamodb.UpdateItemInput{
		Key:       updateKey,
		TableName: aws.String(tableName),
		ExpressionAttributeNames: map[string]*string{
			"#SC": aws.String("SeriesCover"),
			"#SU": aws.String("SeriesShortUrl"),
			"#SS": aws.String("SeriesSynopsis"),
			"#SD": aws.String("ScrapeDate"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sc": {S: aws.String(data.SeriesCover)},
			":su": {S: aws.String(data.SeriesShortUrl)},
			":ss": {S: aws.String(data.SeriesSynopsis)},
			":sd": {S: aws.String(data.ScrapeDate)},
		},
		UpdateExpression: aws.String("SET #SC = :sc, #SU = :su, #SS = :ss, #SD = :sd"),
	}

	_, err = ddbClient.UpdateItem(input)
	if err != nil {
		log.Printf("Failed to store input of %v. Here's why: %v\n", key.SeriesId, err)
		return err
	} else {
		log.Printf("Finished storing input of %v.", key.SeriesId)
		return nil
	}
}

func ChapterListRequest(provider *string, sourceUrl *string, tableName string, ddbClient dynamodbiface.DynamoDBAPI) error {
	err := scraper.ScrapeChaptersList(provider, sourceUrl, tableName)
	return err
}

func ChapterDataRequest(provider *string, sourceUrl *string, tableName string, ddbClient dynamodbiface.DynamoDBAPI) error {
	err := scraper.ScrapeChaptersData(provider, sourceUrl, tableName)
	return err
}
