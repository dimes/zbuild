package artifacts

import (
	"builder/buildlog"
	"builder/model"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/lox/patchwork"
)

const (
	// S3ManagerType is the type identifier for the S3 manager
	S3ManagerType = "s3"
)

// S3Metadata stores the metadata for the S3Manager
type S3Metadata struct {
	bucketName string
	profile    string
}

// S3Manager stores artifacts in S3
type S3Manager struct {
	svc      *s3.S3
	metadata *S3Metadata
}

// NewS3Manager returns a manager backed by S3
func NewS3Manager(svc *s3.S3, bucketName, profile string) (Manager, error) {
	metadata := &S3Metadata{
		bucketName: bucketName,
		profile:    profile,
	}

	return NewS3ManagerFromMetadata(svc, metadata)
}

// NewS3ManagerFromMetadata returns an S3-backed manager from the given metadata
func NewS3ManagerFromMetadata(svc *s3.S3, metadata *S3Metadata) (Manager, error) {
	return &S3Manager{
		svc:      svc,
		metadata: metadata,
	}, nil
}

// Type returns the S3 manager type
func (s *S3Manager) Type() string {
	return S3ManagerType
}

// Setup creates the bucket with the given name
func (s *S3Manager) Setup() error {
	createBucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(s.metadata.bucketName),
	}

	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	_, err := s.svc.CreateBucketWithContext(ctx, createBucketInput)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); !ok || awsErr.Code() != s3.ErrCodeBucketAlreadyOwnedByYou {
			return fmt.Errorf("Error creating bucket %s: %+v", s.metadata.bucketName, err)
		}

		buildlog.Warningf("Bucket %s already existed. It will be used as is", s.metadata.bucketName)
	}

	return nil
}

// OpenReader opens a reader to an artifact stored in S3
func (s *S3Manager) OpenReader(artifact *model.Artifact) (io.ReadCloser, error) {
	buffer, err := patchwork.NewFileBuffer(128 * 1024 * 1024)
	if err != nil {
		return nil, fmt.Errorf("Error creating buffer: %+v", err)
	}
	patchwork := patchwork.New(buffer)

	artifactKey := s.artifactKey(artifact)
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.metadata.bucketName),
		Key:    aws.String(artifactKey),
	}

	go func() {
		s3manager.NewDownloaderWithClient(s.svc).Download(patchwork, input)
		patchwork.Close()
	}()

	return ioutil.NopCloser(patchwork.Reader()), nil
}

// OpenWriter opens a writer that can be used to write an artifact to S3
func (s *S3Manager) OpenWriter(artifact *model.Artifact) (io.WriteCloser, error) {
	reader, writer := io.Pipe()
	artifactKey := s.artifactKey(artifact)
	uploadInput := &s3manager.UploadInput{
		Bucket: aws.String(s.metadata.bucketName),
		Key:    aws.String(artifactKey),
		Body:   reader,
	}

	go func() {
		defer reader.Close()
		if _, err := s3manager.NewUploaderWithClient(s.svc).Upload(uploadInput); err != nil {
			reader.CloseWithError(err)
			return
		}
	}()

	return writer, nil
}

func (s *S3Manager) artifactKey(artifact *model.Artifact) string {
	return fmt.Sprintf("%s/%s/%s/%s", artifact.Namespace, artifact.Name, artifact.Version, artifact.BuildNumber)
}

// PersistMetadata persists metadata for this source set to a writer so it can be read later
func (s *S3Manager) PersistMetadata(writer io.Writer) error {
	return json.NewEncoder(writer).Encode(s.metadata)
}
