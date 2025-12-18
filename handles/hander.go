package handles

import (
	"chihqiang/hoststat/token"
	"encoding/json"
	"github.com/chihqiang/logx"
	"net/http"
	"time"
)

func BusinessRoutes() {
	routes := map[string]http.HandlerFunc{
		"/base":       HandlerBase,
		"/current":    HandlerCurrent,
		"/top/cpu/ps": HandlerTopCpuPs,
		"/top/mem/ps": HandlerTopMemPs,
	}
	for path, handler := range routes {
		http.HandleFunc(path, SecureMiddleware(handler))
		logx.Debug("Registered route | path: %s | handler: %T", path, handler)
	}
}

// SecureMiddleware 安全中间件 - 检查Cookie确保只能从页面本身访问
func SecureMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 处理OPTIONS请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		// 使用token包验证令牌
		if err := token.ValidateToken(r); err != nil {
			logx.Warn(
				"[SECURITY] Token validation failed | remote_ip: %s | path: %s | method: %s | error: %v | timestamp: %s",
				r.RemoteAddr,
				r.URL.Path,
				r.Method,
				err,
				time.Now().Format("2006-01-02 15:04:05.000"),
			)
			http.Error(w, "Token validation failed", http.StatusForbidden)
			return
		}
		// 继续处理请求
		next(w, r)
	}
}

func HandlerBase(w http.ResponseWriter, r *http.Request) {
	info, err := getBaseInfo()
	if err != nil {
		logx.Error("get base info error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		logx.Error("Failed to encode current info JSON | remote_ip: %s | error: %v", r.RemoteAddr, err)
		http.Error(w, "Failed to encode response data", http.StatusInternalServerError)
	}
}

func HandlerCurrent(w http.ResponseWriter, r *http.Request) {
	info, err := getCurrentInfo()
	if err != nil {
		logx.Error("Failed to get current info | remote_ip: %s | error: %v", r.RemoteAddr, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		logx.Error("Failed to encode current info JSON | remote_ip: %s | error: %v", r.RemoteAddr, err)
		http.Error(w, "Failed to encode response data", http.StatusInternalServerError)
	}
}
func HandlerTopCpuPs(w http.ResponseWriter, r *http.Request) {
	info := loadTopCPU()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		logx.Error("Failed to encode current info JSON | remote_ip: %s | error: %v", r.RemoteAddr, err)
		http.Error(w, "Failed to encode response data", http.StatusInternalServerError)
	}
}
func HandlerTopMemPs(w http.ResponseWriter, r *http.Request) {
	info := loadTopMem()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		logx.Error("Failed to encode current info JSON | remote_ip: %s | error: %v", r.RemoteAddr, err)
		http.Error(w, "Failed to encode response data", http.StatusInternalServerError)
	}
}
