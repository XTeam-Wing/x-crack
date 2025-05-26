package telnet

import (
	"regexp"
	"testing"
)

// TestTelnetBrute tests the Telnet brute force functionality
func TestTelnet(t *testing.T) {
	line := "root@USR-G806:~$"
	if regexp.MustCompile(`[#$]\s*$`).MatchString(line) {
		t.Log("Telnet prompt detected")
	}
	if regexp.MustCompile(`^\w+@[\w-]+:\w+[$#]\s*$`).MatchString(line) {
		t.Log("Telnet prompt with username and hostname detected")
	} else {
		t.Error("Telnet prompt not detected")
	}
}
