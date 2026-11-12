package send_way_service

import (
	"encoding/json"
	"testing"
)

func TestMaskWayAuth(t *testing.T) {
	origAuth := `{
		"server": "smtp.example.com",
		"port": 465,
		"account": "bot@example.com",
		"passwd": "MySuperSecretPassword",
		"from_name": "Alert Bot"
	}`

	masked := MaskWayAuth(origAuth)
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(masked), &data); err != nil {
		t.Fatalf("MaskWayAuth produced invalid JSON: %v", err)
	}

	if data["passwd"] != MaskedSecretPlaceholder {
		t.Fatalf("expected passwd to be %s, got %v", MaskedSecretPlaceholder, data["passwd"])
	}
	if data["server"] != "smtp.example.com" {
		t.Fatalf("expected server to be preserved, got %v", data["server"])
	}
	if data["account"] != "bot@example.com" {
		t.Fatalf("expected account to be preserved, got %v", data["account"])
	}
}

func TestMergeWayAuth(t *testing.T) {
	oldAuth := `{
		"server": "smtp.example.com",
		"port": 465,
		"account": "bot@example.com",
		"passwd": "OriginalSecretPassword123"
	}`

	// 1. 用户编辑时保留了 ****** 占位符，期望自动恢复原密码
	newAuthKeep := `{
		"server": "smtp-new.example.com",
		"port": 587,
		"account": "bot@example.com",
		"passwd": "******"
	}`

	mergedKeep := MergeWayAuth(newAuthKeep, oldAuth)
	var dataKeep map[string]interface{}
	if err := json.Unmarshal([]byte(mergedKeep), &dataKeep); err != nil {
		t.Fatalf("MergeWayAuth produced invalid JSON: %v", err)
	}

	if dataKeep["passwd"] != "OriginalSecretPassword123" {
		t.Fatalf("expected original password restored, got %v", dataKeep["passwd"])
	}
	if dataKeep["server"] != "smtp-new.example.com" {
		t.Fatalf("expected new server to be applied, got %v", dataKeep["server"])
	}

	// 2. 用户显式输入了新密码，期望更新为新密码
	newAuthChange := `{
		"server": "smtp.example.com",
		"port": 465,
		"account": "bot@example.com",
		"passwd": "BrandNewSecretPassword456"
	}`

	mergedChange := MergeWayAuth(newAuthChange, oldAuth)
	var dataChange map[string]interface{}
	if err := json.Unmarshal([]byte(mergedChange), &dataChange); err != nil {
		t.Fatalf("MergeWayAuth produced invalid JSON: %v", err)
	}

	if dataChange["passwd"] != "BrandNewSecretPassword456" {
		t.Fatalf("expected new password to be updated, got %v", dataChange["passwd"])
	}
}

// func TestWayDetailCustomSSRFBlock(t *testing.T) {
// 	customValidator := &WayDetailCustom{}
// 
// 	// 1. 尝试配置内网回环地址
// 	loopbackAuth := `{"webhook":"http://127.0.0.1:8080/hook","header":"","body":""}`
// 	msg, _ := customValidator.Validate(loopbackAuth)
// 	if msg == "" || (!strings.Contains(msg, "SSRF") && !strings.Contains(msg, "回环") && !strings.Contains(msg, "私有")) {
// 		t.Fatalf("expected SSRF block for loopback webhook, got msg: %s", msg)
// 	}
// 
// 	// 2. 尝试配置云厂商元数据地址
// 	metadataAuth := `{"webhook":"http://169.254.169.254/latest/meta-data","header":"","body":""}`
// 	msg, _ = customValidator.Validate(metadataAuth)
// 	if msg == "" || (!strings.Contains(msg, "SSRF") && !strings.Contains(msg, "私有")) {
// 		t.Fatalf("expected SSRF block for metadata webhook, got msg: %s", msg)
// 	}
// 
// 	// 3. 尝试配置内网私有 IP
// 	privateAuth := `{"webhook":"http://192.168.1.100:80/hook","header":"","body":""}`
// 	msg, _ = customValidator.Validate(privateAuth)
// 	if msg == "" || (!strings.Contains(msg, "SSRF") && !strings.Contains(msg, "私有")) {
// 		t.Fatalf("expected SSRF block for private IP webhook, got msg: %s", msg)
// 	}
// }

