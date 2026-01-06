package send_message_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"message-nest/models"
	"message-nest/pkg/constant"
	"message-nest/pkg/util"
	"message-nest/service/send_message_service/unified"
	"message-nest/service/send_task_service"
	"message-nest/service/send_way_service"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	SendSuccess = 1
	SendFail    = 0
)

// 发送模式类型
const (
	SendModeTask     = "task"     // 传统任务模式
	SendModeTemplate = "template" // 模板模式
)

func errStrIsSuccess(errStr string) int {
	if errStr == "" {
		return SendSuccess
	}
	return SendFail
}

type SendMessageService struct {
	SendMode   string // 发送模式：task(任务模式) 或 template(模板模式)
	TaskID     string // 任务ID（任务模式）或模板ID（模板模式，用于日志记录）
	TemplateID string // 模板ID（仅模板模式使用）
	Name       string // 任务或模板名称（用于日志记录）
	Title      string
	Text       string
	HTML       string
	URL        string
	MarkDown   string
	CallerIp   string

	// @提及相关字段
	AtMobiles []string
	AtUserIds []string
	AtAll     bool

	// 动态接收者（用于邮箱、微信公众号等支持群发的渠道）
	Recipients []string

	Status    int
	LogOutput []string

	DefaultLogger *logrus.Entry
}

// LogsAndStatusMark 记录执行的日志和状态标记
func (sm *SendMessageService) LogsAndStatusMark(errStr string, status int) {
	sm.LogOutput = append(sm.LogOutput, errStr)
	if status == SendFail {
		sm.Status = SendFail
	}
	sm.DefaultLogger.Infof("%s, 状态：%d", strings.Trim(errStr, "\n"), status)
}

// AsyncSend 异步发送一个消息任务的所有实例
func (sm *SendMessageService) AsyncSend(task models.TaskIns) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Error("AsyncSend: Recovered from panic:", r)
		}
	}()

	// 限制并发异步发送数量
	constant.MaxSendTaskSemaphoreChan <- ""
	defer func() {
		<-constant.MaxSendTaskSemaphoreChan
	}()

	go func() {
		entry := logrus.WithFields(logrus.Fields{
			"prefix": "[Send Goroutine]",
		})
		_, err := sm.Send(task)
		if err != nil {
			entry.Errorf("任务[%s][%s]发送错误： %s", sm.TaskID, sm.Title, err)
		} else {
			entry.Infof("完成任务[%s][%s]发送", sm.TaskID, sm.Title)
		}
	}()
}

// SendPreCheck 发送前数据准备和预检查
// 支持两种模式：
// 1. SendModeTask：传统任务模式，使用 TaskID 查询任务和实例
// 2. SendModeTemplate：模板模式，使用 TemplateID 查询模板关联的实例
func (sm *SendMessageService) SendPreCheck() (models.TaskIns, error) {
	errStr := ""
	entry := logrus.WithFields(logrus.Fields{
		"prefix": "[Message PreChecK]",
	})

	var task models.TaskIns

	switch sm.SendMode {
	case SendModeTemplate:
		// 模板模式：使用模板ID获取实例
		if sm.TemplateID == "" {
			errStr = "模板模式下 TemplateID 不能为空"
			entry.Errorf(errStr)
			return task, errors.New(errStr)
		}

		// 获取模板关联的实例列表
		insList, err := models.GetTemplateInsList(sm.TemplateID)
		if err != nil {
			errStr = fmt.Sprintf("模板[%s]实例查询失败：%s", sm.TemplateID, err)
			entry.Errorf(errStr)
			return task, errors.New(errStr)
		}
		if len(insList) == 0 {
			errStr = fmt.Sprintf("模板[%s]没有关联任何实例！", sm.TemplateID)
			entry.Errorf(errStr)
			return task, errors.New(errStr)
		}

		// 构造虚拟任务对象（用于兼容现有发送逻辑）
		// 将模板ID作为TaskID使用，便于日志记录
		task.ID = sm.TaskID // 使用传入的TaskID（实际是模板ID）
		task.InsData = make([]models.SendTasksInsRes, 0, len(insList))
		for _, ins := range insList {
			task.InsData = append(task.InsData, ins)
		}
		entry.Infof("模板[%s]加载了 %d 个实例", sm.TemplateID, len(insList))
		return task, nil

	case SendModeTask:
		// 传统任务模式：使用任务ID查询
		if sm.TaskID == "" {
			errStr = "任务模式下 TaskID 不能为空"
			entry.Errorf(errStr)
			return task, errors.New(errStr)
		}

		sendTaskService := send_task_service.SendTaskService{
			ID: sm.TaskID,
		}
		task, err := sendTaskService.GetTaskWithIns()
		if err != nil {
			errStr = fmt.Sprintf("任务[%s]查询失败！", sm.TaskID)
			entry.Errorf(errStr)
			return task, errors.New(errStr)
		}
		if task.ID == "" {
			errStr = fmt.Sprintf("任务[%s]不存在！", sm.TaskID)
			entry.Errorf(errStr)
			return task, errors.New(errStr)
		}
		if len(task.InsData) == 0 {
			errStr = fmt.Sprintf("任务[%s]没有关联任何实例！！", sm.TaskID)
			entry.Errorf(errStr)
			return task, errors.New(errStr)
		}
		// 设置任务名称用于日志记录
		if sm.Name == "" {
			sm.Name = task.Name
		}
		return task, nil

	default:
		// SendMode 未设置或无效
		errStr = fmt.Sprintf("SendMode 未设置或无效: %s，必须是 '%s' 或 '%s'", sm.SendMode, SendModeTask, SendModeTemplate)
		entry.Errorf(errStr)
		return task, errors.New(errStr)
	}
}

