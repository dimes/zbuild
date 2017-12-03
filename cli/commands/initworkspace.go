package commands

import (
	"bufio"
	"builder/artifacts"
	"builder/buildlog"
	"builder/local"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/manifoldco/promptui"
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

func (i *initWorkspace) Exec(workingDir string, args ...string) error {
	if workspaceDir, err := local.GetWorkspace(workingDir); err == nil {
		return fmt.Errorf("Workspace already exists at %s", workspaceDir)
	} else if err != local.ErrWorkspaceNotFound {
		return fmt.Errorf("Error validating no existing workspace: %+v", err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to the builder system")
	sourceSetName := readLineWithPrompt("What is the name of your source set? ",
		artifacts.IsValidName)
	backendType, err := getBackendTypeFromUser()
	if err != nil {
		return fmt.Errorf("Error getting backend type: %+v", err)
	}

	fmt.Println("Please provide some info about the resources you'd like to use.")
	fmt.Println("If the resources don't exist, then they can be created for you.")
	manager, sourceSet, err := backendType.getManagerAndSourceSet(reader, sourceSetName)

	fmt.Println("Should the about resources be created?")
	if ok, err := getYnConfirmation(); ok {
		fmt.Printf("here 1")
		if err := manager.Setup(); err != nil {
			buildlog.Fatalf("Error creating manager: %+v", err)
		}

		fmt.Printf("here 2")
		if err := sourceSet.Setup(); err != nil {
			buildlog.Fatalf("Error creating source set: %+v", err)
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
	fmt.Println("Where would you like to store your packages?")

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
	bucketName := readLineWithPrompt("S3 bucket for artifact storage: ", artifacts.IsValidName)
	artifactTableName := readLineWithPrompt("Dynamo table name for artifact storage: ", artifacts.IsValidName)
	sourceSetTableName := readLineWithPrompt("Dynamo table name for source set metadata: ",
		artifacts.IsValidName)
	dynamoRegion := readLineWithPrompt("Dynamo region: ", artifacts.IsValidName)
	profile := readLineWithPrompt("(Optional) What profile should be used for AWS service calls: ",
		func(input string) error {
			if input == "" {
				return nil
			}
			return artifacts.IsValidName(input)
		})
	fmt.Printf(`Does this look right?
			S3 Bucket: %s
			Artifact Table: %s
			Source Set Table: %s
			Dynamo Region: %s
			AWS Profile: %s
			`, bucketName, artifactTableName, sourceSetTableName, dynamoRegion, profile)
	if ok, err := getYnConfirmation(); !ok || err != nil {
		buildlog.Fatalf("Oops. Please try again")
	}

	sess := NewSession(dynamoRegion, profile)
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
