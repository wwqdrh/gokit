package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToIpAndMask(t *testing.T) {
	ip, mask, _ := IpAndMask("10.95.134.192/29")
	require.Equal(t, "10.95.134.192", ip)
	require.Equal(t, "255.255.255.248", mask)
}

func TestIpNetPart(t *testing.T) {
	netpart, _ := IpNetPart("10.95.134.193/29")
	require.Equal(t, "10.95.134.192/29", netpart)
}
