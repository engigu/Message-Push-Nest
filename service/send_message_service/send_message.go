package send_message_service

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"message-nest/models"
	"message-nest/service/send_task_service"
	"message-nest/service/send_way_service"
	"strings"
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
	MarkDown string

	Status    int
	LogOutput []string
}

// LogsAndStatusMark 记录执行的日志和状态标记
func (sm *SendMessageService) LogsAndStatusMark(errStr string, status int) {
	sm.LogOutput = append(sm.LogOutput, errStr)
	if status == SendFail {
		sm.Status = SendFail
	}
	logrus.Error(fmt.Sprintf("%s, 状态：%d", errStr, status))
}

// Send 发送一个消息任务的所有实例
func (sm *SendMessageService) Send() string {
	sm.Status = SendSuccess
	errStr := ""

	sm.LogsAndStatusMark(fmt.Sprintf("开始任务[%s]的发送", sm.TaskID), sm.Status)
	sendTaskService := send_task_service.SendTaskService{
		ID: sm.TaskID,
	}
	task, err := sendTaskService.GetTaskWithIns()
	sm.LogsAndStatusMark(fmt.Sprintf("任务名称：%s", task.Name), sm.Status)
	sm.LogsAndStatusMark(fmt.Sprintf("发送标题：%s \n", sm.Title), sm.Status)
	if err != nil {
		errStr = fmt.Sprintf("任务[%s]不存在！退出发送！", sm.TaskID)
		sm.LogsAndStatusMark(errStr, SendFail)
		return errStr
	}

	for idx, ins := range task.InsData {
		way, err := models.GetWayByID(ins.WayID)
		if err != nil {
			errStr = fmt.Sprintf("渠道[%s]信息不存在！跳过这个实例的发送", ins.WayID)
			sm.LogsAndStatusMark(errStr, SendFail)
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

		// 邮箱类型的实例发送
		emailAuth, ok := msgObj.(send_way_service.WayDetailEmail)
		if ok {
			//continue
			es := EmailService{}
			errMsg := es.SendTaskEmail(emailAuth, ins.SendTasksIns, typeC, sm.Title, content)
			sm.LogsAndStatusMark(sm.TransError(errMsg), errStrIsSuccess(errMsg))
			continue
		}
		// 钉钉类型的实例发送
		dtalkAuth, ok := msgObj.(send_way_service.WayDetailDTalk)
		if ok {
			es := DtalkService{}
			res, errMsg := es.SendDtalkMessage(dtalkAuth, ins.SendTasksIns, typeC, sm.Title, content)
			sm.LogsAndStatusMark(fmt.Sprintf("返回内容：%s", res), sm.Status)
			sm.LogsAndStatusMark(sm.TransError(errMsg), errStrIsSuccess(errMsg))
			continue
		}
		// 自定义webhook类型的实例发送
		customAuth, ok := msgObj.(send_way_service.WayDetailCustom)
		if ok {
			cs := CustomService{}
			res, errMsg := cs.SendCustomMessage(customAuth, ins.SendTasksIns, typeC, sm.Title, content)
			sm.LogsAndStatusMark(fmt.Sprintf("返回内容：%s", res), sm.Status)
			sm.LogsAndStatusMark(sm.TransError(errMsg), errStrIsSuccess(errMsg))
			continue
		}
		sm.LogsAndStatusMark(fmt.Sprintf("发送失败：未知渠道的发信实例: %s\n", ins.ID), SendFail)

	}

	sm.AppendSendContent()
	sm.RecordSendLog()

	if sm.Status == SendSuccess {
		return ""
	}
	return strings.Join(sm.LogOutput, "\n")
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
	if len(sm.LogOutput) <= 0 {
		return
	}
	log := models.SendTasksLogs{
		Log:    strings.Join(sm.LogOutput, "\n"),
		TaskID: sm.TaskID,
		Status: sm.Status,
	}
	err := log.Add()
	if err != nil {
		logrus.Error(fmt.Sprintf("添加日志失败！原因是：%s", err))
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
	data["text"] = sm.Text
	data["html"] = sm.HTML
	data["markdown"] = sm.MarkDown
	content, ok := data[strings.ToLower(ins.ContentType)]
	if !ok || len(content) == 0 {
		content, ok := data["text"]
		if !ok {
			logrus.Error("text节点数据为空！")
			return "text", ""
		} else {
			logrus.Error(fmt.Sprintf("没有找到%s对应的消息，使用text消息替代！", ins.ContentType))
			return "text", content
		}
	} else {
		return strings.ToLower(ins.ContentType), content
	}
}
