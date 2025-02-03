package bcc

import (
	"fmt"
)

type SubnetDNSServer struct {
	DNSServer string `json:"dns_server"`
}

type SubnetRoute struct {
	CIDR    string `json:"cidr"`
	Gateway string `json:"gateway"`
	Metric  int    `json:"metric"`
}

type Subnet struct {
	manager *Manager
	ID      string `json:"id"`
	CIDR    string `json:"cidr"`
	Gateway string `json:"gateway"`
	StartIp string `json:"start_ip"`
	EndIp   string `json:"end_ip"`
	IsDHCP  bool   `json:"enable_dhcp"`
	Locked  bool   `json:"locked"`

	DnsServers   []*SubnetDNSServer `json:"dns_servers"`
	SubnetRoutes []*SubnetRoute     `json:"subnet_routes"`

	network *Network
}

func NewSubnet(cidr string, gateway string, startIp string, endIp string, isDHCP bool) Subnet {
	s := Subnet{CIDR: cidr, Gateway: gateway, StartIp: startIp, EndIp: endIp, IsDHCP: isDHCP}

	s.DnsServers = make([]*SubnetDNSServer, 0)
	s.SubnetRoutes = make([]*SubnetRoute, 0)

	return s
}

func NewSubnetDNSServer(dnsServer string) SubnetDNSServer {
	s := SubnetDNSServer{DNSServer: dnsServer}
	return s
}

func NewSubnetRoute(cidr string, gateway string, metric int) SubnetRoute {
	s := SubnetRoute{CIDR: cidr, Gateway: gateway, Metric: metric}
	return s
}

func (s *Subnet) Delete() error {
	path := fmt.Sprintf("v1/network/%s/subnet/%s", s.network.ID, s.ID)
	return s.manager.Delete(path, Defaults(), nil)
}

func (s *Subnet) update() error {
	path := fmt.Sprintf("v1/network/%s/subnet/%s", s.network.ID, s.ID)
	return s.manager.Request("PUT", path, s, s)
}

func (s *Subnet) EnableDHCP() error {
	s.IsDHCP = true
	return s.update()
}

func (s *Subnet) DisableDHCP() error {
	s.IsDHCP = false
	return s.update()
}

func (s *Subnet) UpdateDNSServers(dnsServers []*SubnetDNSServer) error {
	s.DnsServers = dnsServers
	return s.update()
}

func (s *Subnet) UpdateRoutes(routes []*SubnetRoute) error {
	s.SubnetRoutes = routes
	return s.update()
}

func (s Subnet) WaitLock() (err error) {
	path := fmt.Sprintf("v1/network/%s/subnet/%s", s.network.ID, s.ID)
	return loopWaitLock(s.manager, path)
}
