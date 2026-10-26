package admin

import (
	"os"
	"regexp"
	"testing"

	"zee-mirror/pkg/i18n"
)

// Every key referenced from auth.go must exist in the catalog; a typo would
// otherwise surface to users as the raw key.
func TestAuthKeysExist(t *testing.T) {
	src, err := os.ReadFile("auth.go")
	if err != nil {
		t.Fatal(err)
	}
	keys := regexp.MustCompile(`i18n\.T\(lang, "([a-z_]+)"`).FindAllStringSubmatch(string(src), -1)
	if len(keys) == 0 {
		t.Fatal("found no i18n keys in auth.go")
	}
	for _, m := range keys {
		if _, ok := i18n.Lookup("vi", m[1]); !ok {
			t.Errorf("auth.go uses missing i18n key %q", m[1])
		}
	}
}

// vi stays byte-identical to the strings the handlers used before i18n.
func TestAuthViStringsUnchanged(t *testing.T) {
	for key, want := range map[string]string{
		"auth_title_access_denied": "TRUY CẬP BỊ TỪ CHỐI",
		"admin_need_owner":         "Chỉ Owner mới có thể dùng lệnh này\\.",
		"admin_need_admin":         "Chỉ Admin/Owner mới có thể dùng lệnh này\\.",
		"admin_unauth_usage":       "⚠️ *Sai định dạng*\n\nCách dùng: `/unauth <user_id>`",
		"admin_unauth_done":        "✅ *Đã xóa người dùng 7 khỏi danh sách được cấp quyền*",
		"admin_setexpire_done":     "Thời hạn của người dùng `7` đã được cập nhật đến *01 Jan 2030* \\(5 ngày\\)\\.",
		"admin_authorize_info":     "👤 *Người dùng:* bob\n🏷️ *ID:* `7`\n🔰 *Vai trò:* `Được cấp quyền`\n♾️ *Hạn mức:* `q`",
	} {
		var got string
		switch key {
		case "admin_unauth_done":
			got = i18n.T("vi", key, 7)
		case "admin_setexpire_done":
			got = i18n.T("vi", key, 7, "01 Jan 2030", 5)
		case "admin_authorize_info":
			got = i18n.T("vi", key, "bob", 7, "q")
		default:
			got = i18n.T("vi", key)
		}
		if got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}
