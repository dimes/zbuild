package artifacts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"golang.org/x/sync/errgroup"
)

const (
	// DynamoSourceSetType is the type identifier for Dynamo source sets
	DynamoSourceSetType = "dynamo"

	sourceSetKey   = "sourceSet"
	packageKey     = "package"
	artifactKey    = "artifact"
	buildNumberKey = "buildNumber"
)

// DynamoMetadata is the metadata for the DynamoDB client used by the source set.
type DynamoMetadata struct {
	Region         string `json:"region"`
	SourceSetTable string `json:"sourceSetTable"`
	ArtifactTable  string `json:"artifactTable"`
	Profile        string `json:"profile"`
}

// DynamoSourceSet uses DynamoDB to store package information
type DynamoSourceSet struct {
	svc           *dynamodb.DynamoDB
	sourceSetName string
	metadata      *DynamoMetadata
}

func newPackageKey(namespace, name, version string) string {
	return fmt.Sprintf("%s/%s/%s", namespace, name, version)
}

type sourceSetArtifactKey struct {
	SourceSet string `dynamodbav:"sourceSet,omitempty"`
	Package   string `dynamodbav:"package,omitempty"`
}

func newSourceSetArtifactKey(sourceSet, namespace, name, version string) sourceSetArtifactKey {
	return sourceSetArtifactKey{
		SourceSet: sourceSet,
		Package:   newPackageKey(namespace, name, version),
	}
}

type sourceSetArtifact struct {
	sourceSetArtifactKey
	Artifact    *model.Artifact `dynamodbav:"artifact,omitempty"`
	BuildNumber string          `dynamodbav:"buildNumber,omitempty"`
}

type dynamoArtifactKey struct {
	Package     string `dynamodbav:"package,omitempty"`
	BuildNumber string `dynamodbav:"buildNumber,omitempty"`
}

func newDynamoArtifactKey(namespace, name, version, buildNumber string) dynamoArtifactKey {
	return dynamoArtifactKey{
		Package:     newPackageKey(namespace, name, version),
		BuildNumber: buildNumber,
	}
}

type dynamoArtifact struct {
	dynamoArtifactKey
	Artifact *model.Artifact `dynamodbav:"artifact,omitempty"`
}

func newDynamoArtifact(artifact *model.Artifact) *dynamoArtifact {
	dynamoArtifactKey := newDynamoArtifactKey(
		artifact.Namespace,
		artifact.Name,
		artifact.Version,
		artifact.BuildNumber)
	return &dynamoArtifact{
		dynamoArtifactKey: dynamoArtifactKey,
		Artifact:          artifact,
	}
}

// NewDynamoSourceSet returns a source set backed by DyanmoDB
func NewDynamoSourceSet(svc *dynamodb.DynamoDB,
	sourceSetName,
	sourceSetTable,
	artifactTable,
	profile string) (SourceSet, error) {
	region := ""
	if svc.Config.Region != nil {
		region = *svc.Config.Region
	}

	metadata := &DynamoMetadata{
		Region:         region,
		SourceSetTable: sourceSetTable,
		ArtifactTable:  artifactTable,
		Profile:        profile,
	}

	return NewDynamoSourceSetFromMetadata(svc, sourceSetName, metadata)
}

// NewDynamoSourceSetFromMetadata returns a new dynamo-backed source set from metadata
func NewDynamoSourceSetFromMetadata(svc *dynamodb.DynamoDB,
	sourceSetName string,
	metadata *DynamoMetadata) (SourceSet, error) {
	return &DynamoSourceSet{
		svc:           svc,
		sourceSetName: sourceSetName,
		metadata:      metadata,
	}, nil
}

// Type returns the type identifier for Dynamo source sets
func (d *DynamoSourceSet) Type() string {
	return DynamoSourceSetType
}

// Setup sets up the required Dyanmo tables
func (d *DynamoSourceSet) Setup() error {
	group, _ := errgroup.WithContext(context.Background())
	group.Go(func() error {
		return d.createTableIfNotExists(d.metadata.SourceSetTable, sourceSetKey, packageKey)
	})

	group.Go(func() error {
		return d.createTableIfNotExists(d.metadata.ArtifactTable, packageKey, buildNumberKey)
	})

	if err := group.Wait(); err != nil {
		return fmt.Errorf("Error creating source set metadata tables: %+v", err)
	}

	return nil
}

