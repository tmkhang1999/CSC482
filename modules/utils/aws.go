package utils

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/oleiade/reflections"
	"log"
	"main.go/structs"
	"reflect"
	"time"
)

func CreateDynamoDBClient(region string) *dynamodb.DynamoDB {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	},
	)
	if err != nil {
		log.Fatalf("Got error initializing AWS: %s", err)
	}
	svc := dynamodb.New(sess)

	return svc
}

func createTable(tableName string, client *dynamodb.DynamoDB) {
	// Create the table input in DynamoDB
	tableInput := &dynamodb.CreateTableInput{

		// Define table name
		TableName: aws.String(tableName),

		// Represents an attribute for describing the key schema for the table and indexes.
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("name"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("localtime"),
				AttributeType: aws.String("S"),
			},
		},

		KeySchema: []*dynamodb.KeySchemaElement{
			// Partition key
			{
				AttributeName: aws.String("name"),
				KeyType:       aws.String("HASH"),
			},
			// Sort key
			{
				AttributeName: aws.String("localtime"),
				KeyType:       aws.String("RANGE"),
			},
		},

		// Throughput settings
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	}

	if _, err := client.CreateTable(tableInput); err != nil {
		log.Fatal("Creating table error: ", err.Error())
	}

	log.Printf("Table %s successfully created!", tableName)
}

func deleteTable(tableName string, client *dynamodb.DynamoDB) {
	// Create ListTableInput parameter, then get the list of table output
	listTableInput := &dynamodb.ListTablesInput{}
	listTableOutput, err := client.ListTables(listTableInput)
	if err != nil {
		log.Fatal("Listing table error: ", err.Error())
	}

	// Check if the provided table name exists
	for _, existedTableName := range listTableOutput.TableNames {

		// Delete table if true
		if *existedTableName == tableName {
			log.Printf("Table %s already exists.", tableName)
			deleteTableInput := &dynamodb.DeleteTableInput{
				TableName: aws.String(tableName),
			}

			if _, err := client.DeleteTable(deleteTableInput); err != nil {
				log.Fatal("Deleting table error: ", err.Error())
			}
		}
	}
}

func SetUpTableAWS(tableName string, awsClient *dynamodb.DynamoDB) {
	// Check if the table exists. If yes, delete it
	deleteTable(tableName, awsClient)

	// Wait 10 seconds for AWS to delete the table
	time1 := time.NewTimer(10 * time.Second)
	<-time1.C

	// create a new table with the provided table name
	createTable(tableName, awsClient)

	// Wait 10 seconds for AWS to create the table
	timer2 := time.NewTimer(10 * time.Second)
	<-timer2.C
}

func PutItemInput(tableName string, body structs.Data, awsClient *dynamodb.DynamoDB) {
	// Create the item to add to DynamoDB on AWS
	item := createItemForAWS(body)
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		log.Fatalf("Got error marshalling new network item: %s", err)
	}

	// Create PutItemInput parameter
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	// Put the item input to the DynamoDB table
	_, err = awsClient.PutItem(input)
	if err != nil {
		log.Fatalf("Got error calling PutItem: %s", err)
	}
	fmt.Println("Successfully added to table " + tableName)
}

func createItemForAWS(data structs.Data) structs.Item {
	// Create a new item to send to DynamoDB
	var item structs.Item
	valueItem := reflect.ValueOf(&item).Elem()

	// Get the values of "Location" and "Current" structs in "Data" struct,
	// then pass them into "Item" struct
	fieldNames := []string{"Location", "Current"}
	for _, fieldName := range fieldNames {
		table, _ := reflections.GetField(data, fieldName)
		values := reflect.ValueOf(table)
		types := values.Type()
		for i := 0; i < values.NumField(); i++ {
			valueItem.FieldByName(types.Field(i).Name).Set(values.Field(i))
		}
	}

	return item
}
