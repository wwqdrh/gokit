package dns

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getAllDomainSuffixes(t *testing.T) {
	suffixes := getAllDomainSuffixes(map[string]string{
		"abc.com":   "",
		"a.b.c.net": "",
		"c.b.a.com": "",
		"xyz.net":   "",
	})
	require.True(t, ArrayEquals([]string{"com", "net"}, suffixes))
}
