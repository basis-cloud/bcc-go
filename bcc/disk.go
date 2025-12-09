package bcc

import (
	"fmt"
	"log"
	"net/url"

	"github.com/pkg/errors"
)

type TmpVm struct {
	manager  *Manager
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Cpu      int     `json:"cpu"`
	Ram      float64 `json:"ram"`
	Power    bool    `json:"power"`
	Platform string  `json:"platform,omitempty"`
	Vdc      *Vdc    `json:"vdc"`
}

type Disk struct {
	manager        *Manager
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Scsi           string          `json:"scsi"`
	ExternalID     string          `json:"external_id"`
	IsRoot         bool            `json:"is_root"`
	Size           int             `json:"size"`
	Vdc            *Vdc            `json:"vdc,omitempty"`
	Vm             *TmpVm          `json:"vm"`
	StorageProfile *StorageProfile `json:"storage_profile"`
	Locked         bool            `json:"locked,omitempty"`
	Tags           []Tag           `json:"tags"`
}

func NewDisk(name string, size int, storageProfile *StorageProfile) Disk {
	d := Disk{Name: name, Size: size, StorageProfile: storageProfile}
	return d
}

func (m *Manager) GetDisks(extraArgs ...Arguments) (disks []*Disk, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/disk"
	err = m.GetItems(path, args, &disks)
	for i := range disks {
		disks[i].manager = m
	}
	return
}

func (v *Vdc) GetDisks(extraArgs ...Arguments) (disks []*Disk, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	disks, err = v.manager.GetDisks(args)
	return
}

func (m *Manager) GetDisk(id string) (disk *Disk, err error) {
	path, _ := url.JoinPath("v1/disk", id)
	if err = m.Get(path, Defaults(), &disk); err != nil {
		log.Printf("[REQUEST-ERROR]: getting disk-%s was failed: %s]", id, errors.WithStack(err))
		return disk, err
	} else {
		disk.manager = m
		return disk, nil
	}
}

func (v *Vdc) CreateDisk(disk *Disk) error {
	args := &struct {
		Name           string   `json:"name"`
		Vdc            *string  `json:"vdc,omitempty"`
		Vm             *string  `json:"vm,omitempty"`
		Size           int      `json:"size"`
		StorageProfile string   `json:"storage_profile"`
		Tags           []string `json:"tags"`
	}{
		Name:           disk.Name,
		Vdc:            &v.ID,
		Vm:             nil,
		Size:           disk.Size,
		StorageProfile: disk.StorageProfile.ID,
		Tags:           convertTagsToNames(disk.Tags),
	}

	if disk.Vm != nil {
		args.Vm = &disk.Vm.ID
		args.Vdc = nil
	}

	if err := v.manager.Request("POST", "v1/disk", args, &disk); err != nil {
		log.Printf("[REQUEST-ERROR]: %s]", errors.WithStack(err))
		return errors.Wrapf(err, "Request for create disk-%s was failed", disk.Name)
	} else {
		disk.manager = v.manager
	}

	return nil
}

func (v *Vm) AttachDisk(disk *Disk) error {
	path := fmt.Sprintf("v1/disk/%s/attach", disk.ID)

	args := &struct {
		Vm string `json:"vm"`
	}{
		Vm: v.ID,
	}

	if err := v.manager.Request("POST", path, args, nil); err != nil {
		log.Printf("[REQUEST-ERROR]: %s]", errors.WithStack(err))
		return errors.Wrapf(err, "Request for attaching disk-%s was failed", disk.ID)
	} else {
		v.Disks = append(v.Disks, disk)
	}

	return nil
}

func (v *Vm) DetachDisk(disk *Disk) error {

	path := fmt.Sprintf("v1/disk/%s/detach", disk.ID)
	if err := v.manager.Request("POST", path, nil, nil); err != nil {
		log.Printf("[REQUEST-ERROR]: %s]", errors.WithStack(err))
		return errors.Wrapf(err, "Request for detaching disk-%s was failed", disk.ID)
	} else {
		for i, vmDisk := range v.Disks {
			if vmDisk == disk {
				v.Disks = append(v.Disks[:i], v.Disks[i+1:]...)
				break
			}
		}
	}

	return nil
}

func (d *Disk) UpdateStorageProfile(storageProfile StorageProfile) error {
	d.StorageProfile = &storageProfile
	if err := d.Update(); err != nil {
		log.Printf("[REQUEST-ERROR]: %s]", errors.WithStack(err))
		return errors.Wrapf(err, "Request for update storage-profile %s was failed", err)
	}
	return nil
}

func (d *Disk) Update() error {
	path, _ := url.JoinPath("v1/disk", d.ID)

	args := &struct {
		Name           string   `json:"name"`
		Size           int      `json:"size"`
		StorageProfile string   `json:"storage_profile"`
		Tags           []string `json:"tags"`
	}{
		Name:           d.Name,
		Size:           d.Size,
		StorageProfile: d.StorageProfile.ID,
		Tags:           convertTagsToNames(d.Tags),
	}

	err := d.manager.Request("PUT", path, args, d)
	if err != nil {
		log.Printf("[REQUEST-ERROR]: %s]", errors.WithStack(err))
		return errors.Wrapf(err, "Request for update-disk-%s was failed", d.ID)
	}

	return nil
}

func (d *Disk) Rename(name string) error {
	d.Name = name
	return d.Update()
}

func (d *Disk) Resize(size int) error {
	d.Size = size
	if err := d.Update(); err != nil {
		log.Printf("[REQUEST-ERROR]: %s]", errors.WithStack(err))
		return errors.Wrapf(err, "Request for resize disk was failed.")
	} else {
		return nil
	}
}

func (d *Disk) Delete() error {
	path, _ := url.JoinPath("v1/disk", d.ID)
	if err := d.manager.Delete(path, Defaults(), nil); err != nil {
		log.Printf("[REQUEST-ERROR]: %s]", errors.WithStack(err))
		return errors.Wrapf(err, "Request for delete-disk-%s was failed", d.ID)
	} else {
		return nil
	}
}

func (d Disk) WaitLock() error {
	path, _ := url.JoinPath("v1/disk", d.ID)
	if err := loopWaitLock(d.manager, path); err != nil {
		log.Printf("[REQUEST-ERROR]: %s]", errors.WithStack(err))
		return errors.Wrapf(err, "Request for WaitLock disk was failed")
	} else {
		return nil
	}
}