func (d *DynamoSourceSet) createTableIfNotExists(table, hashKey, rangeKey string) error {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ctxCancel()

	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: aws.String(table),
	}
	_, err := d.svc.DescribeTableWithContext(ctx, describeTableInput)
	if err == nil {
		buildlog.Warningf("Table %s already existed. It will be used as is", table)
		return nil
	}

	if awsErr, ok := err.(awserr.Error); !ok || awsErr.Code() != dynamodb.ErrCodeResourceNotFoundException {
		return fmt.Errorf("Error checking existence of table %s: %+v", table, err)
	}

	attributeDefinitions := []*dynamodb.AttributeDefinition{
		{
			AttributeName: aws.String(hashKey),
			AttributeType: aws.String(dynamodb.ScalarAttributeTypeS),
		},
	}

	keySchema := []*dynamodb.KeySchemaElement{
		{
			AttributeName: aws.String(hashKey),
			KeyType:       aws.String(dynamodb.KeyTypeHash),
		},
	}

	if rangeKey != "" {
		attributeDefinitions = append(attributeDefinitions, &dynamodb.AttributeDefinition{
			AttributeName: aws.String(rangeKey),
			AttributeType: aws.String(dynamodb.ScalarAttributeTypeS),
		})

		keySchema = append(keySchema, &dynamodb.KeySchemaElement{
			AttributeName: aws.String(rangeKey),
			KeyType:       aws.String(dynamodb.KeyTypeRange),
		})
	}

	createTableInput := &dynamodb.CreateTableInput{
		AttributeDefinitions: attributeDefinitions,
		KeySchema:            keySchema,
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableName: aws.String(table),
	}

	if _, err := d.svc.CreateTableWithContext(ctx, createTableInput); err != nil {
		return fmt.Errorf("Error creating table %s: %+v", table, err)
	}

	createCtx, createCtxCancel := context.WithTimeout(context.Background(), time.Minute)
	defer createCtxCancel()

	for describeTableOutput, err := d.svc.DescribeTableWithContext(createCtx, describeTableInput); //
	err != nil || *describeTableOutput.Table.TableStatus != dynamodb.TableStatusActive;            //
	describeTableOutput, err = d.svc.DescribeTableWithContext(createCtx, describeTableInput) {
		time.Sleep(5 * time.Second)
	}

	return nil
}

// Name returns the name of the source set
func (d *DynamoSourceSet) Name() string {
	return d.sourceSetName
}

// GetArtifact returns an artifact stored in the database. If the artifact is not in the
// source set, then an error is returned.
func (d *DynamoSourceSet) GetArtifact(namespace, name, version string) (*model.Artifact, error) {
	sourceSetArtifactKey := newSourceSetArtifactKey(d.sourceSetName, namespace, name, version)
	key, err := dynamodbattribute.ConvertToMap(sourceSetArtifactKey)
	if err != nil {
		return nil, fmt.Errorf("Error serializing key: %+v", err)
	}

	getItemInput := &dynamodb.GetItemInput{
		TableName:      aws.String(d.metadata.SourceSetTable),
		Key:            key,
		ConsistentRead: aws.Bool(true),
	}

	item, err := d.svc.GetItem(getItemInput)
	if err != nil {
		return nil, fmt.Errorf("Error getting artifact %s: %+v", name, err)
	}

	if item == nil || item.Item == nil || item.Item[artifactKey] == nil {
		return nil, ErrArtifactNotFound
	}

	artifact := &model.Artifact{}
	if err = dynamodbattribute.ConvertFrom(item.Item[artifactKey], artifact); err != nil {
		return nil, fmt.Errorf("Error convrting dynamo item to artifact: %+v", err)
	}

	return artifact, nil
}

// GetAllArtifacts returns all artifacts in this source set
func (d *DynamoSourceSet) GetAllArtifacts() ([]*model.Artifact, error) {
	expressionValues := map[string]*dynamodb.AttributeValue{
		":sourceSetName": {
			S: aws.String(d.sourceSetName),
		},
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(d.metadata.SourceSetTable),
		KeyConditionExpression:    aws.String(fmt.Sprintf("%s = :sourceSetName", sourceSetKey)),
		ExpressionAttributeValues: expressionValues,
	}

	queryOutput, err := d.svc.Query(queryInput)
	if err != nil {
		return nil, fmt.Errorf("Error getting artifacts: %+v", err)
	}

	artifacts := make([]*model.Artifact, 0)
	for _, item := range queryOutput.Items {
		if item[artifactKey] == nil {
			continue
		}

		artifact := &model.Artifact{}
		if err = dynamodbattribute.ConvertFrom(item[artifactKey], artifact); err != nil {
			return nil, fmt.Errorf("Error convrting dynamo item to artifact: %+v", err)
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}

// RegisterArtifact registers an artifact as available for consumptions by any source set
func (d *DynamoSourceSet) RegisterArtifact(artifact *model.Artifact) error {
	dynamoArtifact := newDynamoArtifact(artifact)
	item, err := dynamodbattribute.ConvertToMap(dynamoArtifact)
	if err != nil {
		return fmt.Errorf("Error marshaling artifact %+v: %+v", artifact, err)
	}

	putItemInput := &dynamodb.PutItemInput{
		TableName:           aws.String(d.metadata.ArtifactTable),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(package)"),
	}

	if _, err := d.svc.PutItem(putItemInput); err != nil {
		return fmt.Errorf("Error persisting artifact %+v: %+v", artifact, err)
	}

	return nil
}

// PersistMetadata persists metadata for this source set to a writer so it can be read later
func (d *DynamoSourceSet) PersistMetadata(writer io.Writer) error {
	return json.NewEncoder(writer).Encode(d.metadata)
}
