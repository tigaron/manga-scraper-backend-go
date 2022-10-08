package handlers

import (
	"log"
	"manga-scraper-be-go/pkg/scraper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/google/uuid"
)

func SeriesListRequest(provider *string, sourceUrl *string, tableName string, ddbClient dynamodbiface.DynamoDBAPI, queueUrl string, sqsClient sqsiface.SQSAPI) error {
	data, err := scraper.ScrapeSeriesList(provider, sourceUrl, tableName)
	if err != nil {
		return err
	}

	var queues []*sqs.SendMessageBatchRequestEntry
	for _, entry := range data {
		item, err := dynamodbattribute.MarshalMap(entry)
		if err != nil {
			log.Printf("Failed to marshal entry of %v. Here's why: %v\n", entry.SeriesId, err)
		}

		cond := expression.AttributeNotExists(expression.Name("SeriesId"))
		expr, err := expression.NewBuilder().WithCondition(cond).Build()
		if err != nil {
			log.Printf("Failed to build expression for %v. Here's why: %v\n", entry.SeriesId, err)
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
			queue := &sqs.SendMessageBatchRequestEntry{
				Id:                     aws.String(uuid.NewString()),
				MessageDeduplicationId: aws.String(uuid.NewSHA1(uuid.NameSpaceURL, []byte(entry.SeriesUrl)).String()), // TODO add time window of 1 minute
				MessageGroupId:         provider,
				MessageBody:            aws.String("series-data of " + entry.SeriesId),
				MessageAttributes: map[string]*sqs.MessageAttributeValue{
					"RequestType": {
						DataType:    aws.String("String"),
						StringValue: aws.String("series-data"),
					},
					"Provider": {
						DataType:    aws.String("String"),
						StringValue: provider,
					},
					"SourceUrl": {
						DataType:    aws.String("String"),
						StringValue: aws.String(entry.SeriesUrl),
					},
				},
			}
			queues = append(queues, queue)
		}
	}

	batch := 10
	for i := 0; i < len(queues); i += batch {
		j := i + batch
		if j > len(queues) {
			j = len(queues)
		}
		input := &sqs.SendMessageBatchInput{
			QueueUrl: aws.String(queueUrl),
			Entries:  queues[i:j],
		}
		_, err := sqsClient.SendMessageBatch(input)
		if err != nil {
			log.Printf("Failed to send batch message. Here's why: %v\n", err)
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

func ChapterListRequest(provider *string, sourceUrl *string, tableName string, ddbClient dynamodbiface.DynamoDBAPI, queueUrl string, sqsClient sqsiface.SQSAPI) error {
	err := scraper.ScrapeChaptersList(provider, sourceUrl, tableName)
	return err
}

func ChapterDataRequest(provider *string, sourceUrl *string, tableName string, ddbClient dynamodbiface.DynamoDBAPI) error {
	err := scraper.ScrapeChaptersData(provider, sourceUrl, tableName)
	return err
}
