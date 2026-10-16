# Release Notes (v2.2.9)

## 🚀 重大更新与改进

### 🧹 架构职责规范与渠道配置精简
- **全面移除通知渠道冗余启用开关**：
  - 明确系统分层职责：发信渠道（`SendWays`）纯粹作为推送服务的连接凭据与参数配置，具体的任务启停与调度统一由发信实例（`SendTasksIns`）管理。
  - 彻底移除渠道配置弹窗（`WaysForm.vue`）中冗余且未落库的「启用」Switch 开关，彻底解决配置回显不一致问题（[#90](https://github.com/engigu/Message-Push-Nest/issues/90)）。
  - 清理前端渠道管理列表表头与残留状态字段，保持功能纯粹直观。
- **UI 组件健壮性增强**：
  - 完善基础组件 `Switch.vue` 的双向绑定实现，全面兼容 `v-model`（`modelValue`）与 `checked` / `update:checked` 事件模式。
