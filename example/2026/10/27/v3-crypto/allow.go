package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// AllowList 限制服务端允许拨号的目标地址，这是治 SSRF 的那一层。
//
// 规则只接受 IP 或 CIDR，不接受域名：域名会在"校验"和"真正拨号"之间留下二次解析的缝，
// 也就是 DNS rebinding——校验时解析到内网地址，拨号时解析到另一个地址。
// 因此 Resolve 一旦通过，就返回 **IP 形式**的地址，直接拿去拨号，不再解析第二次。
type AllowList struct {
	rules []allowRule
}

type allowRule struct {
	network *net.IPNet // CIDR 规则
	ip      net.IP     // 单个 IP
	port    uint16
	anyPort bool
}

// NewAllowList 解析白名单规则，形如：
//
//	127.0.0.1:8000    精确
//	127.0.0.1:*       该 IP 的任意端口
//	10.0.0.0/8:3306   整个网段 + 指定端口
func NewAllowList(specs []string) (*AllowList, error) {
	a := &AllowList{}
	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}
		rule, err := parseAllowRule(spec)
		if err != nil {
			return nil, err
		}
		a.rules = append(a.rules, rule)
	}
	return a, nil
}

func parseAllowRule(spec string) (allowRule, error) {
	host, portStr, err := net.SplitHostPort(spec)
	if err != nil {
		return allowRule{}, fmt.Errorf("invalid allow rule %q: %w", spec, err)
	}

	var rule allowRule
	switch portStr {
	case "*", "":
		rule.anyPort = true
	default:
		p, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return allowRule{}, fmt.Errorf("invalid port in allow rule %q", spec)
		}
		rule.port = uint16(p)
	}

	if strings.Contains(host, "/") {
		_, network, err := net.ParseCIDR(host)
		if err != nil {
			return allowRule{}, fmt.Errorf("invalid CIDR in allow rule %q", spec)
		}
		rule.network = network
		return rule, nil
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return allowRule{}, fmt.Errorf("allow rule must use IP or CIDR, got %q", host)
	}
	rule.ip = ip
	return rule, nil
}

// Resolve 判断目标地址是否被允许，返回可直接拨号的地址（IP:port）。
//
// 目标写的是域名时先解析，只要解析出的任一 IP 命中规则就放行，
// 但返回的是命中的那个 IP——这样校验和拨号用的是同一个地址，不给 rebinding 留机会。
func (a *AllowList) Resolve(addr string) (string, bool) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", false
	}
	p, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return "", false
	}
	port := uint16(p)

	if ip := net.ParseIP(host); ip != nil {
		if a.matches(ip, port) {
			return net.JoinHostPort(ip.String(), portStr), true
		}
		return "", false
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return "", false
	}
	for _, ip := range ips {
		if a.matches(ip, port) {
			return net.JoinHostPort(ip.String(), portStr), true
		}
	}
	return "", false
}

func (a *AllowList) matches(ip net.IP, port uint16) bool {
	for _, rule := range a.rules {
		if !rule.anyPort && rule.port != port {
			continue
		}
		if rule.network != nil && rule.network.Contains(ip) {
			return true
		}
		if rule.ip != nil && rule.ip.Equal(ip) {
			return true
		}
	}
	return false
}
