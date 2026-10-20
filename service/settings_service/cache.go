package settings_service

import (
	"sync"
	"time"
)

// SiteConfigCache 站点配置缓存
type SiteConfigCache struct {
	data      map[string]string
	lastUpdate time.Time
	mutex     sync.RWMutex
}

var siteConfigCache = &SiteConfigCache{
	data: make(map[string]string),
}

// GetSiteConfigCache 获取站点配置缓存
func GetSiteConfigCache() map[string]string {
	siteConfigCache.mutex.RLock()
	defer siteConfigCache.mutex.RUnlock()
	
	// 返回缓存的副本，避免外部修改
	result := make(map[string]string)
	for k, v := range siteConfigCache.data {
		result[k] = v
	}
	return result
}

// SetSiteConfigCache 设置站点配置缓存
func SetSiteConfigCache(data map[string]string) {
	siteConfigCache.mutex.Lock()
	defer siteConfigCache.mutex.Unlock()
	
	// 清空现有缓存
	siteConfigCache.data = make(map[string]string)
	
	// 设置新数据
	for k, v := range data {
		siteConfigCache.data[k] = v
	}
	
	// 更新缓存时间
	siteConfigCache.lastUpdate = time.Now()
}

// ClearSiteConfigCache 清除站点配置缓存
func ClearSiteConfigCache() {
	siteConfigCache.mutex.Lock()
	defer siteConfigCache.mutex.Unlock()
	
	siteConfigCache.data = make(map[string]string)
	siteConfigCache.lastUpdate = time.Time{}
}

// IsSiteConfigCacheValid 检查缓存是否有效（支持短期缓存过期，避免数据库直改无法感知）
func IsSiteConfigCacheValid() bool {
	siteConfigCache.mutex.RLock()
	defer siteConfigCache.mutex.RUnlock()

	// 如果缓存为空，认为无效
	if len(siteConfigCache.data) == 0 {
		return false
	}

	// 缓存 10 秒后自动过期，既能抗住高频并发请求，又能及时感知到数据库底层变更
	if time.Since(siteConfigCache.lastUpdate) > 10*time.Second {
		return false
	}

	return true
}
