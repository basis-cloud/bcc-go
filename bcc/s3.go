package bcc

import (
	"fmt"
	"log"
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

func NewS3StorageBucket(name string) S3StorageBucket {
	return S3StorageBucket{
		Name: name,
	}
}

func (p *Project) CreateS3Storage(s3 *S3Storage) (err error) {
	path := "v1/s3_storage"
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

	if err = p.manager.Request("POST", path, args, &s3); err != nil {
		log.Printf("[REQUEST-ERROR] create-s3Storage was failed: %s", err)
	} else {
		s3.manager = p.manager
	}

	return
}

func (m *Manager) GetS3Storages(extraArgs ...Arguments) (s3Storages []*S3Storage, err error) {
	path := "v1/s3_storage"
	args := Defaults()
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &s3Storages); err != nil {
		log.Printf("[REQUEST-ERROR] get-s3Storages was failed: %s", err)
	} else {
		for i := range s3Storages {
			s3Storages[i].manager = m
		}
	}

	return
}

func (p *Project) GetS3Storages(extraArgs ...Arguments) (s3Storages []*S3Storage, err error) {
	args := Arguments{
		"project": p.ID,
	}
	args.merge(extraArgs)
	s3Storages, err = p.manager.GetS3Storages(args)
	return
}

func (m *Manager) GetS3Storage(id string) (s3Storages *S3Storage, err error) {
	path, _ := url.JoinPath("v1/s3_storage", id)

	if err = m.Get(path, Defaults(), &s3Storages); err != nil {
		log.Printf("[REQUEST-ERROR] get-s3Storage was failed: %s", err)
	} else {
		s3Storages.manager = m
	}

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

	if err = s3.manager.Request("PUT", path, args, s3); err != nil {
		log.Printf("[REQUEST-ERROR] update-s3Storage was failed: %s", err)
	} else {
		s3.WaitLock()
	}

	return
}

func (s3 *S3Storage) Delete() (err error) {
	path, _ := url.JoinPath("v1/s3_storage", s3.ID)
	err = s3.manager.Delete(path, Defaults(), nil)
	return
}

func (s3 *S3Storage) CreateBucket(bucket *S3StorageBucket) (err error) {
	path := fmt.Sprintf("v1/s3_storage/%s/bucket", s3.ID)
	args := &struct {
		Name string `json:"name"`
	}{
		Name: bucket.Name,
	}

	if err = s3.manager.Request("POST", path, args, &bucket); err != nil {
		log.Printf("[REQUEST-ERROR] create-bucket was failed: %s", err)
	} else {
		bucket.manager = s3.manager
		bucket.S3StorageId = s3.ID
	}

	return
}

func (m *Manager) GetBuckets(id string, extraArgs ...Arguments) (buckets []*S3StorageBucket, err error) {
	path := fmt.Sprintf("v1/s3_storage/%s/bucket", id)
	args := Defaults()
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &buckets); err != nil {
		log.Printf("[REQUEST-ERROR] get-buckets was failed: %s", err)
	} else {
		for i := range buckets {
			buckets[i].manager = m
		}
	}

	return
}

func (s3 *S3Storage) GetBuckets(extraArgs ...Arguments) (buckets []*S3StorageBucket, err error) {
	buckets, err = s3.manager.GetBuckets(s3.ID, extraArgs...)
	return
}

func (s3 *S3Storage) GetBucket(id string) (bucket *S3StorageBucket, err error) {
	path := fmt.Sprintf("v1/s3_storage/%s/bucket/%s", s3.ID, id)

	if err = s3.manager.Get(path, Defaults(), &bucket); err != nil {
		log.Printf("[REQUEST-ERROR] get-bucket was failed: %s", err)
	} else {
		bucket.manager = s3.manager
		bucket.S3StorageId = s3.ID
	}

	return
}

func (b *S3StorageBucket) Update() (err error) {
	path := fmt.Sprintf("v1/s3_storage/%s/bucket/%s", b.S3StorageId, b.ID)
	args := &struct {
		Name string `json:"name"`
	}{
		Name: b.Name,
	}

	if err = b.manager.Request("PUT", path, args, b); err != nil {
		log.Printf("[REQUEST-ERROR] update-bucket was failed: %s", err)
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
