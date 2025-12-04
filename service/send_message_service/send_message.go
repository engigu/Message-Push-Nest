package send_message_service

import (
	"errors"
	"fmt"
	"message-nest/models"
	"message-nest/pkg/constant"
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

func errStrIsSuccess(errStr string) int {
	if errStr == "" {
		return SendSuccess
	}
	return SendFail
}

type SendMessageService struct {
	TaskID   string
	Title    string
	Text     string
	HTML     string
	URL      string
	MarkDown string
	CallerIp string

	// @提及相关字段
	AtMobiles []string
	AtUserIds []string
	AtAll     bool

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
func (sm *SendMessageService) SendPreCheck() (models.TaskIns, error) {
	errStr := ""
	entry := logrus.WithFields(logrus.Fields{
		"prefix": "[Message PreChecK]",
	})
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
	return task, nil
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

		// 发送内容校验绑定
		typeC, content := sm.GetSendMsg(ins.SendTasksIns)
		if content == "" {
			sm.LogsAndStatusMark(fmt.Sprintf("发送内容为空，设置的类型: %s，实际检测的类型: %s", ins.SendTasksIns.ContentType, typeC), SendFail)
			continue
		}

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

		// 构建统一消息内容（支持@功能）
		unifiedContent := &unified.UnifiedMessageContent{
			Title:     sm.Title,
			Text:      sm.Text,
			HTML:      sm.HTML,
			Markdown:  sm.MarkDown,
			URL:       sm.URL,
			AtMobiles: sm.AtMobiles,
			AtUserIds: sm.AtUserIds,
			AtAll:     sm.AtAll,
		}

		// 使用 SendUnified 方法（自动格式转换和@功能支持）
		res, errMsg := channel.SendUnified(msgObj, ins.SendTasksIns, unifiedContent)
		if res != "" {
			sm.LogsAndStatusMark(fmt.Sprintf("返回内容：%s", res), sm.Status)
		} else {
			sm.LogsAndStatusMark(sm.TransError(errMsg), errStrIsSuccess(errMsg))
		}

	}

	// 追加记录发送内容
	sm.AppendSendContent()
	// 日志写到数据库
	sm.RecordSendLog()

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
	log := models.SendTasksLogs{
		Log:      strings.Join(sm.LogOutput, "\n"),
		TaskID:   sm.TaskID,
		Status:   &sm.Status,
		CallerIp: sm.CallerIp,
	}
	err := log.Add()
	if err != nil {
		sm.DefaultLogger.Errorf("添加日志失败！原因是：%s", err)
	}
}

// TransError 转化错误
func (sm *SendMessageService) TransError(err string) string {
	if err == "" {
		return "发送成功！\n"
	} else {
		return fmt.Sprintf("发送失败：%s！\n", err)
	}
}

// GetSendMsg 获取对应消息内容
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
