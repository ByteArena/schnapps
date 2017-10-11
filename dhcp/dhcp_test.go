package dhcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDHCPServer(t *testing.T) {
	assert := assert.New(t)

	cidr := "192.168.0.1/24"

	s, err := NewDHCPServer(cidr)
	assert.NoError(err)

	ip, popErr := s.Pop()
	assert.NoError(popErr)

	assert.Equal("192.168.0.1", ip)

	_, pop2Err := s.Pop()
	assert.NotNil(pop2Err)

	// Release and retry
	s.Release(ip)

	ip2, pop3Err := s.Pop()
	assert.NoError(pop3Err)

	assert.Equal("192.168.0.1", ip2)
}