// Send 发送一个消息任务的所有实例
func (sm *SendMessageService) Send(task models.TaskIns) (string, error) {
	sm.Status = SendSuccess

	sm.LogsAndStatusMark(fmt.Sprintf("发送标题《%s》 \n", sm.Title), sm.Status)
	sm.LogsAndStatusMark(fmt.Sprintf("开始任务[%s]的发送", sm.TaskID), sm.Status)

	for idx, ins := range task.InsData {
		way, err := models.GetWayByID(ins.WayID)
		if err != nil {
			errStr := fmt.Sprintf("渠道[%s]信息不存在！跳过这个实例的发送", ins.WayID)
			sm.LogsAndStatusMark(errStr, SendFail)
			continue
		}

		// 暂停了实例的发送
		if ins.Enable != 1 {
			//sm.LogsAndStatusMark("该实例发送已经被暂停，跳过发送！\n", sm.Status)
			continue
		}

		wayService := send_way_service.SendWay{
			ID:   fmt.Sprintf("%s", way.ID),
			Name: way.Name,
			Auth: way.Auth,
			Type: way.Type,
		}

		sm.LogsAndStatusMark(fmt.Sprintf(">> 实例 %d", idx+1), sm.Status)
		sm.LogsAndStatusMark(fmt.Sprintf("开始发送，实例: %s", ins.WayID), sm.Status)
		sm.LogsAndStatusMark(fmt.Sprintf("实例渠道名: %s", way.Name), sm.Status)
		sm.LogsAndStatusMark(fmt.Sprintf("实例类型: %s + %s", ins.WayType, ins.ContentType), sm.Status)
		sm.LogsAndStatusMark(fmt.Sprintf("实例配置: %s", ins.Config), sm.Status)

		// 发送渠道的校验
		errStr, msgObj := wayService.ValidateDiffWay()
		if errStr != "" {
			sm.LogsAndStatusMark(fmt.Sprintf("实例渠道认证校验失败: %s", errStr), SendFail)
			continue
		}

		// 使用新的Channel架构发送消息
		channelRegistry := unified.GetGlobalChannelRegistry()
		channel, ok := channelRegistry.GetChannel(way.Type)
		if !ok {
			sm.LogsAndStatusMark(fmt.Sprintf("发送失败：未知渠道类型 %s 的发信实例: %s\n", way.Type, ins.ID), SendFail)
			continue
		}

		// 根据发送模式构建消息内容
		var unifiedContent *unified.UnifiedMessageContent
		if sm.SendMode == SendModeTemplate {
			// 模板模式：根据实例的 ContentType 精确发送对应类型的内容
			unifiedContent = sm.BuildTemplateContent(ins.SendTasksIns)
			if unifiedContent == nil {
				sm.LogsAndStatusMark(fmt.Sprintf("模板内容为空，实例类型: %s", ins.ContentType), SendFail)
				continue
			}
		} else {
			// 任务模式：使用现有逻辑（支持内容类型回退）
			typeC, content := sm.GetSendMsg(ins.SendTasksIns)
			if content == "" {
				sm.LogsAndStatusMark(fmt.Sprintf("发送内容为空，设置的类型: %s，实际检测的类型: %s", ins.SendTasksIns.ContentType, typeC), SendFail)
				continue
			}
			// 构建统一消息内容（支持@功能）
			unifiedContent = &unified.UnifiedMessageContent{
				Title:     sm.Title,
				Text:      sm.Text,
				HTML:      sm.HTML,
				Markdown:  sm.MarkDown,
				URL:       sm.URL,
				AtMobiles: sm.AtMobiles,
				AtUserIds: sm.AtUserIds,
				AtAll:     sm.AtAll,
			}
		}

		// 处理动态接收者（邮箱、微信公众号等支持群发的渠道）
		isDynamicMode := sm.isDynamicRecipientMode(ins.SendTasksIns)

		if isDynamicMode && sm.supportsDynamicRecipient(way.Type) && len(sm.Recipients) > 0 {
			// 动态接收模式：使用API传入的Recipients列表（群发）
			sm.LogsAndStatusMark(fmt.Sprintf("动态接收模式，共 %d 个接收者", len(sm.Recipients)), sm.Status)
			for recipientIdx, recipient := range sm.Recipients {
				sm.LogsAndStatusMark(fmt.Sprintf(">>> 接收者 %d/%d: %s", recipientIdx+1, len(sm.Recipients), recipient), sm.Status)

				// 临时修改实例配置中的接收者
				modifiedIns := sm.modifyInsRecipient(ins.SendTasksIns, recipient, way.Type)

				// 使用 SendUnified 方法发送
				res, errMsg := channel.SendUnified(msgObj, modifiedIns, unifiedContent)
				if res != "" {
					sm.LogsAndStatusMark(fmt.Sprintf("返回内容：%s", res), sm.Status)
				} else {
					sm.LogsAndStatusMark(sm.TransError(errMsg), errStrIsSuccess(errMsg))
				}
			}
		} else {
			// 固定接收模式：使用实例配置的to_account
			sm.LogsAndStatusMark("固定接收模式，使用实例配置的接收者", sm.Status)
			res, errMsg := channel.SendUnified(msgObj, ins.SendTasksIns, unifiedContent)
			if res != "" {
				sm.LogsAndStatusMark(fmt.Sprintf("返回内容：%s\n", res), sm.Status)
			} else {
				sm.LogsAndStatusMark(sm.TransError(errMsg), errStrIsSuccess(errMsg))
			}
		}

	}

	// 追加记录发送内容
	sm.AppendSendContent()
	// 日志写到数据库
	sm.RecordSendLog()
	// 更新统计数据（任务级别：一次任务算一次）
	sm.UpdateSendStats()

	totalOutputLog := strings.Join(sm.LogOutput, "\n")
	if sm.Status == SendSuccess {
		return totalOutputLog, nil
	} else {
		return totalOutputLog, errors.New("发送过程中有失败，请检查详细日志")
	}
}

