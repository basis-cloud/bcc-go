package bcc

import (
	"fmt"
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

func (m *Manager) GetDnsRecords(dns_id string, extraArgs ...Arguments) (dns_records []*DnsRecord, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := fmt.Sprintf("v1/dns/%s/dns_record", dns_id)
	err = m.GetItems(path, args, &dns_records)
	for i := range dns_records {
		dns_records[i].manager = m
	}
	return
}

func (d *Dns) GetDnsRecords(extraArgs ...Arguments) (dns_record []*DnsRecord, err error) {
	dns_record, err = d.manager.GetDnsRecords(d.ID, extraArgs...)
	if err != nil {
		return nil, err
	}
	return dns_record, nil
}

func (d *Dns) CreateDnsRecord(dnsRecord *DnsRecord) (err error) {
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

	path := fmt.Sprintf("v1/dns/%s/record", d.ID)
	err = d.manager.Request("POST", path, args, &dnsRecord)
	if err != nil {
		return err
	}
	dnsRecord.manager = d.manager
	dnsRecord.DnsZone = d.ID
	return
}

func (d *Dns) GetDnsRecord(id string) (dns_record *DnsRecord, err error) {
	path := fmt.Sprintf("v1/dns/%s/record/%s", d.ID, id)
	err = d.manager.Get(path, Defaults(), &dns_record)
	if err != nil {
		return
	}
	dns_record.manager = d.manager
	dns_record.DnsZone = d.ID
	return
}

func (d *DnsRecord) Update() error {
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

	path := fmt.Sprintf("v1/dns/%s/record/%s", d.DnsZone, d.ID)
	err := d.manager.Request("PUT", path, args, d)
	if err != nil {
		return err
	}
	return nil
}

func (d *DnsRecord) Delete() error {
	path := fmt.Sprintf("v1/dns/%s/record/%s", d.DnsZone, d.ID)
	return d.manager.Delete(path, Defaults(), nil)
}
