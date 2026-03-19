package bcc

import (
	"fmt"
	"log"
)

type DnsRecord struct {
	manager  *Manager
	DnsZone  string
	ID       string `json:"id"`
	Data     string `json:"data"`
	Flag     int    `json:"flag"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Priority int    `json:"priority"`
	Tag      string `json:"tag"`
	Ttl      int    `json:"ttl"`
	Type     string `json:"type"`
	Weight   int    `json:"weight"`
}

func NewDnsRecord(data string, flag int, host string, port int, priority int, tag string, ttl int, dns_type string, weight int) DnsRecord {
	d := DnsRecord{Data: data, Flag: flag, Host: host, Port: port, Priority: priority, Tag: tag, Ttl: ttl, Type: dns_type, Weight: weight}
	return d
}

func (m *Manager) GetDnsRecords(dnsId string, extraArgs ...Arguments) (dnsRecord []*DnsRecord, err error) {
	path := fmt.Sprintf("v1/dns/%s/dns_record", dnsId)
	args := Defaults()
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &dnsRecord); err != nil {
		log.Printf("[REQUEST-ERROR] get-dnsRecord's for dns with id='%s' was failed: %s", dnsId, err)
	} else {
		for i := range dnsRecord {
			dnsRecord[i].manager = m
		}
	}

	return
}

func (d *Dns) GetDnsRecords(extraArgs ...Arguments) (dnsRecord []*DnsRecord, err error) {
	dnsRecord, err = d.manager.GetDnsRecords(d.ID, extraArgs...)
	return
}

func (d *Dns) CreateDnsRecord(dnsRecord *DnsRecord) (err error) {
	path := fmt.Sprintf("v1/dns/%s/record", d.ID)
	args := &struct {
		manager  *Manager
		ID       string  `json:"id"`
		Data     string  `json:"data"`
		Flag     int     `json:"flag"`
		Host     string  `json:"host"`
		Port     *int    `json:"port"`
		Priority *int    `json:"priority"`
		Tag      *string `json:"tag"`
		Ttl      int     `json:"ttl"`
		Type     string  `json:"type"`
		Weight   *int    `json:"weight"`
	}{
		ID:       dnsRecord.ID,
		Data:     dnsRecord.Data,
		Host:     dnsRecord.Host,
		Ttl:      dnsRecord.Ttl,
		Type:     dnsRecord.Type,
		Weight:   nil,
		Flag:     0,
		Tag:      nil,
		Priority: nil,
		Port:     nil,
	}

	if dnsRecord.Type == "CAA" {
		args.Tag = &dnsRecord.Tag
		args.Flag = dnsRecord.Flag
	} else if dnsRecord.Type == "MX" {
		args.Priority = &dnsRecord.Priority
	} else if dnsRecord.Type == "SRV" {
		args.Priority = &dnsRecord.Priority
		args.Weight = &dnsRecord.Weight
		args.Port = &dnsRecord.Port
	}

	if err = d.manager.Request("POST", path, args, &dnsRecord); err != nil {
		log.Printf("[REQUEST-ERROR] create-dnsRecord's was failed: %s", err)
	} else {
		dnsRecord.manager = d.manager
		dnsRecord.DnsZone = d.ID
	}

	return
}

func (d *Dns) GetDnsRecord(id string) (dnsRecord *DnsRecord, err error) {
	path := fmt.Sprintf("v1/dns/%s/record/%s", d.ID, id)

	if err = d.manager.Get(path, Defaults(), &dnsRecord); err != nil {
		log.Printf("[REQUEST-ERROR] get-dnsRecord with id='%s' was failed: %s", id, err)
	} else {
		dnsRecord.manager = d.manager
		dnsRecord.DnsZone = d.ID
	}

	return
}

func (d *DnsRecord) Update() (err error) {
	path := fmt.Sprintf("v1/dns/%s/record/%s", d.DnsZone, d.ID)
	args := &struct {
		Data     string  `json:"data"`
		Flag     int     `json:"flag"`
		Host     string  `json:"host"`
		Port     *int    `json:"port"`
		Priority *int    `json:"priority"`
		Tag      *string `json:"tag"`
		Ttl      int     `json:"ttl"`
		Type     string  `json:"type"`
		Weight   *int    `json:"weight"`
	}{
		Data:     d.Data,
		Host:     d.Host,
		Ttl:      d.Ttl,
		Type:     d.Type,
		Weight:   nil,
		Flag:     0,
		Tag:      nil,
		Priority: nil,
		Port:     nil,
	}

	if d.Type == "CAA" {
		args.Tag = &d.Tag
		args.Flag = d.Flag
	} else if d.Type == "MX" {
		args.Priority = &d.Priority
	} else if d.Type == "SRV" {
		args.Priority = &d.Priority
		args.Weight = &d.Weight
		args.Port = &d.Port
	}

	if err = d.manager.Request("PUT", path, args, d); err != nil {
		log.Printf("[REQUEST-ERROR] update-dnsRecord's was failed: %s", err)
	}

	return
}

func (d *DnsRecord) Delete() error {
	path := fmt.Sprintf("v1/dns/%s/record/%s", d.DnsZone, d.ID)
	return d.manager.Delete(path, Defaults(), nil)
}