// AppendSendContent 添加发送内容
func (sm *SendMessageService) AppendSendContent() {
	sm.LogOutput = append(sm.LogOutput, fmt.Sprintf(">> 发送的内容:"))
	if sm.Text != "" {
		sm.LogOutput = append(sm.LogOutput, fmt.Sprintf("Text: %s \n", sm.Text))
	}
	if sm.HTML != "" {
		sm.LogOutput = append(sm.LogOutput, fmt.Sprintf("HTML: %s \n", sm.HTML))
	}
	if sm.MarkDown != "" {
		sm.LogOutput = append(sm.LogOutput, fmt.Sprintf("MarkDown: %s \n", sm.MarkDown))
	}
}

// RecordSendLog 记录发送日志
func (sm *SendMessageService) RecordSendLog() {
	// 确定日志类型
	logType := "task"
	if sm.SendMode == SendModeTemplate {
		logType = "template"
	}

	log := models.SendTasksLogs{
		Log:      strings.Join(sm.LogOutput, "\n"),
		TaskID:   sm.TaskID,
		Type:     logType,
		Name:     sm.Name,
		Status:   &sm.Status,
		CallerIp: sm.CallerIp,
	}
	err := log.Add()
	if err != nil {
		sm.DefaultLogger.Errorf("添加日志失败！原因是：%s", err)
	}
}

// UpdateSendStats 更新发送统计数据（任务级别）
func (sm *SendMessageService) UpdateSendStats() {
	// 获取当前日期
	currentDay := sm.getCurrentDay()

	// 确定任务类型
	taskType := "task"
	if sm.SendMode == SendModeTemplate {
		taskType = "template"
	}

	// 根据任务的最终状态更新统计（一次任务只记录一次）
	var status string
	if sm.Status == SendSuccess {
		status = "success"
	} else {
		status = "failed"
	}

	// 更新统计：每次任务执行记录为1次
	err := models.IncrementSendStats(sm.TaskID, taskType, currentDay, status, 1)
	if err != nil {
		logrus.Errorf("更新发送统计失败：%s", err)
	}
}

// getCurrentDay 获取当前日期（YYYY-MM-DD格式）
func (sm *SendMessageService) getCurrentDay() string {
	return util.GetNowTimeStr()[:10]
}

