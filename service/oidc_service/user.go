package oidc_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"message-nest/models"
	"message-nest/pkg/util"
)

// ErrUserNotLinked 提示账号未关联且系统未开启自动建号
var ErrUserNotLinked = errors.New("当前账号未关联任何系统账号，请联系管理员开通或开启首次登录自动建号")

// Claims OIDC / OAuth 2.0 统一用户身份声明
type Claims struct {
	Sub               string   `json:"sub"`
	Email             string   `json:"email"`
	PreferredUsername string   `json:"preferred_username"`
	Name              string   `json:"name"`
	Groups            []string `json:"groups"`
}

// ResolveUser 根据 subject 查验用户、绑定已有账号或自动创建新账号
func ResolveUser(cfg *Config, claims *Claims) (*models.Auth, bool, error) {
	if claims == nil || claims.Sub == "" {
		return nil, false, fmt.Errorf("身份标识 (sub) 不能为空")
	}

	// 1. 尝试从独立身份关联表（并平滑兜底 Auth 表）匹配该 Provider 和 Subject
	user, err := models.GetUserBySSO(cfg.Type, claims.Sub)
	if err == nil && user != nil && user.ID != "" {
		return user, false, nil
	}

	// 2. 匹配已有账号（相同邮箱），自动关联至该 Provider
	if claims.Email != "" {
		user, err = models.GetUserByEmail(claims.Email)
		if err == nil && user != nil && user.ID != "" {
			_ = models.BindUserSSO(user.ID, cfg.Type, claims.Sub, claims.Email)
			user.OidcSubject = claims.Sub
			user.AuthProvider = cfg.Type
			return user, false, nil
		}
	}

	if !cfg.AutoCreate {
		return nil, false, ErrUserNotLinked
	}

	// 3. 自动建号
	username := claims.PreferredUsername
	if username == "" {
		username = claims.Email
	}
	if username == "" {
		username = "sso_" + claims.Sub
	}

	finalUsername := UniqueUsername(username)
	randomPassword := util.GenerateRandomString(16)

	created, err := models.AddSSOUser(finalUsername, randomPassword, claims.Email, claims.Sub, cfg.Type)
	if err != nil {
		return nil, false, fmt.Errorf("自动开通账号失败: %w", err)
	}

	// 在独立关联表中登记绑定
	_ = models.BindUserSSO(created.ID, cfg.Type, claims.Sub, claims.Email)

	return created, true, nil
}

// UniqueUsername 确保生成的系统唯一用户名，避免冲突
func UniqueUsername(base string) string {
	candidate := base
	if candidate == "" {
		candidate = "user"
	}
	for i := 0; i < 20; i++ {
		u, err := models.GetUserByUsername(candidate)
		if err != nil || u == nil || u.ID == "" {
			return candidate
		}
		candidate = fmt.Sprintf("%s_%s", base, util.GenerateRandomString(4))
	}
	return fmt.Sprintf("%s_%s", base, util.GenerateRandomString(8))
}

// FetchOAuth2Userinfo 通用 Access Token 请求 Userinfo 端点获取用户信息
func FetchOAuth2Userinfo(ctx context.Context, cfg *Config, accessToken string) (*Claims, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserinfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("构建用户信息请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Message-Nest-SSO/1.0")

	client := NewSSOHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求用户信息接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("用户信息接口响应状态异常 (%d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, fmt.Errorf("读取用户信息失败: %w", err)
	}

	claims, err := ParseOAuth2Userinfo(body, cfg.UserKey)
	if err != nil {
		return nil, err
	}

	// 针对 GitHub 等平台，如果用户邮箱为私有，尝试从 /user/emails 获取
	if claims.Email == "" && strings.Contains(cfg.UserinfoURL, "api.github.com") {
		if ghEmail := FetchGitHubPrimaryEmail(ctx, accessToken); ghEmail != "" {
			claims.Email = ghEmail
		}
	}

	return claims, nil
}

// ParseOAuth2Userinfo 解析 OAuth 2.0 用户信息响应 JSON 提取关键字段
func ParseOAuth2Userinfo(raw []byte, customUserKey string) (*Claims, error) {
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("解析用户信息 JSON 失败: %w", err)
	}

	claims := &Claims{}

	// 1. 获取唯一标识 Sub
	if customUserKey != "" {
		if val, ok := data[customUserKey]; ok && val != nil {
			claims.Sub = strings.TrimSpace(fmt.Sprintf("%v", val))
		}
	}
	if claims.Sub == "" {
		candidates := []string{"sub", "id", "user_id", "uid", "openid", "login", "username"}
		for _, key := range candidates {
			if val, ok := data[key]; ok && val != nil {
				str := strings.TrimSpace(fmt.Sprintf("%v", val))
				if str != "" && str != "<nil>" {
					claims.Sub = str
					break
				}
			}
		}
	}
	if claims.Sub == "" {
		return nil, fmt.Errorf("未能从用户信息中识别出唯一身份标识 (ID)")
	}

	// 2. 获取用户名
	userCandidates := []string{"login", "username", "preferred_username", "name", "nickname"}
	for _, key := range userCandidates {
		if val, ok := data[key]; ok && val != nil {
			str := strings.TrimSpace(fmt.Sprintf("%v", val))
			if str != "" && str != "<nil>" {
				claims.PreferredUsername = str
				break
			}
		}
	}
	if claims.PreferredUsername == "" {
		claims.PreferredUsername = claims.Sub
	}

	// 3. 获取邮箱
	emailCandidates := []string{"email", "mail", "user_email"}
	for _, key := range emailCandidates {
		if val, ok := data[key]; ok && val != nil {
			str := strings.TrimSpace(fmt.Sprintf("%v", val))
			if str != "" && strings.Contains(str, "@") {
				claims.Email = str
				break
			}
		}
	}

	// 4. 获取显示名
	if val, ok := data["name"]; ok && val != nil {
		claims.Name = fmt.Sprintf("%v", val)
	}

	return claims, nil
}

// FetchGitHubPrimaryEmail 通过 GitHub /user/emails API 获取已验证的主邮箱
func FetchGitHubPrimaryEmail(ctx context.Context, token string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Message-Nest-SSO/1.0")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return ""
	}
	for _, e := range emails {
		if e.Primary && e.Verified && e.Email != "" {
			return e.Email
		}
	}
	if len(emails) > 0 {
		return emails[0].Email
	}
	return ""
}
