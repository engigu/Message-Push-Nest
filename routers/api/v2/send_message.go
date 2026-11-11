package v2

import (
	"encoding/json"
	"fmt"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/e"
	"message-nest/pkg/logging"
	utilpkg "message-nest/pkg/util"
	"message-nest/service/send_message_service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type SendMessageByTemplateReq struct {
	Token        string                 `json:"token" validate:"required" label:"模板token"`
	Title        string                 `json:"title" validate:"required" label:"消息标题"`
	Placeholders map[string]interface{} `json:"placeholders" label:"占位符"`
	Recipients   []string               `json:"recipients" label:"接收者列表"`
}

type TemplatePlaceholder struct {
	Key     string `json:"key"`
	Default string `json:"default"`
}

func selectDeclaredTemplateParams(definitions []TemplatePlaceholder, values map[string]interface{}) map[string]interface{} {
	selected := make(map[string]interface{}, len(definitions))
	for _, definition := range definitions {
		if definition.Key == "" {
			continue
		}
		if value, exists := values[definition.Key]; exists {
			selected[definition.Key] = value
		}
	}
	return selected
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

	var (
		template       *models.TemplateResult
		effectiveOwner string
	)
	user, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		// 外部系统免登录发信（必须持有正确的加密 token）
		var errTmpl error
		template, errTmpl = models.GetTemplateByID(templateID)
		if errTmpl != nil || template == nil {
			appG.CResponse(http.StatusNotFound, "模板不存在", nil)
			return
		}
		ownerID = template.CreatedBy
		effectiveOwner = template.CreatedBy
	} else {
		// 已登录用户（包含租户与管理员）：强校验模板归属，杜绝跨租户越权发信
		var errTmpl error
		template, errTmpl = models.GetTemplateByIDForOwner(templateID, ownerID)
		if errTmpl != nil || template == nil {
			appG.CResponse(http.StatusNotFound, "模板不存在或无权访问", nil)
			return
		}
		effectiveOwner = user.GetEffectiveUserID()
	}

	// 检查模板状态
	if template.Status != "enabled" {
		appG.CResponse(http.StatusBadRequest, "模板已禁用", nil)
		return
	}

	// 填充默认占位符
	var placeholderDefs []TemplatePlaceholder
	if template.Placeholders != "" {
		if err := json.Unmarshal([]byte(template.Placeholders), &placeholderDefs); err == nil {
			if req.Placeholders == nil {
				req.Placeholders = make(map[string]interface{})
			}
			for _, def := range placeholderDefs {
				// 如果请求中没有该占位符，且配置了默认值，则使用默认值
				if _, exists := req.Placeholders[def.Key]; !exists && def.Default != "" {
					req.Placeholders[def.Key] = def.Default
				}
			}
		} else {
			logging.TemplateSend.Errorf("解析模板占位符配置失败: %v", err)
		}
	}

	// 仅把模板定义过的业务变量传给短信结构化参数路径；额外请求字段仍
	// 可用于既有文本替换，但不会因此成为阿里云审核模板的可发送参数。
	templateParams := selectDeclaredTemplateParams(placeholderDefs, req.Placeholders)

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

	// 获取模板关联的实例列表（租户安全范围）
	insList, err := models.GetTemplateInsListForOwner(templateID, ownerID)
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
		SendMode:       send_message_service.SendModeTemplate, // 明确标记为模板模式
		TaskID:         templateID,                            // 使用模板ID作为TaskID（用于日志记录）
		TemplateID:     templateID,                            // 模板ID
		CreatedBy:      effectiveOwner,
		Name:           template.Name,                         // 模板名称
		Title:          req.Title,
		Text:           textContent,
		HTML:           htmlContent,
		MarkDown:       markdownContent,
		CallerIp:       c.ClientIP(),
		AtMobiles:      atMobiles,
		AtUserIds:      atUserIds,
		AtAll:          template.IsAtAll,
		Recipients:     req.Recipients, // 动态接收者列表
		TemplateParams: templateParams,
		DefaultLogger:  logging.TemplateSend,
	}

	// 发送前检查
	task, err := msgService.SendPreCheck()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("发送检查不通过：%s", err), nil)
		return
	}

	// 检查是否包含 MessageNest 类型的启用实例，并生成托管键与链接
	hostedMsgKey, hostMessageUrl := send_message_service.CheckAndGenerateHostedKey(task, c.Request.TLS != nil, c.Request.Host)

	responseData := map[string]interface{}{
		"token": req.Token,
		"count": enabledCount,
	}
	if hostMessageUrl != "" {
		responseData["hostMessageUrl"] = hostMessageUrl
	}

	// 传导预生成的 key 到后台协程
	msgService.HostedMsgKey = hostedMsgKey

	// 异步发送
	msgService.AsyncSend(task)
	appG.CResponse(http.StatusOK, "success", responseData)
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
