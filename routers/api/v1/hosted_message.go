package v1

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/util"
	"message-nest/service/hosted_message_service"
	"message-nest/service/settings_service"
	"net/http"
)

// GetHostMessageList 获取托管消息列表
func GetHostMessageList(c *gin.Context) {
	appG := app.Gin{C: c}
	text := c.Query("text")

	offset, limit := util.GetPageSize(c)
	messageService := hosted_message_service.HostMessageService{
		Text:     text,
		PageNum:  offset,
		PageSize: limit,
	}
	ways, err := messageService.GetAll()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取托管消息失败！", nil)
		return
	}

	count, err := messageService.Count()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取托管消息总数失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取托管消息成功", map[string]interface{}{
		"lists": ways,
		"total": count,
	})
}

// HostedMessagePreviewVO 托管消息预览的 VO 视图层
type HostedMessagePreviewVO struct {
	Title      string `json:"title"`      // 密文 Title
	Content    string `json:"content"`    // 密文 Content
	Type       string `json:"type"`       // 消息格式类型
	CreatedOn  string `json:"created_on"` // 创建时间
	S          string `json:"s"`          // 解密密钥 (混淆后的)
}

// obfuscateKey 混淆密钥 (邻近交换 + 对调前后半段)
func obfuscateKey(s string) string {
	runes := []rune(s)
	n := len(runes)
	if n < 2 {
		return s
	}
	// 1. 邻近两两交换
	for i := 0; i < n-1; i += 2 {
		runes[i], runes[i+1] = runes[i+1], runes[i]
	}
	// 2. 对调前后半段
	mid := n / 2
	for i := 0; i < mid; i++ {
		runes[i], runes[mid+i] = runes[mid+i], runes[i]
	}
	return string(runes)
}

// GetHostMessagePreview 公开获取单条托管消息预览（免认证，带安全加密与VO过滤）
func GetHostMessagePreview(c *gin.Context) {
	appG := app.Gin{C: c}
	key := c.Query("key")
	if key == "" {
		appG.CResponse(http.StatusBadRequest, "参数错误！", nil)
		return
	}

	// 1. 检查站点配置是否允许公开预览
	settingsService := settings_service.UserSettings{}
	siteConfig, err := settingsService.GetUserSetting("site_config")
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取配置失败！", nil)
		return
	}

	enableVal, exists := siteConfig["enable_public_preview"]
	if exists && enableVal == "false" {
		appG.CResponse(http.StatusForbidden, "站点已禁用公开消息预览功能！", nil)
		return
	}

	// 2. 查询消息
	msg, err := models.GetHostMessageByUniqueKey(key)
	if err != nil {
		appG.CResponse(http.StatusNotFound, "消息不存在或已被删除！", nil)
		return
	}

	// 3. 生成随机 24 字节 3DES 密钥
	rawKey := []byte(util.GenerateRandomString(24))

	// 4. 对 Title 和 Content 进行 3DES 加密
	encryptedTitle, err := util.TripleDesEncrypt([]byte(msg.Title), rawKey)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "消息数据处理失败！", nil)
		return
	}

	encryptedContent, err := util.TripleDesEncrypt([]byte(msg.Content), rawKey)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "消息数据处理失败！", nil)
		return
	}

	// 将 Base64 密钥混淆
	base64Key := base64.StdEncoding.EncodeToString(rawKey)
	obfuscatedKey := obfuscateKey(base64Key)

	// 5. 组装 VO，只暴露必要的信息并包含混淆秘钥
	vo := HostedMessagePreviewVO{
		Title:      encryptedTitle,
		Content:    encryptedContent,
		Type:       msg.Type,
		CreatedOn:  msg.CreatedOn.String(),
		S:          obfuscatedKey,
	}

	appG.CResponse(http.StatusOK, "获取消息成功", vo)
}
