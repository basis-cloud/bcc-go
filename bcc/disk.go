package bcc

import (
	"fmt"
	"log"
	"net/url"
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
	path := "v1/disk"
	args := Defaults()
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &disks); err != nil {
		log.Printf("[REQUEST-ERROR] get-disks was failed: %s", err)
	} else {
		for i := range disks {
			disks[i].manager = m
		}
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
		log.Printf("[REQUEST-ERROR]: getting disk with id='%s' was failed: %s]", id, err)
	} else {
		disk.manager = m
	}

	return
}

func (v *Vdc) CreateDisk(disk *Disk) (err error) {
	path := "v1/disk"
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

	if err = v.manager.Request("POST", path, args, &disk); err != nil {
		log.Printf("[REQUEST-ERROR] disk create was failed: %s", err)
	} else {
		disk.manager = v.manager
	}

	return
}

func (v *Vm) AttachDisk(disk *Disk) (err error) {
	path := fmt.Sprintf("v1/disk/%s/attach", disk.ID)

	args := &struct {
		Vm string `json:"vm"`
	}{
		Vm: v.ID,
	}

	if err = v.manager.Request("POST", path, args, nil); err != nil {
		log.Printf("[REQUEST-ERROR] disk attach with id ='%s' was failed : %s", disk.ID, err)
	} else {
		v.Disks = append(v.Disks, disk)
	}

	return
}

func (v *Vm) DetachDisk(disk *Disk) (err error) {
	path := fmt.Sprintf("v1/disk/%s/detach", disk.ID)

	if err = v.manager.Request("POST", path, nil, nil); err != nil {
		log.Printf("[REQUEST-ERROR] disk detach with id='%s' was failed: %s", disk.ID, err)
	} else {
		for i, vmDisk := range v.Disks {
			if vmDisk == disk {
				v.Disks = append(v.Disks[:i], v.Disks[i+1:]...)
				break
			}
		}
	}

	return
}

func (d *Disk) UpdateStorageProfile(storageProfile StorageProfile) (err error) {
	d.StorageProfile = &storageProfile

	if err = d.Update(); err != nil {
		log.Printf("[REQUEST-ERROR]: storage-profile update was failed %s", err)
	}

	return nil
}

func (d *Disk) Update() (err error) {
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

	if err = d.manager.Request("PUT", path, args, d); err != nil {
		log.Printf("[REQUEST-ERROR] disk update with id='%s' was failed: %s", d.ID, err)
	}

	return nil
}

func (d *Disk) Rename(name string) error {
	d.Name = name
	return d.Update()
}

func (d *Disk) Resize(size int) (err error) {
	d.Size = size

	if err = d.Update(); err != nil {
		log.Printf("[REQUEST-ERROR] disk-resize with id='%s' was failed: %s", d.ID, err)
	}

	return
}

func (d *Disk) Delete() (err error) {
	path, _ := url.JoinPath("v1/disk", d.ID)

	if err = d.manager.Delete(path, Defaults(), nil); err != nil {
		log.Printf("[REQUEST-ERROR] disk-delete with id='%s' was failed: %s", d.ID, err)
	}

	return
}

func (d Disk) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/disk", d.ID)

	if err = loopWaitLock(d.manager, path); err != nil {
		log.Printf("[REQUEST-ERROR] disk waitlock with id='%s' was failed: %s", d.ID, err)
	}

	return
}
