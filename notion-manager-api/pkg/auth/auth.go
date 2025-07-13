package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Corray333/employee_dashboard/internal/entities"
)

type TelegramCredentials struct {
	id       int64
	username string
}

func (c TelegramCredentials) GetUserID() int64 {
	return c.id
}

func (c TelegramCredentials) GetUsername() string {
	return c.username
}

func CheckTelegramAuth(initData string) (TelegramCredentials, error) {

	parsedData, _ := url.QueryUnescape(initData)
	chunks := strings.Split(parsedData, "&")
	var dataPairs [][]string
	hash := ""
	user := &struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		IsPremium bool   `json:"is_premium"`
	}{}
	// Filter and split the chunks
	for _, chunk := range chunks {
		if strings.HasPrefix(chunk, "user=") {
			parsedData = strings.TrimPrefix(chunk, "user=")
			if err := json.Unmarshal([]byte(parsedData), user); err != nil {
				slog.Error("Failed to unmarshal user data", "error", err)
				return TelegramCredentials{}, err
			}
		}
		if strings.HasPrefix(chunk, "hash=") {
			hash = strings.TrimPrefix(chunk, "hash=")
		} else {
			pair := strings.SplitN(chunk, "=", 2)
			dataPairs = append(dataPairs, pair)
		}
	}

	// Sort the data pairs by the key
	sort.Slice(dataPairs, func(i, j int) bool {
		return dataPairs[i][0] < dataPairs[j][0]
	})

	// Join the sorted data pairs into the initData string
	var sortedData []string
	for _, pair := range dataPairs {
		sortedData = append(sortedData, fmt.Sprintf("%s=%s", pair[0], pair[1]))
	}
	initData = strings.Join(sortedData, "\n")
	// Create the secret key using HMAC and the given token
	h := hmac.New(sha256.New, []byte("WebAppData"))
	h.Write([]byte(os.Getenv("BOT_TOKEN")))
	secretKey := h.Sum(nil)

	// Create the data check using the secret key and initData
	h = hmac.New(sha256.New, secretKey)
	h.Write([]byte(initData))
	dataCheck := h.Sum(nil)

	if fmt.Sprintf("%x", dataCheck) != hash {
		return TelegramCredentials{}, fmt.Errorf("invalid hash")
	}

	return TelegramCredentials{
		id:       user.ID,
		username: user.Username,
	}, nil
}

const (
	AccessTokenLifeTime  = time.Minute * 60
	RefreshTokenLifeTime = time.Hour * 24 * 7
)

func NewTelegramCredentialsMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		slog.Info("auth middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			creds, err := CheckTelegramAuth(authHeader)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				slog.Error("Unauthorized", "error", err)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), entities.ContextKeyUserCredentials, creds))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
