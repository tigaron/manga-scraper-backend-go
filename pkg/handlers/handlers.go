package handlers

import (
	"log"
	"manga-scraper-be-go/pkg/scraper"
	"strconv"
	"time"

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
	data, err := scraper.ScrapeSeriesList(provider, sourceUrl)
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
				MessageDeduplicationId: aws.String(strconv.Itoa(time.Now().Minute()) + "-" + uuid.NewSHA1(uuid.NameSpaceURL, []byte(entry.SeriesUrl)).String()),
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
	key, data, err := scraper.ScrapeSeriesData(provider, sourceUrl)
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
	data, err := scraper.ScrapeChapterList(provider, sourceUrl)
	if err != nil {
		return err
	}

	var queues []*sqs.SendMessageBatchRequestEntry
	for _, entry := range data {
		item, err := dynamodbattribute.MarshalMap(entry)
		if err != nil {
			log.Printf("Failed to marshal entry of %v. Here's why: %v\n", entry.ChapterId, err)
		}

		cond := expression.AttributeNotExists(expression.Name("ChapterId"))
		expr, err := expression.NewBuilder().WithCondition(cond).Build()
		if err != nil {
			log.Printf("Failed to build expression for %v. Here's why: %v\n", entry.ChapterId, err)
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
			log.Printf("Failed to store input of %v. Here's why: %v\n", entry.ChapterId, err)
		} else {
			log.Printf("Finished storing input of %v.", entry.ChapterId)
			queue := &sqs.SendMessageBatchRequestEntry{
				Id:                     aws.String(uuid.NewString()),
				MessageDeduplicationId: aws.String(strconv.Itoa(time.Now().Minute()) + "-" + uuid.NewSHA1(uuid.NameSpaceURL, []byte(entry.ChapterUrl)).String()),
				MessageGroupId:         provider,
				MessageBody:            aws.String("chapter-data of " + entry.ChapterId),
				MessageAttributes: map[string]*sqs.MessageAttributeValue{
					"RequestType": {
						DataType:    aws.String("String"),
						StringValue: aws.String("chapter-data"),
					},
					"Provider": {
						DataType:    aws.String("String"),
						StringValue: provider,
					},
					"SourceUrl": {
						DataType:    aws.String("String"),
						StringValue: aws.String(entry.ChapterUrl),
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

func ChapterDataRequest(provider *string, sourceUrl *string, tableName string, ddbClient dynamodbiface.DynamoDBAPI) error {
	key, data, err := scraper.ScrapeChapterData(provider, sourceUrl)
	if err != nil {
		return err
	}

	updateKey, err := dynamodbattribute.MarshalMap(key)
	if err != nil {
		log.Printf("Failed to marshal update key of %v. Here's why: %v\n", key.ChapterId, err)
		return err
	}

	input := &dynamodb.UpdateItemInput{
		Key:       updateKey,
		TableName: aws.String(tableName),
		ExpressionAttributeNames: map[string]*string{
			"#CT": aws.String("ChapterTitle"),
			"#CU": aws.String("ChapterShortUrl"),
			"#CP": aws.String("ChapterPrev"),
			"#CN": aws.String("ChapterNext"),
			"#CC": aws.String("ChapterContent"),
			"#SD": aws.String("ScrapeDate"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":ct": {S: aws.String(data.ChapterTitle)},
			":cu": {S: aws.String(data.ChapterShortUrl)},
			":cp": {S: aws.String(data.ChapterPrev)},
			":cn": {S: aws.String(data.ChapterNext)},
			":cc": {S: aws.StringSlice(data.ChapterContent)[0]},
			":sd": {S: aws.String(data.ScrapeDate)},
		},
		UpdateExpression: aws.String("SET #CT = :ct, #CU = :cu, #CP = :cp, #CN = :cn, #CC = :cc, #SD = :sd"),
	}

	_, err = ddbClient.UpdateItem(input)
	if err != nil {
		log.Printf("Failed to store input of %v. Here's why: %v\n", key.ChapterId, err)
		return err
	} else {
		log.Printf("Finished storing input of %v.", key.ChapterId)
		return nil
	}
}
