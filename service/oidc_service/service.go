package oidc_service

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"message-nest/models"
	"message-nest/pkg/constant"
)

// BuildAuthURL 构建前端登录重定向至服务提供商的授权重定向 URL
func BuildAuthURL(ctx context.Context, cfg *Config, state, verifier string) (string, error) {
	_, oauthCfg, err := GetProviderFor(ctx, cfg)
	if err != nil {
		return "", err
	}
	authURL := oauthCfg.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	return authURL, nil
}

// ExchangeAndResolve 用授权码换取用户信息并解析/匹配/开通系统用户
func ExchangeAndResolve(ctx context.Context, cfg *Config, code, verifier string) (*models.Auth, bool, error) {
	if cfg == nil {
		return nil, false, fmt.Errorf("登录配置不能为空")
	}

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	verifierObj, oauthCfg, err := GetProviderFor(reqCtx, cfg)
	if err != nil {
		return nil, false, fmt.Errorf("无法连接登录提供商: %w", err)
	}

	token, err := oauthCfg.Exchange(reqCtx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		token, err = oauthCfg.Exchange(reqCtx, code)
	}
	if err != nil {
		return nil, false, fmt.Errorf("授权码换取令牌失败: %w", err)
	}

	var claims Claims
	if cfg.Type == constant.SSOTypeOAuth2 {
		cObj, err := FetchOAuth2Userinfo(reqCtx, cfg, token.AccessToken)
		if err != nil {
			return nil, false, fmt.Errorf("获取用户信息失败: %w", err)
		}
		claims = *cObj
	} else {
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok || rawIDToken == "" {
			return nil, false, fmt.Errorf("OIDC 提供商未返回 id_token")
		}

		idToken, err := verifierObj.Verify(reqCtx, rawIDToken)
		if err != nil {
			return nil, false, fmt.Errorf("id_token 校验失败: %w", err)
		}
		if err := idToken.Claims(&claims); err != nil {
			return nil, false, fmt.Errorf("无法解析 OIDC 声明: %w", err)
		}
	}

	user, created, err := ResolveUser(cfg, &claims)
	if err != nil {
		return nil, false, err
	}
	if user == nil {
		return nil, false, fmt.Errorf("当前身份未关联任何系统账号，且未自动开通")
	}

	return user, created, nil
}
