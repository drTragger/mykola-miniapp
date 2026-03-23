package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"

	"github.com/drTragger/mykola-miniapp/internal/config"
)

type telegramInitDataUser struct {
	ID int64 `json:"id"`
}

func telegramAuthMiddleware(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		initData := r.Header.Get("X-Telegram-Init-Data")
		if initData == "" {
			http.Error(w, "missing telegram init data", http.StatusUnauthorized)
			return
		}

		userID, ok := validateTelegramInitData(initData, cfg.Telegram.Token)
		if !ok {
			http.Error(w, "invalid telegram init data", http.StatusUnauthorized)
			return
		}

		if len(cfg.Telegram.AdminIDs) > 0 && !slices.Contains(cfg.Telegram.AdminIDs, userID) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validateTelegramInitData(initDataRaw string, botToken string) (int64, bool) {
	values, err := url.ParseQuery(initDataRaw)
	if err != nil {
		return 0, false
	}

	hash := values.Get("hash")
	if hash == "" {
		return 0, false
	}

	values.Del("hash")

	dataCheckParts := make([]string, 0, len(values))
	for key, vals := range values {
		if len(vals) == 0 {
			continue
		}
		dataCheckParts = append(dataCheckParts, key+"="+vals[0])
	}
	sort.Strings(dataCheckParts)

	dataCheckString := strings.Join(dataCheckParts, "\n")

	secretOuter := hmac.New(sha256.New, []byte("WebAppData"))
	secretOuter.Write([]byte(botToken))
	secretKey := secretOuter.Sum(nil)

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedHash), []byte(hash)) {
		return 0, false
	}

	userRaw := values.Get("user")
	if userRaw == "" {
		return 0, false
	}

	var user telegramInitDataUser
	if err := json.Unmarshal([]byte(userRaw), &user); err != nil {
		return 0, false
	}

	if user.ID == 0 {
		return 0, false
	}

	return user.ID, true
}
