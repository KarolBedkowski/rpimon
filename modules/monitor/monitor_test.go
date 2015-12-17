package monitor

import (
	"testing"
	//	"time"
)

func TestParseIpResult1if(t *testing.T) {
	data1 := `1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    link/ether 11:22:33:44:55:66 brd ff:ff:ff:ff:ff:ff
    inet6 fe80::1234:aff:fecd:490e/64 scope link
       valid_lft forever preferred_lft forever`

	ifaces := parseIPResult(data1)
	if ifaces == nil || len(ifaces) == 0 {
		t.Errorf("Missing data %#v", ifaces)
	}

	if len(ifaces) != 1 {
		t.Errorf("Wrong interfaces number %#v", ifaces)
	}

	iface := ifaces[0]
	if iface.Name != "eth0" {
		t.Errorf("Wrong interface name %#v", ifaces)
	}
	if iface.Address != "" {
		t.Errorf("Wrong interface Address %#v", ifaces)
	}
	if iface.Address6 != "fe80::1234:aff:fecd:490e/64" {
		t.Errorf("Wrong interface Address6 %#v", ifaces)
	}
	if iface.State != "UP" {
		t.Errorf("Wrong interface State %#v", ifaces)
	}
	if iface.Mac != "11:22:33:44:55:66" {
		t.Errorf("Wrong interface Mac %#v", ifaces)
	}
	if iface.Kind != "ether" {
		t.Errorf("Wrong interface Kind %#v", ifaces)
	}
}

func TestParseIpResult2if(t *testing.T) {
	data2 := `1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    link/ether 11:22:33:44:55:66 brd ff:ff:ff:ff:ff:ff
    inet6 fe80::1234:aff:fecd:490e/64 scope link
       valid_lft forever preferred_lft forever
3: wlan0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether 18:11:22:33:44:55 brd ff:ff:ff:ff:ff:ff
    inet 192.168.1.12/24 brd 192.168.1.255 scope global wlan0
       valid_lft forever preferred_lft forever
    inet6 fe80::a22:9111:1234:4567/64 scope link
       valid_lft forever preferred_lft forever`

	ifaces := parseIPResult(data2)
	if ifaces == nil || len(ifaces) == 0 {
		t.Errorf("Missing data %#v", ifaces)
	}

	if len(ifaces) != 2 {
		t.Errorf("Wrong interfaces number %#v", ifaces)
	}
	iface := ifaces[0]
	if iface.Name != "eth0" {
		t.Errorf("Wrong interface name %#v", ifaces)
	}
	if iface.Address != "" {
		t.Errorf("Wrong interface Address %#v", ifaces)
	}
	if iface.Address6 != "fe80::1234:aff:fecd:490e/64" {
		t.Errorf("Wrong interface Address6 %#v", ifaces)
	}
	if iface.State != "UP" {
		t.Errorf("Wrong interface State %#v", ifaces)
	}
	if iface.Mac != "11:22:33:44:55:66" {
		t.Errorf("Wrong interface Mac %#v", ifaces)
	}
	if iface.Kind != "ether" {
		t.Errorf("Wrong interface Kind %#v", ifaces)
	}
	iface = ifaces[1]
	if iface.Name != "wlan0" {
		t.Errorf("Wrong interface name %#v", ifaces)
	}
	if iface.Address != "192.168.1.12/24" {
		t.Errorf("Wrong interface Address %#v", ifaces)
	}
	if iface.Address6 != "fe80::a22:9111:1234:4567/64" {
		t.Errorf("Wrong interface Address6 %#v", ifaces)
	}
	if iface.State != "UP" {
		t.Errorf("Wrong interface State %#v", ifaces)
	}
	if iface.Mac != "18:11:22:33:44:55" {
		t.Errorf("Wrong interface Mac %#v", ifaces)
	}
	if iface.Kind != "ether" {
		t.Errorf("Wrong interface Kind %#v", ifaces)
	}
}

func TestValidateTcpAddress(t *testing.T) {
	testsPositive := []string{
		"123.123.123.123:23",
		"dlkalkd.daldalk.com:43",
		"adlkal-daslak:123",
	}
	for _, str := range testsPositive {
		if !reValidateTCPAddress.MatchString(str) {
			t.Errorf("Wrong validate  %#v", str)
		}
	}
	testsNegative := []string{
		"http://123.123.123.123",
		"123.123.123.123",
		"abcd:232153",
		"dlkal kd.dald alk.com",
		"ldkalkdla",
		"1234567890123456789012345678901234567890123456789012345678901234567890:1",
	}
	for _, str := range testsNegative {
		if reValidateTCPAddress.MatchString(str) {
			t.Errorf("Wrong validate  %#v", str)
		}
	}
}

func TestValidateHttpAddress(t *testing.T) {
	testsPositive := []string{
		"123.123.123.123:23",
		"dlkalkd.daldalk.com:80",
		"adlkal-daslak",
		"http://adlkal-daslak",
		"https://adlkal-daslak",
		"http://adlkal-daslak/",
		"http://adlkal-daslak/asdada",
		"http://adlkal-daslak/asdada/dlakdlak",
		"https://adlkal-daslak",
		"https://adlkal-daslak/dalkdalk",
		"adlkal-daslak/dalkdalk",
		"adlkal-daslak:123/dalkdalk",
	}
	for _, str := range testsPositive {
		if !reValidateHTTPAddress.MatchString(str) {
			t.Errorf("Wrong validate  %#v", str)
		}
	}
	testsNegative := []string{
		"//123.123.123.123",
		"123.123.123.123:1234567",
		"abcd:232153",
		"dadkla://ldkalkdla",
		"1234567890123456789012345678901234567890123456789012345678901234567890",
	}
	for _, str := range testsNegative {
		if reValidateHTTPAddress.MatchString(str) {
			t.Errorf("Wrong validate  %#v", str)
		}
	}
}

func TestValidatePingAddress(t *testing.T) {
	testsPositive := []string{
		"123.123.123.123",
		"adlkal-daslak",
	}
	for _, str := range testsPositive {
		if !reValidatePingAddress.MatchString(str) {
			t.Errorf("Wrong validate  %#v", str)
		}
	}
	testsNegative := []string{
		"//123.123.123.123",
		"123.123.123.123:1234567",
		"abcd:232153",
		"dadkla://ldkalkdla",
		"1234567890123456789012345678901234567890123456789012345678901234567890",
	}
	for _, str := range testsNegative {
		if reValidatePingAddress.MatchString(str) {
			t.Errorf("Wrong validate  %#v", str)
		}
	}
}
