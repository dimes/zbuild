package artifacts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/dimes/zbuild/buildlog"
	"github.com/dimes/zbuild/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	// S3ManagerType is the type identifier for the S3 manager
	S3ManagerType = "s3"
)

// S3Metadata stores the metadata for the S3Manager
type S3Metadata struct {
	BucketName string `json:"bucketName"`
	Region     string `json:"region,omitempty"`
	Profile    string `json:"profile,omitempty"`
}

// S3Manager stores artifacts in S3
type S3Manager struct {
	svc      *s3.S3
	metadata *S3Metadata
}

// NewS3Manager returns a manager backed by S3
func NewS3Manager(svc *s3.S3, bucketName, region, profile string) (Manager, error) {
	metadata := &S3Metadata{
		BucketName: bucketName,
		Region:     region,
		Profile:    profile,
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
		Bucket: aws.String(s.metadata.BucketName),
	}

	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	_, err := s.svc.CreateBucketWithContext(ctx, createBucketInput)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); !ok || awsErr.Code() != s3.ErrCodeBucketAlreadyOwnedByYou {
			return fmt.Errorf("Error creating bucket %s: %+v", s.metadata.BucketName, err)
		}

		buildlog.Warningf("Bucket %s already existed. It will be used as is", s.metadata.BucketName)
	}

	return nil
}

// OpenReader opens a reader to an artifact stored in S3
func (s *S3Manager) OpenReader(artifact *model.Artifact) (io.ReadCloser, error) {
	artifactKey := s.artifactKey(artifact)
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.metadata.BucketName),
		Key:    aws.String(artifactKey),
	}

	output, err := s.svc.GetObject(input)
	if err != nil {
		return nil, fmt.Errorf("Error getting artifact %s: %+v", artifactKey, err)
	}

	return output.Body, nil
}

// OpenWriter opens a writer that can be used to write an artifact to S3
func (s *S3Manager) OpenWriter(artifact *model.Artifact) (io.WriteCloser, error) {
	artifactKey := s.artifactKey(artifact)
	headObjectInput := &s3.HeadObjectInput{
		Bucket: aws.String(s.metadata.BucketName),
		Key:    aws.String(artifactKey),
	}
	if _, err := s.svc.HeadObject(headObjectInput); err == nil {
		return nil, fmt.Errorf("The artifact %+v already exists", artifact)
	}

	buildlog.Infof("Opening writer to %+v", artifact)

	reader, writer := io.Pipe()
	uploadInput := &s3manager.UploadInput{
		Bucket: aws.String(s.metadata.BucketName),
		Key:    aws.String(artifactKey),
		Body:   reader,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer reader.Close()
		defer wg.Done()
		if _, err := s3manager.NewUploaderWithClient(s.svc).Upload(uploadInput); err != nil {
			buildlog.Errorf("Error uploading artifact: %+v", err)
			reader.CloseWithError(err)
			return
		}
	}()

	s3Writer := &s3Writer{
		WriteCloser: writer,
		wg:          &wg,
	}

	return s3Writer, nil
}

// PersistMetadata persists metadata for this source set to a writer so it can be read later
func (s *S3Manager) PersistMetadata(writer io.Writer) error {
	return json.NewEncoder(writer).Encode(s.metadata)
}

func (s *S3Manager) artifactKey(artifact *model.Artifact) string {
	return fmt.Sprintf("%s/%s/%s/%s", artifact.Namespace, artifact.Name, artifact.Version, artifact.BuildNumber)
}

type s3Writer struct {
	io.WriteCloser
	wg     *sync.WaitGroup
	closed bool
}

func (s *s3Writer) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true

	err := s.WriteCloser.Close()
	if err != nil {
		return err
	}

	buildlog.Infof("S3 writer closed. Waiting for upload to finish...")
	s.wg.Wait()
	return nil
}
