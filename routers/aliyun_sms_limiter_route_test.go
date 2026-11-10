package routers

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"message-nest/pkg/setting"
)

func TestAliyunSMSLimiterRoutesAreRegisteredBehindJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousPrefix := setting.ServerSetting.UrlPrefix
	setting.ServerSetting.UrlPrefix = ""
	t.Cleanup(func() { setting.ServerSetting.UrlPrefix = previousPrefix })

	router := InitRouter(embed.FS{})
	for _, path := range []string{"/api/v1/templates/ins/aliyun/edit", "/api/v1/templates/ins/aliyun/test"} {
		request := httptest.NewRequest(http.MethodPost, path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("route %s should be registered and reject missing JWT, status=%d body=%s", path, response.Code, response.Body.String())
		}
	}
}
