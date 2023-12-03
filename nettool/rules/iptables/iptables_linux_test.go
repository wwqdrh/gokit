package iptables

import "testing"

func TestRuleAddDelete(t *testing.T) {
	r := NewIptablesRule()
	if err := r.NatAddPortRedirect("127.0.0.1/32", "4533", "14553"); err != nil {
		t.Error(err)
	}
	r.NatDelPortRedirect("127.0.0.1/32", "4533", "14553")
}

func TestRuleNumber(t *testing.T) {
	r := NewIptablesRule()
	r.ListRuleNumber("nat")
}
