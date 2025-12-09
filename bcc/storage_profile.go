package bcc

import (
	"net/url"

	"github.com/pkg/errors"
)

type StorageProfile struct {
	manager     *Manager
	ID          string `json:"id"`
	Name        string `json:"name"`
	MaxDiskSize int    `json:"max_disk_size"`
}

func (v *Vdc) GetStorageProfiles(extraArgs ...Arguments) (storageProfiles []*StorageProfile, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)

	path := "v1/storage_profile"
	err = v.manager.GetItems(path, args, &storageProfiles)
	for i := range storageProfiles {
		storageProfiles[i].manager = v.manager
	}
	return
}

func (v *Vdc) GetStorageProfile(id string) (storageProfile *StorageProfile, err error) {
	args := Arguments{
		"vdc": v.ID,
	}

	path, _ := url.JoinPath("v1/storage_profile", id)
	if err = v.manager.Get(path, args, &storageProfile); err != nil {
		return nil, errors.Wrapf(err, "crash via getting StorageProfile-%s", id)
	} else {
		storageProfile.manager = v.manager
		return
	}
}
