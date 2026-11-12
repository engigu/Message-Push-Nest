package util

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// IsPrivateOrLocalIP 检查 IP 地址是否属于私有网络、本地回环、链路本地或云厂商元数据地址
func IsPrivateOrLocalIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	// 环境变量支持自建内网发信豁免（默认严格拦截）
	if os.Getenv("ALLOW_INTRANET_WEBHOOK") == "true" {
		return false
	}

	// 转换为 4 字节标准 IPv4（如果是 IPv4 映射的 IPv6 地址）
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}

	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}

	// 特殊拦截：云厂商元数据地址 (169.254.169.254)
	if ip.String() == "169.254.169.254" {
		return true
	}

	// 广播地址
	if ip.Equal(net.IPv4bcast) {
		return true
	}

	return false
}

// ValidateSafeURL 验证 URL 协议及目标 Host 是否安全（防 SSRF）
func ValidateSafeURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return errors.New("目标地址不能为空")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("非法 URL 格式: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("不支持的协议 [%s]，仅支持 http 和 https", parsed.Scheme)
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return errors.New("目标主机名不能为空")
	}

	lowerHost := strings.ToLower(hostname)
	if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".local") {
		return errors.New("禁止向本地回环地址发起请求 (SSRF Protection)")
	}

	// 解析 Host IP
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("域名解析失败 [%s]: %w", hostname, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("未能解析出有效的 IP 地址: %s", hostname)
	}

	for _, ip := range ips {
		if IsPrivateOrLocalIP(ip) {
			return fmt.Errorf("禁止向私有网络或本地回环地址 [%s] 发起请求 (SSRF Protection)", ip.String())
		}
	}

	return nil
}

// NewSafeHTTPClient 创建防 SSRF（支持防 DNS 重新绑定攻击）的 HTTP 客户端
func NewSafeHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, err
			}

			for _, ip := range ips {
				if IsPrivateOrLocalIP(ip) {
					return nil, fmt.Errorf("SSRF 拦截：禁止连接私有网络或本地地址 [%s]", ip.String())
				}
			}

			if len(ips) == 0 {
				return nil, fmt.Errorf("SSRF 拦截：无有效解析 IP")
			}

			// 选择第一个安全的 IP 连接
			targetAddr := net.JoinHostPort(ips[0].String(), port)
			return dialer.DialContext(ctx, network, targetAddr)
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("stopped after 5 redirects")
			}
			return ValidateSafeURL(req.URL.String())
		},
	}
}

