package commands

import (
	"bufio"
	"builder/artifacts"
	"builder/buildlog"
	"builder/local"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/manifoldco/promptui"
	"golang.org/x/sync/errgroup"
)

var (
	backendTypeAWS         backendType = &awsBackendType{}
	backendTypeGoogleCloud backendType = &gcloudBackendType{}
)

type backendType interface {
	getManagerAndSourceSet(reader *bufio.Reader,
		sourceSetName string) (artifacts.Manager, artifacts.SourceSet, error)
}

type awsBackendType struct{}

type gcloudBackendType struct{}

type initWorkspace struct{}

func (i *initWorkspace) Describe() string {
	return "Initializes a workspace"
}

func (i *initWorkspace) Exec(workingDir string, args ...string) error {
	if workspaceDir, err := local.GetWorkspace(workingDir); err == nil {
		return fmt.Errorf("Workspace already exists at %s", workspaceDir)
	} else if err != local.ErrWorkspaceNotFound {
		return fmt.Errorf("Error validating no existing workspace: %+v", err)
	}

	reader := bufio.NewReader(os.Stdin)
	buildlog.Infof("Welcome to the builder system")
	sourceSetName := readLineWithPrompt("Source set name",
		artifacts.IsValidName, "")
	backendType, err := getBackendTypeFromUser()
	if err != nil {
		return fmt.Errorf("Error getting backend type: %+v", err)
	}

	buildlog.Infof("Please provide some info about the resources you'd like to use.")
	buildlog.Infof("If the resources don't exist, then they can be created for you.")
	manager, sourceSet, err := backendType.getManagerAndSourceSet(reader, sourceSetName)
	if err != nil {
		return err
	}

	if ok, err := getYnConfirmation("Create resources"); ok {
		group, _ := errgroup.WithContext(context.Background())
		group.Go(manager.Setup)
		group.Go(sourceSet.Setup)
		if err := group.Wait(); err != nil {
			return fmt.Errorf("Error creating manager and source set: %+v", err)
		}
	} else if err != nil {
		return fmt.Errorf("Error getting confirmation for resource creation: %+v", err)
	}

	if err = local.InitWorkspace(workingDir, sourceSet, manager); err != nil {
		return fmt.Errorf("Error initializing workspace: %+v", err)
	}

	return nil
}

func getBackendTypeFromUser() (backendType, error) {
	type backendOption struct {
		name        string
		backendType backendType
	}

	options := []backendOption{
		{
			name:        "AWS (DynamoDB and S3)",
			backendType: backendTypeAWS,
		},
		{
			name:        "Google Cloud (Datastore and GCS)",
			backendType: backendTypeGoogleCloud,
		},
	}

	items := make([]string, len(options))
	for i, option := range options {
		items[i] = option.name
	}

	prompt := promptui.Select{
		Label: "Select backend type",
		Items: items,
	}

	selectedIndex, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return options[selectedIndex].backendType, nil
}

func (a *awsBackendType) getManagerAndSourceSet(reader *bufio.Reader,
	sourceSetName string) (artifacts.Manager, artifacts.SourceSet, error) {
	bucketName := readLineWithPrompt("S3 bucket for artifact storage", artifacts.IsValidName, "")
	artifactTableName := readLineWithPrompt("Dynamo table name for artifact storage", artifacts.IsValidName,
		"builder-artifact-metadata")
	sourceSetTableName := readLineWithPrompt("Dynamo table name for source set metadata",
		artifacts.IsValidName, "builder-source-set-metadata")
	dynamoRegion := readLineWithPrompt("Dynamo region", artifacts.IsValidName, "us-east-1")
	profile := readLineWithPrompt("(Optional) AWS credentials profile",
		func(input string) error {
			if input == "" {
				return nil
			}
			return artifacts.IsValidName(input)
		}, "")
	buildlog.Infof(`

			S3 Bucket: %s
			Artifact Table: %s
			Source Set Table: %s
			Dynamo Region: %s
			AWS Profile: %s
			
			`, bucketName, artifactTableName, sourceSetTableName, dynamoRegion, profile)
	if ok, err := getYnConfirmation("Is this correct"); !ok || err != nil {
		return nil, nil, fmt.Errorf("User must re-enter information")
	}

	sess := local.NewSession(dynamoRegion, profile)
	s3Svc := s3.New(sess)

	manager, err := artifacts.NewS3Manager(s3Svc, bucketName, profile)
	if err != nil {
		return nil, nil, err
	}

	dynamoSvc := dynamodb.New(sess)
	sourceSet, err := artifacts.NewDynamoSourceSet(dynamoSvc, sourceSetName, sourceSetTableName,
		artifactTableName, profile)
	if err != nil {
		return nil, nil, err
	}

	return manager, sourceSet, err
}

func (a *gcloudBackendType) getManagerAndSourceSet(reader *bufio.Reader,
	sourceSetName string) (artifacts.Manager, artifacts.SourceSet, error) {
	return nil, nil, fmt.Errorf("Sorry! Google Cloud support is coming soon")
}
