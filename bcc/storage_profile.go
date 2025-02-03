package bcc

import (
	"net/url"
)

type StorageProfile struct {
	manager     *Manager
	ID          string `json:"id"`
	Name        string `json:"name"`
	MaxDiskSize int    `json:"max_disk_size"`
}

func (v *Vdc) GetStorageProfiles() (storageProfiles []*StorageProfile, err error) {
	args := Arguments{
		"vdc": v.ID,
	}

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
	err = v.manager.Get(path, args, &storageProfile)
	if err != nil {
		return
	}
	storageProfile.manager = v.manager
	return
}
