package zeroconf

import (
	"log"
	"net"
	"strings"

	"errors"

	"github.com/miekg/dns"
)

var edgeServiceType string

var EdgeExportedServiceEntry chan *ServiceEntry

func EdgeGetSubscriberChan() (chan *ServiceEntry, error) {
	if EdgeExportedServiceEntry != nil {
		return nil, errors.New("Subscribe Chan is Made Before")
	}
	EdgeExportedServiceEntry = make(chan *ServiceEntry, 32)
	return EdgeExportedServiceEntry, nil
}

//EdgeRegisterProxy for Edge
func EdgeRegisterProxy(instance, service, domain string, port int,
	host string, ips []string, text []string,
	ifaces []net.Interface) (*Server, *ServiceEntry, error) {

	edgeServiceType = service

	server, err := RegisterProxy(instance, service, domain, port, host, ips, text, ifaces)
	if err != nil {
		return nil, nil, err
	}
	err = server.EdgeAdvertise()
	if err != nil {
		return nil, nil, err
	}
	return server, server.service, err
}

//EdgeGetText returns server's text field
func (s *Server) EdgeGetText() []string {
	return s.service.Text
}

//EdgeResetServer react to interface change.
//should be called when interface changed.
func (s *Server) EdgeResetServer(newipv4s []net.IP) {
	s.service.AddrIPv4 = newipv4s
	select {
	case EdgeExportedServiceEntry <- nil:
	default:
		log.Println("send Chan Full")
	}
	s.EdgeAdvertise()
	return
}

// EdgeHandleQuery is used to handle an incoming query
func (s *Server) edgeHandleQuery(msg *dns.Msg, ifIndex int, from net.Addr) error {
	//IsFromIPv4?
	_, err := s.edgeParseIPv4(from)
	if err != nil {
		return err
	}
	//Parse Txt and A Record
	entry, err := s.edgeParseServiceEntry(msg)
	if err != nil {
		return err
	}

	select {
	case EdgeExportedServiceEntry <- entry:
	default:
		log.Println("send Chan Full")
	}

	if msg.Question != nil {
		err = s.edgeSendUnicastResponse(msg, ifIndex, from)
	}
	return err
}

//EdgeParseIPv4 Palse ipv4 from net.Addr
func (Server) edgeParseIPv4(from net.Addr) (string, error) {
	deviceIPPORT := strings.Split(from.String(), ":")
	srcIP := deviceIPPORT[0]

	isV4 := net.ParseIP(srcIP)
	if isV4.To4() == nil {
		return "", errors.New("Do Not Handle IPv6")
	}
	return srcIP, nil
}

//EdgeParseServiceEntry parse ServiceEntry from dns msg
//filter messages by serviceType "_orchestration._tcp"
func (Server) edgeParseServiceEntry(msg *dns.Msg) (*ServiceEntry, error) {
	sections := append(msg.Answer, msg.Ns...)
	sections = append(sections, msg.Extra...)

	entry := newEntryFromAnswer(sections)
	if entry == nil {
		return nil, errors.New("NO TXT && NO SRV")
	}

	appendEntryIP(entry, sections)

	return entry, nil
}

func newEntryFromAnswer(sections []dns.RR) *ServiceEntry {
	var entry *ServiceEntry
	params := defaultParams(edgeServiceType)
	for _, answer := range sections {
		switch rr := answer.(type) {
		case *dns.TXT:
			if !serviceNameChecker(params, rr.Hdr.Name) {
				continue
			}
			if entry == nil {
				entry = NewServiceEntry(trimDot(strings.Replace(rr.Hdr.Name,
					params.ServiceName(), "", 1)), params.Service, params.Domain)
			}
			entry.Text = rr.Txt
			entry.TTL = rr.Hdr.Ttl
		case *dns.SRV:
			//SRV To Get HostName
			if !serviceNameChecker(params, rr.Hdr.Name) {
				continue
			}
			if entry == nil {
				entry = NewServiceEntry(trimDot(strings.Replace(rr.Hdr.Name,
					params.ServiceName(), "", 1)), params.Service, params.Domain)
			}
			entry.HostName = rr.Target
			entry.Port = int(rr.Port)
			entry.TTL = rr.Hdr.Ttl
		default:
			continue
		}
	}
	return entry
}

func serviceNameChecker(params *LookupParams, rrHdrName string) bool {
	if params.ServiceInstanceName() != "" && params.ServiceInstanceName() != rrHdrName {
		return false
	} else if !strings.HasSuffix(rrHdrName, params.ServiceName()) {
		return false
	}
	return true
}

func appendEntryIP(entry *ServiceEntry, sections []dns.RR) {
	for _, answer := range sections {
		switch rr := answer.(type) {
		case *dns.A:
			if entry.HostName == rr.Hdr.Name {
				entry.AddrIPv4 = append(entry.AddrIPv4, rr.A)
			}
		case *dns.AAAA:
			if entry.HostName == rr.Hdr.Name {
				entry.AddrIPv6 = append(entry.AddrIPv6, rr.AAAA)
			}
		}
	}
}

//EdgeSendResponse send resp if it is question
//ToDo : is all msg has from?
func (s *Server) edgeSendUnicastResponse(msg *dns.Msg, ifIndex int, from net.Addr) error {
	resp := dns.Msg{}
	resp.SetReply(msg)
	resp.Compress = true
	resp.RecursionDesired = false
	resp.Authoritative = true
	resp.Question = nil // RFC6762 section 6 "responses MUST NOT contain any questions"
	s.setExtraField(&resp)
	err := s.unicastResponse(&resp, ifIndex, from)
	return err
}

//EdgeAdvertise Send TXT for serviceinfo, SRV for Hostname, A/AAAA for IPs
//and request unicast response.
func (s *Server) EdgeAdvertise() error {
	var resp dns.Msg
	resp.Question = []dns.Question{
		dns.Question{Name: s.service.ServiceInstanceName(), Qtype: dns.TypeTXT, Qclass: dns.ClassINET},
	}
	s.setExtraField(&resp)
	err := s.multicastResponse(&resp, 0)
	return err
}

func (s *Server) setExtraField(resp *dns.Msg) {
	txt := &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   s.service.ServiceInstanceName(),
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    s.ttl,
		},
		Txt: s.service.Text,
	}
	srv := &dns.SRV{
		Hdr: dns.RR_Header{
			Name:   s.service.ServiceInstanceName(),
			Rrtype: dns.TypeSRV,
			Class:  dns.ClassINET | qClassCacheFlush,
			Ttl:    s.ttl,
		},
		Priority: 0,
		Weight:   0,
		Port:     uint16(s.service.Port),
		Target:   s.service.HostName,
	}
	resp.Extra = append(resp.Extra, txt, srv)
	resp.Extra = s.appendAddrs(resp.Extra, s.ttl, 0, false)
}
