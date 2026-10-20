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

// ExchangeClaims 用授权码换取用户信息 Claims
func ExchangeClaims(ctx context.Context, cfg *Config, code, verifier string) (*Claims, error) {
	if cfg == nil {
		return nil, fmt.Errorf("登录配置不能为空")
	}

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	verifierObj, oauthCfg, err := GetProviderFor(reqCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("无法连接登录提供商: %w", err)
	}

	token, err := oauthCfg.Exchange(reqCtx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		token, err = oauthCfg.Exchange(reqCtx, code)
	}
	if err != nil {
		return nil, fmt.Errorf("授权码换取令牌失败: %w", err)
	}

	var claims Claims
	if cfg.Type == constant.SSOTypeOAuth2 {
		cObj, err := FetchOAuth2Userinfo(reqCtx, cfg, token.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("获取用户信息失败: %w", err)
		}
		claims = *cObj
	} else {
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok || rawIDToken == "" {
			return nil, fmt.Errorf("OIDC 提供商未返回 id_token")
		}

		idToken, err := verifierObj.Verify(reqCtx, rawIDToken)
		if err != nil {
			return nil, fmt.Errorf("id_token 校验失败: %w", err)
		}
		if err := idToken.Claims(&claims); err != nil {
			return nil, fmt.Errorf("无法解析 OIDC 声明: %w", err)
		}
	}
	return &claims, nil
}

// ExchangeAndResolve 用授权码换取用户信息并解析/匹配/开通系统用户
func ExchangeAndResolve(ctx context.Context, cfg *Config, code, verifier string) (*models.Auth, bool, error) {
	claims, err := ExchangeClaims(ctx, cfg, code, verifier)
	if err != nil {
		return nil, false, err
	}

	user, created, err := ResolveUser(cfg, claims)
	if err != nil {
		return nil, false, err
	}
	if user == nil {
		return nil, false, fmt.Errorf("当前身份未关联任何系统账号，且未自动开通")
	}

	return user, created, nil
}

// BindUserWithClaims 将第三方 Claims 绑定到指定已有系统用户 (如管理员账号，支持多SSO同时绑定)
func BindUserWithClaims(cfg *Config, userID string, claims *Claims) (*models.Auth, error) {
	if claims == nil || claims.Sub == "" {
		return nil, fmt.Errorf("第三方身份标识 (sub) 不能为空")
	}

	targetUser, err := models.GetUserByID(userID)
	if err != nil || targetUser == nil || targetUser.ID == "" {
		return nil, fmt.Errorf("要绑定的系统账号不存在")
	}

	// 1. 调用多SSO绑定方法更新/新增关系
	err = models.BindUserSSO(userID, cfg.Type, claims.Sub, claims.Email)
	if err != nil {
		return nil, err
	}

	targetUser.OidcSubject = claims.Sub
	targetUser.AuthProvider = cfg.Type

	// 2. 若系统账号未填写邮箱，自动同步第三方提供的邮箱
	if targetUser.Email == "" && claims.Email != "" {
		_ = models.EditUserByID(targetUser.ID, map[string]interface{}{"email": claims.Email})
		targetUser.Email = claims.Email
	}

	return targetUser, nil
}
