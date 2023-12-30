package send_message_service

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/logging"
	"message-nest/pkg/message"
	"message-nest/service/send_task_service"
	"message-nest/service/send_way_service"
	"strings"
)

type SendMessageService struct {
	TaskID   string
	Text     string
	HTML     string
	MarkDown string
}

func (sm *SendMessageService) Send() string {
	var logOutput []string
	status := 1

	sendTaskService := send_task_service.SendTaskService{
		ID: sm.TaskID,
	}
	task, err := sendTaskService.GetTaskWithIns()
	if err != nil {
		return fmt.Sprintf("任务不存在！任务id: %s", sm.TaskID)
	}

	for idx, ins := range task.InsData {
		way, err := models.GetWayByID(ins.WayID)
		if err != nil {
			logOutput = append(logOutput, fmt.Sprintf("渠道信息不存在！渠道id：%s", ins.WayID))
		}
		wayService := send_way_service.SendWay{
			ID:   fmt.Sprintf("%s", way.ID),
			Name: way.Name,
			Auth: way.Auth,
			Type: way.Type,
		}

		logOutput = append(logOutput, fmt.Sprintf(">> 实例 %d", idx+1))
		logOutput = append(logOutput, fmt.Sprintf("开始发送，实例: %s", ins.WayID))
		logOutput = append(logOutput, fmt.Sprintf("实例类型: %s + %s", ins.WayType, ins.ContentType))
		logOutput = append(logOutput, fmt.Sprintf("实例配置: %s", ins.Config))

		errStr, msgObj := wayService.ValidateDiffWay()
		if errStr != "" {
			sm.MarkStatus(errStr, &status)
			logOutput = append(logOutput, fmt.Sprintf("实例渠道认证校验失败: %s", errStr))
			continue
		}

		// 邮箱类型的实例
		emailAuth, ok := msgObj.(send_way_service.WayDetailEmail)
		if ok {
			errMsg := sm.SendTaskEmail(emailAuth)
			sm.MarkStatus(errMsg, &status)
			logOutput = append(logOutput, sm.TransError(errMsg))
			continue
		}

		logOutput = append(logOutput, fmt.Sprintf("未知渠道的发信实例: %s", ins.ID))

	}

	logOutput = sm.FormatSendContent(logOutput)
	sm.RecordSendLog(logOutput, status)

	return ""
}

// FormatSendContent 格式化输出的发送内容
func (sm *SendMessageService) FormatSendContent(logOutput []string) []string {
	logOutput = append(logOutput, fmt.Sprintf(">> 发送的内容:"))
	if sm.Text != "" {
		logOutput = append(logOutput, fmt.Sprintf("Text: %s", sm.Text))
	}
	if sm.HTML != "" {
		logOutput = append(logOutput, fmt.Sprintf("HTML: %s", sm.HTML))
	}
	if sm.MarkDown != "" {
		logOutput = append(logOutput, fmt.Sprintf("MarkDown: %s", sm.MarkDown))
	}
	return logOutput
}

// MarkStatus 标记任务状态
func (sm *SendMessageService) MarkStatus(errStr string, status *int) {
	if errStr != "" {
		*status = 0
	}
}

// RecordSendLog 记录发送日志
func (sm *SendMessageService) RecordSendLog(logOutput []string, status int) {
	if len(logOutput) <= 0 {
		return
	}
	log := models.SendTasksLogs{
		Log:    strings.Join(logOutput, "\n"),
		TaskID: sm.TaskID,
		Status: status,
	}
	err := log.Add()
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("添加日志失败！原因是：%s", err))
	}
}

// TransError 转化错误
func (sm *SendMessageService) TransError(err string) string {
	if err == "" {
		return "发送成功！\n"
	} else {
		return fmt.Sprintf("发送失败：%s！", err)
	}
}

// SendTaskEmail 执行发送邮件
func (sm *SendMessageService) SendTaskEmail(auth send_way_service.WayDetailEmail) string {
	var emailer message.EmailMessage
	emailer.Init(auth.Server, auth.Port, auth.Account, auth.Passwd)
	//errMsg := emailer.SendTextMessage("sayheya@qq.com", "test", "This is a test email from message-nest.")
	//return errMsg
	return ""
}
