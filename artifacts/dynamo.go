package artifacts

import (
	"builder/model"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	sourceSetNameKey = "sourceSetName"
	packageKey       = "package"
	artifactItemKey  = "artifact"
)

// DynamoSourceSet uses DynamoDB to store package information
type DynamoSourceSet struct {
	svc            *dynamodb.DynamoDB
	sourceSetName  string
	sourceSetTable string
	artifactTable  string
}

// NewDynamoSourceSet returns a source set backed by DyanmoDB
func NewDynamoSourceSet(svc *dynamodb.DynamoDB,
	sourceSetName,
	sourceSetTable,
	artifactTable string) (SourceSet, error) {
	return &DynamoSourceSet{
		svc:            svc,
		sourceSetName:  sourceSetName,
		sourceSetTable: sourceSetTable,
		artifactTable:  artifactTable,
	}, nil
}

// Name returns the name of the source set
func (d *DynamoSourceSet) Name() string {
	return d.sourceSetName
}

// GetArtifact returns an artifact stored in the database. If the artifact is not in the
// source set, then an error is returned.
func (d *DynamoSourceSet) GetArtifact(namespace, name, version string) (*model.Artifact, error) {
	artifactRequest := new(dynamodb.GetItemInput).
		SetTableName(d.sourceSetTable).
		SetKey(d.sourceSetPackageKey(namespace, name, version)).
		SetConsistentRead(true)
	item, err := d.svc.GetItem(artifactRequest)
	if err != nil {
		return nil, fmt.Errorf("Error getting artifact %s: %+v", name, err)
	}

	if item == nil || item.Item == nil || item.Item[artifactItemKey] == nil {
		return nil, ErrArtifactNotFound
	}

	artifact := &model.Artifact{}
	if err = dynamodbattribute.ConvertFrom(item.Item[artifactItemKey], artifact); err != nil {
		return nil, fmt.Errorf("Error convrting dynamo item to artifact: %+v", err)
	}

	return artifact, nil
}

// GetAllArtifacts returns all artifacts in this source set
func (d *DynamoSourceSet) GetAllArtifacts() ([]*model.Artifact, error) {
	queryInput := new(dynamodb.QueryInput).
		SetTableName(d.sourceSetTable).
		SetKeyConditionExpression(fmt.Sprintf("%s = :sourceSetName", sourceSetNameKey)).
		SetExpressionAttributeValues(d.sourceSetKey(":sourceSetName"))
	queryOutput, err := d.svc.Query(queryInput)
	if err != nil {
		return nil, fmt.Errorf("Error getting artifacts: %+v", err)
	}

	artifacts := make([]*model.Artifact, 0)
	for _, item := range queryOutput.Items {
		artifact := &model.Artifact{}
		if err = dynamodbattribute.ConvertFrom(item[artifactItemKey], artifact); err != nil {
			return nil, fmt.Errorf("Error convrting dynamo item to artifact: %+v", err)
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}

// sourceSetKey creates a key into the source set artifact table (used to query entire source set)
func (d *DynamoSourceSet) sourceSetKey(key string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		key: {
			S: aws.String(d.Name()),
		},
	}
}

func (d *DynamoSourceSet) sourceSetPackageKey(namespace, name,
	version string) map[string]*dynamodb.AttributeValue {
	key := d.sourceSetKey(sourceSetNameKey)
	packageValue := fmt.Sprintf("%s/%s/%s", namespace, name, version)
	key[packageKey] = &dynamodb.AttributeValue{
		S: aws.String(packageValue),
	}
	return key
}
