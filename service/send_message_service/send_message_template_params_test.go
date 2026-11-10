package send_message_service

import (
	"reflect"
	"testing"

	"message-nest/models"
)

func TestBuildTemplateContentCarriesTemplateParams(t *testing.T) {
	params := map[string]interface{}{"name": "Ada", "count": 2}
	service := &SendMessageService{Text: "hello Ada", TemplateParams: params}
	content := service.BuildTemplateContent(models.SendTasksIns{ContentType: "text"})
	if content == nil || content.Text != "hello Ada" {
		t.Fatalf("unexpected template content: %#v", content)
	}
	if !reflect.DeepEqual(content.TemplateParams, params) {
		t.Fatalf("TemplateParams = %#v, want %#v", content.TemplateParams, params)
	}
	if content.Extra != nil {
		t.Fatalf("structured template params must not be coupled to Extra: %#v", content.Extra)
	}
}
