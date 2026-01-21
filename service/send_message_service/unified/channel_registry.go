package unified

import (
	"fmt"
	"message-nest/service/send_message_service/unified/channels"
	"sync"
)

// 重导出channels包的类型，方便外部使用
type (
	Channel               = channels.Channel
	UnifiedMessageContent = channels.UnifiedMessageContent
)

// 重导出常量
const (
	FormatTypeText     = channels.FormatTypeText
	FormatTypeHTML     = channels.FormatTypeHTML
	FormatTypeMarkdown = channels.FormatTypeMarkdown

	MessageTypeEmail           = channels.MessageTypeEmail
	MessageTypeDtalk           = channels.MessageTypeDtalk
	MessageTypeQyWeiXin        = channels.MessageTypeQyWeiXin
	MessageTypeFeishu          = channels.MessageTypeFeishu
	MessageTypeCustom          = channels.MessageTypeCustom
	MessageTypeWeChatOFAccount = channels.MessageTypeWeChatOFAccount
	MessageTypeMessageNest     = channels.MessageTypeMessageNest
	MessageTypeAliyunSMS       = channels.MessageTypeAliyunSMS
	MessageTypeTelegram        = channels.MessageTypeTelegram
	MessageTypeBark            = channels.MessageTypeBark
)

// ChannelRegistry 渠道注册表
type ChannelRegistry struct {
	channels map[string]Channel
	mu       sync.RWMutex
}

// NewChannelRegistry 创建渠道注册表
func NewChannelRegistry() *ChannelRegistry {
	return &ChannelRegistry{
		channels: make(map[string]Channel),
	}
}

// Register 注册渠道
func (r *ChannelRegistry) Register(channel Channel) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.channels[channel.GetType()] = channel
}

// GetChannel 获取渠道
func (r *ChannelRegistry) GetChannel(channelType string) (Channel, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	channel, ok := r.channels[channelType]
	return channel, ok
}

// GetAllChannels 获取所有渠道
func (r *ChannelRegistry) GetAllChannels() map[string]Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Channel, len(r.channels))
	for k, v := range r.channels {
		result[k] = v
	}
	return result
}

// ListChannels 列出所有渠道
func (r *ChannelRegistry) ListChannels() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.channels))
	for t := range r.channels {
		types = append(types, t)
	}
	return types
}

var (
	globalChannelRegistry *ChannelRegistry
	channelRegistryOnce   sync.Once
)

// GetGlobalChannelRegistry 获取全局渠道注册表（单例）
func GetGlobalChannelRegistry() *ChannelRegistry {
	channelRegistryOnce.Do(func() {
		globalChannelRegistry = NewChannelRegistry()

		// 注册所有渠道
		globalChannelRegistry.Register(channels.NewEmailChannel())
		globalChannelRegistry.Register(channels.NewDtalkChannel())
		globalChannelRegistry.Register(channels.NewQyWeiXinChannel())
		globalChannelRegistry.Register(channels.NewFeishuChannel())
		globalChannelRegistry.Register(channels.NewCustomChannel())
		globalChannelRegistry.Register(channels.NewWeChatOFAccountChannel())
		globalChannelRegistry.Register(channels.NewMessageNestChannel())
		globalChannelRegistry.Register(channels.NewAliyunSMSChannel())
		globalChannelRegistry.Register(channels.NewTelegramChannel())
		globalChannelRegistry.Register(channels.NewBarkChannel())
	})
	return globalChannelRegistry
}

// GetChannelInfo 获取渠道信息（用于调试和文档）
func GetChannelInfo(channelType string) (string, error) {
	registry := GetGlobalChannelRegistry()
	channel, ok := registry.GetChannel(channelType)
	if !ok {
		return "", fmt.Errorf("未知的渠道类型: %s", channelType)
	}

	info := fmt.Sprintf("渠道: %s\n支持格式: %v\n",
		channel.GetType(),
		channel.GetSupportedFormats())
	return info, nil
}

// ListAllChannels 列出所有渠道信息
func ListAllChannels() string {
	registry := GetGlobalChannelRegistry()
	allChannels := registry.GetAllChannels()

	result := "已注册的渠道:\n"
	for channelType, channel := range allChannels {
		result += fmt.Sprintf("  - %s: %v\n", channelType, channel.GetSupportedFormats())
	}
	return result
}
