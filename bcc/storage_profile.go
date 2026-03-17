package bcc

import (
	"log"
	"net/url"

	"github.com/pkg/errors"
)

type StorageProfile struct {
	manager     *Manager
	ID          string `json:"id"`
	Name        string `json:"name"`
	MaxDiskSize int    `json:"max_disk_size"`
	Enabled     bool   `json:"enabled"`
}

func (v *Vdc) GetStorageProfiles(extraArgs ...Arguments) (storageProfiles []*StorageProfile, err error) {
	path := "v1/storage_profile"
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)

	if err = v.manager.GetItems(path, args, &storageProfiles); err != nil {
		log.Printf("[REQUEST-ERROR] get-storageProfiles was failed: %s", err)
	} else {
		for i := range storageProfiles {
			storageProfiles[i].manager = v.manager
		}
	}

	return
}

func (v *Vdc) GetStorageProfile(id string) (storageProfile *StorageProfile, err error) {
	path, _ := url.JoinPath("v1/storage_profile", id)
	args := Arguments{
		"vdc": v.ID,
	}

	if err = v.manager.Get(path, args, &storageProfile); err != nil {
		log.Printf("[REQUEST-ERROR] get-storageProfile was failed: %s", errors.WithStack(err))
	} else {
		storageProfile.manager = v.manager
	}

	return
}
