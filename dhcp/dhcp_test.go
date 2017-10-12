package dhcp

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newInt(x int) *int {
	return &x
}

func TestDHCPServer(t *testing.T) {
	type Result struct {
		IP    string
		Error error
	}

	examples := []struct {
		Name          string
		Network       string
		InitError     error
		Results       []Result
		UsedIPs       []string
		StartIndex    *int
		FreeIpAtIndex map[int]string
	}{
		{
			Name:      "it should send an error cidr is not correctly formatted",
			Network:   "abcd",
			InitError: errors.New("invalid CIDR address: abcd"),
		},
		{
			Name:    "it should generate an ip",
			Network: "192.168.0.1/24",
			Results: []Result{{IP: "192.168.0.2"}},
		},
		{
			Name:    "it should generate an ip starting from the CIDR ip",
			Network: "192.168.0.100/24",
			Results: []Result{{IP: "192.168.0.101"}},
		},
		{
			Name:    "it should send an error when there is no IP available",
			Network: "10.0.0.0/30",
			Results: []Result{{IP: "10.0.0.1"}, {IP: "10.0.0.2"}, {Error: errors.New("No ip left")}},
		},
		{
			Name:    "it should skip an ip if this ip is used",
			Network: "10.0.0.0/24",
			UsedIPs: []string{"10.0.0.2"},
			Results: []Result{{IP: "10.0.0.1"}, {IP: "10.0.0.3"}},
		},
		{
			Name:    "it should send an error when all ips are in use",
			Network: "10.0.0.0/30",
			UsedIPs: []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"},
			Results: []Result{{Error: errors.New("No ip left")}},
		},
		{
			Name:      "it should send an error when the network is too small",
			Network:   "10.0.0.0/31",
			InitError: errors.New("Subnet is too small, CIDR should be at least /30"),
		},
		{
			Name:       "it should loop back when the network is missing ips",
			Network:    "10.0.0.0/24",
			StartIndex: newInt(253),
			Results:    []Result{{IP: "10.0.0.254"}, {IP: "10.0.0.1"}},
		},
		{
			Name:       "it should continue to the next byte when the first one is full",
			Network:    "10.0.0.0/16",
			StartIndex: newInt(254),
			Results:    []Result{{IP: "10.0.0.255"}, {IP: "10.0.1.0"}, {IP: "10.0.1.1"}},
		},
		{
			Name:          "it should re-use released ips",
			Network:       "10.0.0.0/30",
			FreeIpAtIndex: map[int]string{2: "10.0.0.2"},
			Results:       []Result{{IP: "10.0.0.1"}, {IP: "10.0.0.2"}, {Error: errors.New("No ip left")}, {IP: "10.0.0.2"}},
		},
	}

	for _, example := range examples {
		t.Run(example.Name, func(t *testing.T) {
			// Init server
			server, err := NewDHCPServer(example.Network)
			if example.InitError != nil {
				require.NotNil(t, err)
				assert.Equal(t, example.InitError.Error(), err.Error())
				return
			} else {
				require.Nil(t, err)
			}

			// Init Used
			for _, ip := range example.UsedIPs {
				server.Used[ip] = true
			}

			if example.StartIndex != nil {
				// Set start index
				server.Current = *example.StartIndex
			}

			// Test results
			for i := 0; i < len(example.Results); i++ {
				ip, err := server.Pop()
				if example.Results[i].Error != nil {
					require.NotNil(t, err)
					require.Equal(t, example.Results[i].Error.Error(), err.Error())
				} else {
					require.Nil(t, err)
				}

				assert.Equal(t, example.Results[i].IP, ip)

				if example.FreeIpAtIndex[i] != "" {
					server.Release(example.FreeIpAtIndex[i])
				}
			}
		})
	}
}
