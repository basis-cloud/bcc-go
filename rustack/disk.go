package rustack

import (
	"fmt"
	"net/url"
)

type Disk struct {
	manager        *Manager
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Scsi           string          `json:"scsi"`
	Size           int             `json:"size"`
	Vm             *Vm             `json:"vm"`
	StorageProfile *StorageProfile `json:"storage_profile"`
	Locked         bool            `json:"locked,omitempty"`
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
	err = m.Get(path, Defaults(), &disk)
	if err != nil {
		return
	}
	disk.manager = m
	return
}

func (v *Vm) AttachDisk(disk *Disk) error {
	path := fmt.Sprintf("v1/disk/%s/attach", disk.ID)

	args := &struct {
		Vm string `json:"vm"`
	}{
		Vm: v.ID,
	}

	err := v.manager.Request("POST", path, args, nil)
	if err != nil {
		return err
	}
	v.Disks = append(v.Disks, disk)

	return nil
}

func (v *Vm) DetachDisk(disk *Disk) error {

	path := fmt.Sprintf("v1/disk/%s/detach", disk.ID)
	err := v.manager.Request("POST", path, nil, nil)
	if err != nil {
		return err
	}
	for i, vmDisk := range v.Disks {
		if vmDisk == disk {
			v.Disks = append(v.Disks[:i], v.Disks[i+1:]...)
			break
		}
	}

	return nil
}

func (d *Disk) Update(name string, size int, storageProfileId string) error {
	path, _ := url.JoinPath("v1/disk", d.ID)

	args := &struct {
		Name           string `json:"name"`
		Size           int    `json:"size"`
		StorageProfile string `json:"storage_profile"`
	}{
		Name:           name,
		Size:           size,
		StorageProfile: storageProfileId,
	}

	err := d.manager.Request("PUT", path, args, d)
	if err != nil {
		return err
	}

	return nil
}

func (d *Disk) Rename(name string) error {
	return d.Update(name, d.Size, d.StorageProfile.ID)
}

func (d *Disk) Resize(size int) error {
	return d.Update(d.Name, size, d.StorageProfile.ID)
}

func (d *Disk) UpdateStorageProfile(storageProfile StorageProfile) error {
	return d.Update(d.Name, d.Size, storageProfile.ID)
}

func (d *Disk) Delete() error {
	path, _ := url.JoinPath("v1/disk", d.ID)
	return d.manager.Delete(path, Defaults(), nil)
}

func (d Disk) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/disk", d.ID)
	return loopWaitLock(d.manager, path)
}
