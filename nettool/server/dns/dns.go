package dns

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/wwqdrh/gokit/logger"
	"go.uber.org/zap"
	"golang.org/x/net/dns/dnsmessage"
)

type IDns interface {
	Server(context.Context) error
}

type dnsServer struct {
	a          arecord
	primaryDNS string
	rtr        rtrrecord
	port       int
}

type (
	arecord   func(string) [4]byte
	rtrrecord func(string) string
)

func NewDnsServer(a arecord, rtr rtrrecord, port int) IDns {
	primary := GetNameServer() + ":53"
	logger.DefaultLogger.Info("a primarydns: " + primary)
	return &dnsServer{
		a:          a,
		rtr:        rtr,
		port:       port,
		primaryDNS: primary,
	}
}

func (d *dnsServer) Server(ctx context.Context) error {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: d.port})
	if err != nil {
		return err
	}
	defer conn.Close()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			buf := make([]byte, 512)
			_, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				logger.DefaultLogger.Error(err.Error())
				continue
			}
			var msg dnsmessage.Message
			if err := msg.Unpack(buf); err != nil {
				logger.DefaultLogger.Error(err.Error())
				continue
			}
			go d.handle(addr, conn, msg)
		}
	}
}

func (d *dnsServer) handle(addr *net.UDPAddr, conn *net.UDPConn, msg dnsmessage.Message) {
	// query info
	if len(msg.Questions) < 1 {
		return
	}
	question := msg.Questions[0]
	var (
		queryTypeStr = question.Type.String()
		queryNameStr = question.Name.String()
		queryType    = question.Type
	)
	logger.DefaultLogger.Infox("[%s] queryName: [%s]\n", nil, queryTypeStr, queryNameStr)

	// find record
	var (
		resource dnsmessage.Resource
		err      error
	)
	switch queryType {
	case dnsmessage.TypeA:
		resource, err = d.getARecord(queryNameStr)
	case dnsmessage.TypePTR:
		resource, err = d.getRTRRecord(queryNameStr)
	default:
		logger.DefaultLogger.Infox("not support dns queryType: [%s] \n", nil, queryTypeStr)
		return
	}

	if err != nil {
		logger.DefaultLogger.Warn(err.Error())
		d.response(addr, conn, msg)
	}

	// send response
	msg.Response = true
	msg.Answers = append(msg.Answers, resource)
	d.response(addr, conn, msg)
}

// Response return
func (d *dnsServer) response(addr *net.UDPAddr, conn *net.UDPConn, msg dnsmessage.Message) {
	packed, err := msg.Pack()
	if err != nil {
		logger.DefaultLogger.Error(err.Error())
		return
	}
	if _, err := conn.WriteToUDP(packed, addr); err != nil {
		logger.DefaultLogger.Error(err.Error())
	}
}

func (d *dnsServer) getARecord(queryNameStr string) (dnsmessage.Resource, error) {
	queryName, _ := dnsmessage.NewName(queryNameStr)
	rst := d.a(queryNameStr)
	if rst == [4]byte{} {
		var err error
		rst, err = d.nslookup(queryNameStr, dnsmessage.TypeA)
		if err != nil {
			logger.DefaultLogger.Warnx("not fount PTR record queryName: [%s] \n", nil, queryNameStr)
			return dnsmessage.Resource{}, errors.New("not found")
		}
	}

	return dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  queryName,
			Class: dnsmessage.ClassINET,
			TTL:   600,
		},
		Body: &dnsmessage.AResource{
			A: rst,
		},
	}, nil
}

func (d *dnsServer) nslookup(domain string, t dnsmessage.Type) ([4]byte, error) {
	m, err := NsLookup(domain, uint16(t), "udp", d.primaryDNS)
	if err != nil {
		return [4]byte{}, err
	}

	if len(m.Answer) > 0 {
		for _, item := range m.Answer {
			logger.DefaultLogger.Info(item.String(), zap.String("key", "forward"))
		}
		c := m.Answer[0].String()
		return string2IP(strings.Split(c, "\t")[4]), nil
	} else {
		return [4]byte{}, errors.New("not found")
	}
}

func (d *dnsServer) getRTRRecord(queryNameStr string) (dnsmessage.Resource, error) {
	queryName, _ := dnsmessage.NewName(queryNameStr)
	rst := d.rtr(queryName.String())
	if rst == "" {
		logger.DefaultLogger.Warnx("not fount PTR record queryName: [%s] \n", nil, queryNameStr)
		return dnsmessage.Resource{}, errors.New("not found")
	}

	name, _ := dnsmessage.NewName(rst)

	return dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  queryName,
			Class: dnsmessage.ClassINET,
		},
		Body: &dnsmessage.PTRResource{
			PTR: name,
		},
	}, nil
}
