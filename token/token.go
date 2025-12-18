package token

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
)

func SetToken(w http.ResponseWriter, r *http.Request) {
	headerByte, _ := json.Marshal(r.Header)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    base64.StdEncoding.EncodeToString(headerByte),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false,
		MaxAge:   86400,
	})
}

// ValidateToken 验证token是否有效
func ValidateToken(r *http.Request) error {
	// 从请求中获取token cookie
	cookie, err := r.Cookie("token")
	if err != nil {
		return err
	}
	tokenValue := cookie.Value
	decoded, err := base64.StdEncoding.DecodeString(tokenValue)
	if err != nil {
		return err
	}
	var tokenHeader http.Header
	if err = json.Unmarshal(decoded, &tokenHeader); err != nil {
		return err
	}
	vHeaders := []string{
		"User-Agent",
		"Accept-Encoding",
		"Host",
	}
	for _, header := range vHeaders {
		if r.Header.Get(header) != tokenHeader.Get(header) {
			return errors.New("token header mismatch: " + header)
		}
	}
	if referer := r.Header.Get("Referer"); referer == "" {
		return errors.New("referer header is missing")
	}
	return nil
}
