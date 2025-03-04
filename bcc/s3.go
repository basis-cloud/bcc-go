package bcc

import (
	"fmt"
	"net/url"
)

type S3Storage struct {
	manager        *Manager
	ID             string `json:"id"`
	Locked         bool   `json:"locked"`
	JobId          string `json:"job_id"`
	ClientEndpoint string `json:"client_endpoint"`
	AccessKey      string `json:"access_key"`
	SecretKey      string `json:"secret_key"`
	Backend        string `json:"backend"`

	Name    string   `json:"name"`
	Project *Project `json:"project"`
	Tags    []Tag    `json:"tags"`
}

type S3StorageBucket struct {
	manager      *Manager
	ID           string `json:"id"`
	ExternalName string `json:"external_name"`
	S3StorageId  string

	Name string `json:"name"`
}

func NewS3Storage(name string, backend string) S3Storage {
	return S3Storage{
		Name:    name,
		Backend: backend,
	}
}

func (p *Project) CreateS3Storage(s3 *S3Storage) (err error) {
	args := &struct {
		Name    string   `json:"name"`
		Project string   `json:"project"`
		Backend string   `json:"backend"`
		Tags    []string `json:"tags"`
	}{
		Name:    s3.Name,
		Project: p.ID,
		Backend: s3.Backend,
		Tags:    convertTagsToNames(s3.Tags),
	}

	path := "v1/s3_storage"
	err = p.manager.Request("POST", path, args, &s3)
	s3.manager = p.manager
	return
}

func (m *Manager) GetS3Storages(extraArgs ...Arguments) (s3_storages []*S3Storage, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/s3_storage"
	err = m.GetItems(path, args, &s3_storages)
	for i := range s3_storages {
		s3_storages[i].manager = m
	}
	return
}

func (p *Project) GetS3Storages(extraArgs ...Arguments) (s3_storages []*S3Storage, err error) {
	args := Arguments{
		"project": p.ID,
	}
	args.merge(extraArgs)
	s3_storages, err = p.manager.GetS3Storages(args)
	return
}

func (m *Manager) GetS3Storage(id string) (s3_storage *S3Storage, err error) {
	path, _ := url.JoinPath("v1/s3_storage", id)
	err = m.Get(path, Defaults(), &s3_storage)
	if err != nil {
		return
	}
	s3_storage.manager = m
	return
}

func (s3 *S3Storage) Update() (err error) {
	path, _ := url.JoinPath("v1/s3_storage", s3.ID)
	args := &struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}{
		Name: s3.Name,
		Tags: convertTagsToNames(s3.Tags),
	}

	err = s3.manager.Request("PUT", path, args, s3)
	s3.WaitLock()
	return
}

func (s3 *S3Storage) Delete() (err error) {
	path, _ := url.JoinPath("v1/s3_storage", s3.ID)
	err = s3.manager.Delete(path, Defaults(), nil)
	return
}

func NewS3StorageBucket(name string) S3StorageBucket {
	return S3StorageBucket{
		Name: name,
	}
}

func (s3 *S3Storage) CreateBucket(bucket *S3StorageBucket) (err error) {
	args := &struct {
		Name string `json:"name"`
	}{
		Name: bucket.Name,
	}

	path := fmt.Sprintf("v1/s3_storage/%s/bucket", s3.ID)
	err = s3.manager.Request("POST", path, args, &bucket)
	bucket.manager = s3.manager
	bucket.S3StorageId = s3.ID
	return
}

func (m *Manager) GetBuckets(s3_id string, extraArgs ...Arguments) (buckets []*S3StorageBucket, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := fmt.Sprintf("v1/s3_storage/%s/bucket", s3_id)
	err = m.GetItems(path, args, &buckets)
	for i := range buckets {
		buckets[i].manager = m
	}
	return
}

func (s3 *S3Storage) GetBuckets(extraArgs ...Arguments) (buckets []*S3StorageBucket, err error) {
	buckets, err = s3.manager.GetBuckets(s3.ID, extraArgs...)
	if err != nil {
		return nil, err
	}
	return buckets, nil
}

func (s3 *S3Storage) GetBucket(id string) (bucket *S3StorageBucket, err error) {
	path := fmt.Sprintf("v1/s3_storage/%s/bucket/%s", s3.ID, id)
	err = s3.manager.Get(path, Defaults(), &bucket)
	if err != nil {
		return
	}
	bucket.manager = s3.manager
	bucket.S3StorageId = s3.ID
	return
}

func (b *S3StorageBucket) Update() (err error) {
	args := &struct {
		Name string `json:"name"`
	}{
		Name: b.Name,
	}
	path := fmt.Sprintf("v1/s3_storage/%s/bucket/%s", b.S3StorageId, b.ID)
	err = b.manager.Request("PUT", path, args, b)
	if err != nil {
		return err
	}
	return
}

func (b *S3StorageBucket) Delete() (err error) {
	path := fmt.Sprintf("v1/s3_storage/%s/bucket/%s", b.S3StorageId, b.ID)
	err = b.manager.Delete(path, Defaults(), nil)
	return
}

func (s3 S3Storage) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/s3_storage", s3.ID)
	return loopWaitLock(s3.manager, path)
}
