package testcase

import (
	"message-nest/models"
	"testing"
)

// TestAllModelsCRUDConsistency 验证所有数据表模型的增删改查函数与历史行为完全一致
func TestAllModelsCRUDConsistency(t *testing.T) {
	models.InitTestDB("crud_consistency_test")

	// ==========================================
	// 1. 发信渠道 (SendWays) 增删改查一致性
	// ==========================================
	t.Run("SendWays_CRUD", func(t *testing.T) {
		wayName := "Consistency_Way_Email"
		authJSON := `{"smtp_host":"smtp.test.com","smtp_port":465}`
		wayType := "email"
		creator := "UR_TEST_CREATOR_01"

		// 1.1 增 (AddSendWay)
		err := models.AddSendWay(wayName, authJSON, wayType, creator, creator)
		if err != nil {
			t.Fatalf("AddSendWay failed: %v", err)
		}

		// 1.2 查 (GetWayByName)
		wayByName, err := models.GetWayByName(wayName)
		if err != nil || wayByName.ID == "" {
			t.Fatalf("GetWayByName failed: %v", err)
		}
		if wayByName.Name != wayName || wayByName.Type != wayType || wayByName.CreatedBy != creator {
			t.Fatalf("GetWayByName field mismatch: %+v", wayByName)
		}
		wayID := wayByName.ID

		// 1.3 查 (GetWayByID)
		wayByID, err := models.GetWayByID(wayID)
		if err != nil || wayByID.ID != wayID {
			t.Fatalf("GetWayByID failed: %v", err)
		}

		// 1.4 查列表与总数 (GetSendWays & GetSendWaysTotal)
		maps := map[string]interface{}{"created_by": creator}
		waysList, err := models.GetSendWays(0, 10, "Consistency_Way", wayType, maps)
		if err != nil || len(waysList) != 1 {
			t.Fatalf("GetSendWays failed: got %d, err: %v", len(waysList), err)
		}
		total, err := models.GetSendWaysTotal("Consistency_Way", wayType, maps)
		if err != nil || total != 1 {
			t.Fatalf("GetSendWaysTotal expected 1, got %d, err: %v", total, err)
		}

		// 1.5 改 (EditSendWay)
		newName := "Consistency_Way_Email_Renamed"
		err = models.EditSendWay(wayID, map[string]interface{}{
			"name":        newName,
			"modified_by": creator,
		})
		if err != nil {
			t.Fatalf("EditSendWay failed: %v", err)
		}
		wayUpdated, _ := models.GetWayByID(wayID)
		if wayUpdated.Name != newName {
			t.Fatalf("EditSendWay did not update name, got: %s", wayUpdated.Name)
		}

		// 1.6 删 (DeleteMsgWay)
		err = models.DeleteMsgWay(wayID)
		if err != nil {
			t.Fatalf("DeleteMsgWay failed: %v", err)
		}
		wayDeleted, _ := models.GetWayByID(wayID)
		if wayDeleted.ID != "" {
			t.Fatalf("DeleteMsgWay failed: record still exists")
		}
	})

	// ==========================================
	// 2. 发信任务与实例 (SendTasks & SendTasksIns) 增删改查及关联查询一致性
	// ==========================================
	t.Run("SendTasks_And_Ins_CRUD", func(t *testing.T) {
		taskName := "Consistency_Task_01"
		creator := "UR_TEST_CREATOR_02"

		// 2.1 增任务 (AddSendTask)
		err := models.AddSendTask(taskName, creator)
		if err != nil {
			t.Fatalf("AddSendTask failed: %v", err)
		}

		// 查任务列表
		tasks, err := models.GetSendTasks(0, 10, taskName, map[string]interface{}{})
		if err != nil || len(tasks) == 0 {
			t.Fatalf("GetSendTasks failed: %v", err)
		}
		taskID := tasks[0].ID

		// 2.2 查单条 (GetTaskByID)
		taskObj, err := models.GetTaskByID(taskID)
		if err != nil || taskObj.ID != taskID || taskObj.Name != taskName {
			t.Fatalf("GetTaskByID failed: %+v, err: %v", taskObj, err)
		}

		// 2.3 增指定 ID 的任务 (AddSendTaskWithID)
		customID := "TK_CUSTOM_999"
		err = models.AddSendTaskWithID("Custom Task", customID, creator)
		if err != nil {
			t.Fatalf("AddSendTaskWithID failed: %v", err)
		}
		customTask, err := models.GetTaskByID(customID)
		if err != nil || customTask.ID != customID {
			t.Fatalf("GetTaskByID customID failed: %v", err)
		}

		// 2.4 改任务 (EditSendTask)
		err = models.EditSendTask(taskID, map[string]interface{}{"name": "Consistency_Task_Edited"})
		if err != nil {
			t.Fatalf("EditSendTask failed: %v", err)
		}
		taskObjUpdated, _ := models.GetTaskByID(taskID)
		if taskObjUpdated.Name != "Consistency_Task_Edited" {
			t.Fatalf("EditSendTask failed, got name: %s", taskObjUpdated.Name)
		}

		// 2.5 关联实例测试：添加渠道并在任务下挂载实例
		wayName := "Way_For_Ins_Test"
		_ = models.AddSendWay(wayName, `{}`, "webhook", creator, creator)
		wayObj, _ := models.GetWayByName(wayName)

		insItem := models.SendTasksIns{
			UUIDModel: models.UUIDModel{
				ID:         "INS_TEST_01",
				CreatedBy:  creator,
				ModifiedBy: creator,
			},
			TaskID:  taskID,
			WayID:   wayObj.ID,
			WayType: "webhook",
		}
		err = models.AddTaskInsOne(insItem)
		if err != nil {
			t.Fatalf("AddTaskInsOne failed: %v", err)
		}

		// 验证 GetTasksIns (核心：必须通过 JOIN 渠道表查出 way_name)
		taskInsResult, err := models.GetTasksIns(taskID)
		if err != nil {
			t.Fatalf("GetTasksIns failed: %v", err)
		}
		if taskInsResult.ID != taskID || len(taskInsResult.InsData) != 1 {
			t.Fatalf("GetTasksIns result mismatch: %+v", taskInsResult)
		}
		if taskInsResult.InsData[0].WayName != wayName {
			t.Fatalf("GetTasksIns way_name JOIN mismatch: expected %s, got %s", wayName, taskInsResult.InsData[0].WayName)
		}

		// 验证 FindTaskByWayId
		relatedTasks := models.FindTaskByWayId(wayObj.ID)
		if len(relatedTasks) != 1 || relatedTasks[0].ID != taskID {
			t.Fatalf("FindTaskByWayId failed: %+v", relatedTasks)
		}

		// 2.6 删任务与实例联动 (DeleteMsgTask)
		err = models.DeleteMsgTask(taskID)
		if err != nil {
			t.Fatalf("DeleteMsgTask failed: %v", err)
		}
		// 校验任务与实例均被清理
		_, err = models.GetTaskByID(taskID)
		if err == nil {
			t.Fatalf("Task %s should have been deleted", taskID)
		}
	})

	// ==========================================
	// 3. 消息模板 (Template) 增删改查一致性
	// ==========================================
	t.Run("Template_CRUD", func(t *testing.T) {
		name := "Tmpl_Test"
		textTmpl := "Hello {{name}}"
		htmlTmpl := "<b>Hello {{name}}</b>"
		creator := "UR_TEST_CREATOR_03"
		tmplID := models.GenerateTemplateUniqueID()

		tmpl := models.Template{
			UUIDModel: models.UUIDModel{
				ID:         tmplID,
				CreatedBy:  creator,
				ModifiedBy: creator,
			},
			Name:         name,
			TextTemplate: textTmpl,
			HTMLTemplate: htmlTmpl,
			Status:       "enabled",
		}

		// 3.1 增 (Template.Add)
		err := tmpl.Add()
		if err != nil {
			t.Fatalf("Template.Add failed: %v", err)
		}

		// 3.2 查单条 (GetTemplateByID)
		tmplObj, err := models.GetTemplateByID(tmplID)
		if err != nil || tmplObj.ID != tmplID || tmplObj.TextTemplate != textTmpl {
			t.Fatalf("GetTemplateByID failed: %+v, err: %v", tmplObj, err)
		}

		// 3.3 检查是否存在 (ExistTemplateByID)
		exists, err := models.ExistTemplateByID(tmplID)
		if err != nil || !exists {
			t.Fatalf("ExistTemplateByID failed: %v, exists: %v", err, exists)
		}

		// 3.4 查列表与总数 (GetTemplates & GetTemplatesTotal)
		maps := map[string]interface{}{"created_by": creator}
		list, err := models.GetTemplates(0, 10, "Tmpl", maps)
		if err != nil || len(list) != 1 {
			t.Fatalf("GetTemplates failed: got %d, err: %v", len(list), err)
		}
		total, err := models.GetTemplatesTotal("Tmpl", maps)
		if err != nil || total != 1 {
			t.Fatalf("GetTemplatesTotal failed: %d", total)
		}

		// 3.5 改 (Template.Update)
		newText := "Updated {{name}}"
		tmpl.TextTemplate = newText
		err = tmpl.Update()
		if err != nil {
			t.Fatalf("Template.Update failed: %v", err)
		}
		tmplUpdated, _ := models.GetTemplateByID(tmplID)
		if tmplUpdated.TextTemplate != newText {
			t.Fatalf("Template.Update did not update: got %s", tmplUpdated.TextTemplate)
		}

		// 3.6 删 (Template.Delete)
		err = tmpl.Delete()
		if err != nil {
			t.Fatalf("Template.Delete failed: %v", err)
		}
		existsAfterDelete, _ := models.ExistTemplateByID(tmplID)
		if existsAfterDelete {
			t.Fatalf("Template.Delete failed: record still exists")
		}
	})

	// ==========================================
	// 4. 定时消息 (CronMessages) 增删改查一致性
	// ==========================================
	t.Run("CronMessages_CRUD", func(t *testing.T) {
		name := "Cron_Test"
		taskID := "TK_CRON_01"
		cronExpr := "0 10 * * *"
		title := "Test Title"
		content := "Test Content"
		url := "https://example.com"
		creator := "UR_TEST_CREATOR_04"

		// 4.1 增 (AddSendCronMsg)
		cronID, err := models.AddSendCronMsg(name, taskID, cronExpr, title, content, url, creator)
		if err != nil || cronID == "" {
			t.Fatalf("AddSendCronMsg failed: %v", err)
		}

		// 4.2 查单条 (GetCronMsgByID)
		cronObj, err := models.GetCronMsgByID(cronID)
		if err != nil || cronObj.ID != cronID || cronObj.Title != title {
			t.Fatalf("GetCronMsgByID failed: %+v, err: %v", cronObj, err)
		}

		// 4.3 查列表与总数 (GetCronMessages & GetCronMessagesTotal)
		maps := map[string]interface{}{"created_by": creator}
		list, err := models.GetCronMessages(0, 10, "Cron", maps)
		if err != nil || len(list) != 1 {
			t.Fatalf("GetCronMessages failed: got %d, err: %v", len(list), err)
		}
		total, err := models.GetCronMessagesTotal("Cron", maps)
		if err != nil || total != 1 {
			t.Fatalf("GetCronMessagesTotal failed: %d", total)
		}

		// 4.4 改 (EditCronMsg)
		newTitle := "New Cron Title"
		err = models.EditCronMsg(cronID, map[string]interface{}{"title": newTitle})
		if err != nil {
			t.Fatalf("EditCronMsg failed: %v", err)
		}
		cronUpdated, _ := models.GetCronMsgByID(cronID)
		if cronUpdated.Title != newTitle {
			t.Fatalf("EditCronMsg failed: got %s", cronUpdated.Title)
		}

		// 4.5 删 (DeleteCronMsg)
		err = models.DeleteCronMsg(cronID)
		if err != nil {
			t.Fatalf("DeleteCronMsg failed: %v", err)
		}
		_, err = models.GetCronMsgByID(cronID)
		if err == nil {
			t.Fatalf("DeleteCronMsg failed: record still exists")
		}
	})

	// ==========================================
	// 5. 发信日志 (SendTasksLogs) 增查一致性
	// ==========================================
	t.Run("SendTasksLogs_CRUD", func(t *testing.T) {
		creator := "UR_TEST_CREATOR_05"
		logItem := models.SendTasksLogs{
			TaskID:    "TK_LOG_01",
			Type:      "task",
			Name:      "Log Task",
			Log:       "Log content message",
			CreatedBy: creator,
		}

		// 5.1 增 (Add)
		err := logItem.Add()
		if err != nil || logItem.ID == 0 {
			t.Fatalf("SendTasksLogs.Add failed: %v", err)
		}

		// 5.2 查列表与总数 (GetSendLogs & GetSendLogsTotal)
		maps := map[string]interface{}{"created_by": creator}
		logsList, err := models.GetSendLogs(0, 10, "Log Task", "TK_LOG_01", maps)
		if err != nil || len(logsList) != 1 {
			t.Fatalf("GetSendLogs failed: got %d, err: %v", len(logsList), err)
		}
		if logsList[0].Log != "Log content message" {
			t.Fatalf("Log content mismatch: %s", logsList[0].Log)
		}

		total, err := models.GetSendLogsTotal("Log Task", "TK_LOG_01", maps)
		if err != nil || total != 1 {
			t.Fatalf("GetSendLogsTotal failed: %d", total)
		}
	})

	// ==========================================
	// 6. 托管消息 (HostedMessage) 增查与删除一致性
	// ==========================================
	t.Run("HostedMessage_CRUD", func(t *testing.T) {
		creator := "UR_TEST_CREATOR_06"
		hostMsg := models.HostedMessage{
			Title:     "Hosted Notice",
			Content:   "Hosted Content Here",
			CreatedBy: creator,
		}

		// 6.1 增 (Add)
		err := hostMsg.Add()
		if err != nil || hostMsg.ID == 0 || hostMsg.UniqueKey == "" {
			t.Fatalf("HostedMessage.Add failed: %v", err)
		}
		uniqueKey := hostMsg.UniqueKey

		// 6.2 查单条 (GetHostMessageByUniqueKey)
		obj, err := models.GetHostMessageByUniqueKey(uniqueKey)
		if err != nil || obj.UniqueKey != uniqueKey || obj.Title != "Hosted Notice" {
			t.Fatalf("GetHostMessageByUniqueKey failed: %+v, err: %v", obj, err)
		}

		// 6.3 查列表与总数 (GetHostMessages & GetHostMessagesTotal)
		maps := map[string]interface{}{"created_by": creator}
		list, err := models.GetHostMessages(0, 10, "Hosted", maps)
		if err != nil || len(list) != 1 {
			t.Fatalf("GetHostMessages failed: got %d, err: %v", len(list), err)
		}
		total, err := models.GetHostMessagesTotal("Hosted", maps)
		if err != nil || total != 1 {
			t.Fatalf("GetHostMessagesTotal failed: %d", total)
		}

		// 6.4 删 (删除指定托管消息)
		err = models.GetDB().Where("id = ?", hostMsg.ID).Delete(&models.HostedMessage{}).Error
		if err != nil {
			t.Fatalf("Delete hosted message failed: %v", err)
		}
		deletedObj, _ := models.GetHostMessageByUniqueKey(uniqueKey)
		if deletedObj != nil && deletedObj.ID != 0 {
			t.Fatalf("Deleted hosted message still exists: %+v", deletedObj)
		}
	})

	// ==========================================
	// 7. 发送统计 (SendStats) 累加与聚合查询一致性
	// ==========================================
	t.Run("SendStats_CRUD", func(t *testing.T) {
		taskID := "TK_STAT_01"
		taskType := "task"
		day := "2026-10-20"
		creator := "UR_TEST_CREATOR_07"

		// 7.1 首次增加统计
		err := models.IncrementSendStats(taskID, taskType, day, "success", 5, creator)
		if err != nil {
			t.Fatalf("IncrementSendStats insert failed: %v", err)
		}

		// 7.2 再次增加统计（验证 Upsert 累加 num + 3 = 8）
		err = models.IncrementSendStats(taskID, taskType, day, "success", 3, creator)
		if err != nil {
			t.Fatalf("IncrementSendStats upsert failed: %v", err)
		}

		// 增加一条失败统计
		_ = models.IncrementSendStats(taskID, taskType, day, "failed", 2, creator)

		// 7.3 验证 GetSendStatsData 聚合
		statsData, err := models.GetSendStatsData(30, creator)
		if err != nil {
			t.Fatalf("GetSendStatsData failed: %v", err)
		}
		if statsData.TotalSuccNum != 8 {
			t.Fatalf("Expected TotalSuccNum = 8, got %d", statsData.TotalSuccNum)
		}
		if statsData.TotalFailedNum != 2 {
			t.Fatalf("Expected TotalFailedNum = 2, got %d", statsData.TotalFailedNum)
		}
		if statsData.TotalNum != 10 {
			t.Fatalf("Expected TotalNum = 10, got %d", statsData.TotalNum)
		}

		// 7.4 验证 GetSendStatsByTask 任务指定查询
		taskStats, err := models.GetSendStatsByTask(taskID, 30, creator)
		if err != nil {
			t.Fatalf("GetSendStatsByTask failed: %v", err)
		}
		if taskStats.TotalNum != 10 || taskStats.TotalSuccNum != 8 {
			t.Fatalf("GetSendStatsByTask result mismatch: %+v", taskStats)
		}
	})

	// ==========================================
	// 8. 系统与用户配置 (Settings) 读写与回退一致性
	// ==========================================
	t.Run("Settings_CRUD", func(t *testing.T) {
		adminID := "UR_ADMIN_GLOBAL"
		userID := "UR_USER_PRIVATE"

		// 8.1 写入全局配置 (AddOneSetting)
		globalSetting := models.Settings{
			Section:    "site_config",
			Key:        "app_title",
			Value:      "MessageNest Global",
			CreatedBy:  adminID,
			ModifiedBy: adminID,
		}
		err := models.AddOneSetting(globalSetting)
		if err != nil {
			t.Fatalf("AddOneSetting global failed: %v", err)
		}

		// 8.2 用户未设置时，GetSettingByKey 回退读取全局配置
		settingFallback, err := models.GetSettingByKey("site_config", "app_title", userID)
		if err != nil || settingFallback.Value != "MessageNest Global" {
			t.Fatalf("GetSettingByKey fallback failed: %v, got: %s", err, settingFallback.Value)
		}

		// 8.3 用户写入自己的私有配置
		userSetting := models.Settings{
			Section:    "site_config",
			Key:        "app_title",
			Value:      "My Private Nest",
			CreatedBy:  userID,
			ModifiedBy: userID,
		}
		err = models.AddOneSetting(userSetting)
		if err != nil {
			t.Fatalf("AddOneSetting user private failed: %v", err)
		}

		// 用户再次读取，优先获得自己的私有配置
		settingPrivate, err := models.GetSettingByKey("site_config", "app_title", userID)
		if err != nil || settingPrivate.Value != "My Private Nest" {
			t.Fatalf("GetSettingByKey private priority failed, got: %s", settingPrivate.Value)
		}

		// 8.4 验证 GetSettingByUser 精确查询
		exactUserSetting, err := models.GetSettingByUser("site_config", "app_title", userID)
		if err != nil || exactUserSetting.Value != "My Private Nest" {
			t.Fatalf("GetSettingByUser failed: %+v, err: %v", exactUserSetting, err)
		}

		// 8.5 验证 EditSetting 修改
		err = models.EditSetting(exactUserSetting.ID, map[string]interface{}{"value": "My Private Nest Updated"})
		if err != nil {
			t.Fatalf("EditSetting failed: %v", err)
		}
		settingEdited, _ := models.GetSettingByUser("site_config", "app_title", userID)
		if settingEdited.Value != "My Private Nest Updated" {
			t.Fatalf("EditSetting did not update value: %s", settingEdited.Value)
		}

		// 8.6 验证 DeleteSettingByKey 删除
		err = models.DeleteSettingByKey("site_config", "app_title", userID)
		if err != nil {
			t.Fatalf("DeleteSettingByKey failed: %v", err)
		}
		// 删除用户私有配置后，用户再次读取又回退到全局配置
		settingAfterDelete, err := models.GetSettingByKey("site_config", "app_title", userID)
		if err != nil || settingAfterDelete.Value != "MessageNest Global" {
			t.Fatalf("Expected fallback to global after delete, got: %+v", settingAfterDelete)
		}
	})

	// ==========================================
	// 9. 登录日志 (LoginLog) 增查一致性
	// ==========================================
	t.Run("LoginLog_CRUD", func(t *testing.T) {
		userID := "UR_USER_LOGIN_01"
		username := "test_user_log"
		ip := "127.0.0.1"
		ua := "Mozilla/5.0 Test"

		// 9.1 增 (AddLoginLog)
		err := models.AddLoginLog(userID, username, ip, ua)
		if err != nil {
			t.Fatalf("AddLoginLog failed: %v", err)
		}

		// 9.2 查 (GetRecentLoginLogs)
		recentLogs, err := models.GetRecentLoginLogs(10)
		if err != nil || len(recentLogs) == 0 {
			t.Fatalf("GetRecentLoginLogs failed: %v, count: %d", err, len(recentLogs))
		}
		found := false
		for _, l := range recentLogs {
			if l.UserID == userID && l.Username == username && l.IP == ip {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Added login log not found in GetRecentLoginLogs: %+v", recentLogs)
		}
	})
}
