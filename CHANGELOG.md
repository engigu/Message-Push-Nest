# Release Notes (v2.2.9)

## 🚀 重大更新与改进

### 🧹 架构职责规范与渠道配置精简
- **全面移除通知渠道冗余启用开关**：
  - 明确系统分层职责：发信渠道（`SendWays`）纯粹作为推送服务的连接凭据与参数配置，具体的任务启停与调度统一由发信实例（`SendTasksIns`）管理。
  - 彻底移除渠道配置弹窗（`WaysForm.vue`）中冗余且未落库的「启用」Switch 开关，彻底解决配置回显不一致问题（[#90](https://github.com/engigu/Message-Push-Nest/issues/90)）。
  - 清理前端渠道管理列表表头与残留状态字段，保持功能纯粹直观。
- **UI 组件健壮性增强**：
  - 完善基础组件 `Switch.vue` 的双向绑定实现，全面兼容 `v-model`（`modelValue`）与 `checked` / `update:checked` 事件模式。

### ⚡ 弹窗交互与硬件加速动效优化
- **Dialog 硬件加速与动效重构**：
  - 彻底移除大尺寸弹窗在大面积重绘时的 scale 缩放开销，改为 120ms GPU 硬件加速微动效（`translate3d` 纯透明度渐变 + 4px 微浮动），消除大尺寸弹窗打开时的卡顿、抖动与位移冲突。
  - 升级 Dialog 遮罩层视觉质感，由原本沉重的纯黑优化为通透现代的 `bg-black/50 backdrop-blur-[2px]`，带来毫秒级丝滑秒开体验。
