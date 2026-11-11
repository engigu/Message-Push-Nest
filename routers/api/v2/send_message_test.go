package v2

import "testing"

func TestSelectDeclaredTemplateParamsExcludesUndeclaredRequestValues(t *testing.T) {
	definitions := []TemplatePlaceholder{{Key: "name"}, {Key: "code"}, {Key: "missing"}}
	values := map[string]interface{}{"name": "Ada", "code": 123, "extra": "not a template placeholder"}
	got := selectDeclaredTemplateParams(definitions, values)
	if len(got) != 2 || got["name"] != "Ada" || got["code"] != 123 {
		t.Fatalf("selected params = %#v, want only declared values", got)
	}
	if _, exists := got["extra"]; exists {
		t.Fatal("undeclared request placeholder leaked into structured SMS params")
	}
}
