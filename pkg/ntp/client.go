package ntp

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type Client struct {
	conn       *net.UDPConn
	remoteAddr *net.UDPAddr
}

const (
	network  = "udp4"
	respSize = 48
)

var (
	ErrKissOfDeath     = errors.New("received kiss of death")
	ErrEmptyOriginTime = errors.New("origin time is empty")
	ErrEmptyRefTime    = errors.New("reference time is empty")
	ErrEmptyRecvTime   = errors.New("receive time is empty")
)

func NewClient(remote *net.UDPAddr) (*Client, error) {
	conn, err := net.DialUDP(network, nil, remote)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn, remoteAddr: remote}, nil
}

func (c *Client) Exchange() (*Result, error) {
	sendHeader := initPacket()
	sendTime := time.Now()
	resp, err := c.Send(sendHeader.toBytes())
	if err != nil {
		return nil, err
	}

	recvTime := time.Now()
	recvHeader := packetFromBytes(resp)

	if recvHeader.Stratum == 0 {
		return nil, ErrKissOfDeath
	}

	if recvHeader.RefTS == ntpTime(0) {
		return nil, ErrEmptyRefTime
	}

	if recvHeader.OriginTS == ntpTime(0) {
		return nil, ErrEmptyOriginTime
	}

	if recvHeader.ReceiveTS == ntpTime(0) {
		return nil, ErrEmptyRecvTime
	}

	recvHeader.OriginTS = toNtpTime(sendTime)

	res := resultFromHeader(recvHeader, toNtpTime(recvTime))

	return res, nil
}

func (c *Client) Send(payload []byte) ([]byte, error) {
	count, err := c.conn.Write(payload)
	if err != nil {
		return nil, err
	}

	resp := make([]byte, 512)
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	count, err = c.conn.Read(resp)
	if err != nil {
		return nil, err
	}

	return resp[:count], nil
}

type Result struct {
	ServerTime     time.Time
	ClockOffset    time.Duration
	RTDelay        time.Duration
	Precision      time.Duration
	Stratum        uint8
	RefID          uint32
	RefTime        time.Time
	RootDelay      time.Duration
	RootDispersion time.Duration
	RootDistance   time.Duration
	LI             uint8
	MinError       time.Duration
	KissCode       string
	Poll           time.Duration
}

// this is supposed to be used on the server response header
func resultFromHeader(h *packet, clientRecvTime ntpTime) *Result {
	res := &Result{
		ServerTime:     h.TransmitTS.Time(),
		ClockOffset:    h.offset(clientRecvTime.Time()),
		RTDelay:        h.rtDelay(clientRecvTime.Time()),
		Precision:      exponentInterval(h.Precision),
		Stratum:        h.Stratum,
		RefID:          h.RefID,
		RefTime:        h.RefTS.Time(),
		RootDelay:      h.RootDelay.Duration(),
		RootDispersion: h.RootDispersion.Duration(),
		LI:             h.getLI(),
		MinError:       h.minError(clientRecvTime),
		KissCode:       "", // TODO: add KISS code parsing
		Poll:           exponentInterval(h.Poll),
	}

	res.calcRootDistance()

	return res
}

/*
It's a simplified formula that works for a 2-packet client-server exchange
A server implementation would require many more variables to be included
*/
func (r *Result) calcRootDistance() {
	totalDelay := r.RTDelay + r.RootDelay
	r.RootDistance = totalDelay/2 + r.RootDispersion
}

func SetupRemoteAddr(host string, port int) (*net.UDPAddr, error) {
	if port <= 0 {
		return nil, fmt.Errorf("port value must be greater than zero")
	}

	addrs, err := net.LookupHost(host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("no IP addresses found for %s", host)
	}

	ips := filterIPV4(addrs)
	if len(ips) == 0 {
		return nil, fmt.Errorf("no IPv4 addresses found during the lookup of %s", host)
	}

	return &net.UDPAddr{IP: ips[0], Port: port}, nil
}

func filterIPV4(addrs []string) []net.IP {
	res := []net.IP{}
	for _, addr := range addrs {
		parts := strings.Split(addr, ":")

		if ip := net.ParseIP(parts[0]); ip != nil && ip.To4() != nil {
			res = append(res, ip)
		}
	}

	return res
}
