package v2

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/e"
	utilpkg "message-nest/pkg/util"
	"message-nest/service/send_message_service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SendMessageByTemplateReq struct {
	Token        string                 `json:"token" validate:"required" label:"模板token"`
	Title        string                 `json:"title" validate:"required" label:"消息标题"`
	Placeholders map[string]interface{} `json:"placeholders" label:"占位符"`
}

// DoSendMessageByTemplate 使用模板发送消息
func DoSendMessageByTemplate(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  SendMessageByTemplateReq
	)

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	// 解析 token 为模板 ID
	templateID, err := utilpkg.DecryptTokenHex(req.Token, 71) // 71 为简单对称密钥
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("token解析失败：%v", err), nil)
		return
	}

	// 获取模板
	template, err := models.GetTemplateByID(templateID)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("模板不存在：%s", err), nil)
		return
	}

	// 检查模板状态
	if template.Status != "enabled" {
		appG.CResponse(http.StatusBadRequest, "模板已禁用", nil)
		return
	}

	// 替换占位符
	textContent := replacePlaceholders(template.TextTemplate, req.Placeholders)
	htmlContent := replacePlaceholders(template.HTMLTemplate, req.Placeholders)
	markdownContent := replacePlaceholders(template.MarkdownTemplate, req.Placeholders)

	// 解析@提醒配置
	var atMobiles []string
	var atUserIds []string
	if template.AtMobiles != "" {
		atMobiles = strings.Split(template.AtMobiles, ",")
		// 去除空格
		for i := range atMobiles {
			atMobiles[i] = strings.TrimSpace(atMobiles[i])
		}
	}
	if template.AtUserIds != "" {
		atUserIds = strings.Split(template.AtUserIds, ",")
		// 去除空格
		for i := range atUserIds {
			atUserIds[i] = strings.TrimSpace(atUserIds[i])
		}
	}

	// 获取模板关联的实例列表
	insList, err := models.GetTemplateInsList(templateID)
	if err != nil || len(insList) == 0 {
		appG.CResponse(http.StatusBadRequest, "模板没有配置发送实例", nil)
		return
	}

	// 过滤启用的实例
	var enabledCount int
	for _, ins := range insList {
		if ins.Enable == 1 {
			enabledCount++
		}
	}

	if enabledCount == 0 {
		appG.CResponse(http.StatusBadRequest, "模板没有启用的发送实例", nil)
		return
	}

	// 使用发送服务进行发送
	// 将模板ID作为TaskID传入，用于日志记录
	msgService := send_message_service.SendMessageService{
		SendMode:   send_message_service.SendModeTemplate, // 明确标记为模板模式
		TaskID:     templateID,                            // 使用模板ID作为TaskID（用于日志记录）
		TemplateID: templateID,                            // 模板ID
		Name:       template.Name,                         // 模板名称
		Title:      req.Title,
		Text:       textContent,
		HTML:       htmlContent,
		MarkDown:   markdownContent,
		CallerIp:   c.ClientIP(),
		AtMobiles:  atMobiles,
		AtUserIds:  atUserIds,
		AtAll:      template.IsAtAll,
		DefaultLogger: logrus.WithFields(logrus.Fields{
			"prefix": "[Template Send]",
		}),
	}

	// 发送前检查
	task, err := msgService.SendPreCheck()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("发送检查不通过：%s", err), nil)
		return
	}

	// 异步发送
	msgService.AsyncSend(task)
	appG.CResponse(http.StatusOK, "success", map[string]interface{}{
		"token": req.Token,
		"count": enabledCount,
	})
}

// replacePlaceholders 替换模板中的占位符
func replacePlaceholders(template string, placeholders map[string]interface{}) string {
	if template == "" || placeholders == nil {
		return template
	}

	result := template
	for key, value := range placeholders {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}
