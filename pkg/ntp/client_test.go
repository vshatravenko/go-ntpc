package ntp

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testRemoteHost = "time.google.com"
	testRemotePort = 123
)

func TestExchange(t *testing.T) {
	remoteAddr, err := setupTestRemoteAddr(testRemoteHost)
	assert.NoError(t, err)
	client, err := NewClient(remoteAddr)
	assert.NoError(t, err)

	_, err = client.Exchange()
	assert.NoError(t, err)
}

func setupTestRemoteAddr(host string) (*net.UDPAddr, error) {
	addrs, err := net.LookupHost(host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("no IP addresses found for %s", testRemoteHost)
	}

	ips := filterIPV4(addrs)
	if len(ips) == 0 {
		return nil, fmt.Errorf("no IPv4 addresses found during the lookup of %s", testRemoteHost)
	}

	return &net.UDPAddr{IP: ips[0], Port: testRemotePort}, nil
}

func setupTestRemoteSingleAddr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   net.IPv4(192, 95, 27, 155),
		Port: testRemotePort,
	}
}

func setupTestLocalAddr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 1123,
	}
}