// TransError 转化错误
func (sm *SendMessageService) TransError(err string) string {
	if err == "" {
		return "发送成功！\n"
	} else {
		return fmt.Sprintf("发送失败：%s！\n", err)
	}
}

// BuildTemplateContent 构建模板模式的消息内容
// 模板模式：根据实例的 ContentType 精确匹配对应类型的内容，只传递该类型的内容
func (sm *SendMessageService) BuildTemplateContent(ins models.SendTasksIns) *unified.UnifiedMessageContent {
	contentType := strings.ToLower(ins.ContentType)

	// 内容类型映射表
	contentMap := map[string]string{
		unified.FormatTypeText:     sm.Text,
		unified.FormatTypeHTML:     sm.HTML,
		unified.FormatTypeMarkdown: sm.MarkDown,
	}

	// 检查内容是否存在
	contentValue, exists := contentMap[contentType]
	if !exists {
		logrus.Warnf("模板模式：未知的内容类型 %s", ins.ContentType)
		return nil
	}
	if contentValue == "" {
		logrus.Warnf("模板模式：实例要求的 %s 类型内容为空", contentType)
		return nil
	}

	// 构建消息内容，只填充实例要求的类型
	content := &unified.UnifiedMessageContent{
		Title:     sm.Title,
		URL:       sm.URL,
		AtMobiles: sm.AtMobiles,
		AtUserIds: sm.AtUserIds,
		AtAll:     sm.AtAll,
	}

	// 根据类型填充对应字段
	switch contentType {
	case unified.FormatTypeText:
		content.Text = contentValue
	case unified.FormatTypeHTML:
		content.HTML = contentValue
	case unified.FormatTypeMarkdown:
		content.Markdown = contentValue
	}

	return content
}

// supportsDynamicRecipient 判断渠道是否支持动态接收者
func (sm *SendMessageService) supportsDynamicRecipient(wayType string) bool {
	// 支持动态接收者的渠道类型
	supportedTypes := map[string]bool{
		constant.MessageTypeEmail:           true,
		constant.MessageTypeWeChatOFAccount: true,
		constant.MessageTypeAliyunSMS:       true,
		// 可以继续添加其他支持动态接收者的渠道
	}
	return supportedTypes[wayType]
}

// isDynamicRecipientMode 判断实例是否配置为动态接收模式
func (sm *SendMessageService) isDynamicRecipientMode(ins models.SendTasksIns) bool {
	// 解析实例配置
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(ins.Config), &config); err != nil {
		return false
	}

	// 检查allowMultiRecip字段
	// allowMultiRecip=true: 动态模式
	// allowMultiRecip=false或不存在: 固定模式
	if allowMultiRecip, ok := config["allowMultiRecip"]; ok {
		if allow, ok := allowMultiRecip.(bool); ok {
			return allow
		}
	}

	// 默认为固定模式（兼容历史数据）
	return false
}

// modifyInsRecipient 临时修改实例配置中的接收者
func (sm *SendMessageService) modifyInsRecipient(ins models.SendTasksIns, recipient string, wayType string) models.SendTasksIns {
	// 创建副本
	modifiedIns := ins

	// 根据渠道类型修改配置
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(ins.Config), &config); err != nil {
		sm.LogsAndStatusMark(fmt.Sprintf("解析实例配置失败: %s", err.Error()), SendFail)
		return ins
	}

	// 修改接收者字段
	config["to_account"] = recipient

	// 序列化回JSON
	modifiedConfigBytes, err := json.Marshal(config)
	if err != nil {
		sm.LogsAndStatusMark(fmt.Sprintf("序列化实例配置失败: %s", err.Error()), SendFail)
		return ins
	}

	modifiedIns.Config = string(modifiedConfigBytes)
	return modifiedIns
}

// GetSendMsg 获取对应消息内容（任务模式使用）
// 先根据实例设置的类型取，取不到或者取到的是空，则使用text发送
func (sm *SendMessageService) GetSendMsg(ins models.SendTasksIns) (string, string) {
	data := map[string]string{}
	data[unified.FormatTypeText] = sm.Text
	data[unified.FormatTypeHTML] = sm.HTML
	data[unified.FormatTypeMarkdown] = sm.MarkDown
	content, ok := data[strings.ToLower(ins.ContentType)]
	if !ok || len(content) == 0 {
		content, ok := data[unified.FormatTypeText]
		if !ok {
			logrus.Error("text节点数据为空！")
			return unified.FormatTypeText, ""
		} else {
			logrus.Error(fmt.Sprintf("没有找到%s对应的消息，使用text消息替代！", ins.ContentType))
			return unified.FormatTypeText, content
		}
	} else {
		return strings.ToLower(ins.ContentType), content
	}
}
