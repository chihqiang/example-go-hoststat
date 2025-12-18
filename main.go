package main

import (
	"chihqiang/hoststat/handles"
	"chihqiang/hoststat/logx"
	"chihqiang/hoststat/token"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed index.html favicon.ico
var embedFs embed.FS

var indexTemplate *template.Template

const (
	serverAddr   = ":8080"
	readTimeout  = 15 * time.Second
	writeTimeout = 15 * time.Second
	idleTimeout  = 60 * time.Second
)

// 初始化函数：提前解析模板、校验静态资源，避免运行时错误
func init() {
	indexTemplate = template.Must(template.ParseFS(embedFs, "index.html"))
}

func main() {
	registerRoutes()
	// 2. 配置HTTP服务器（添加超时、优雅关闭）
	server := &http.Server{
		Addr:         serverAddr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
	// 3. 启动服务器（goroutine+优雅关闭）
	logx.Info("HTTP server starting at %s", serverAddr)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// 增强错误日志：包含地址、错误详情、时间戳
			logx.Error(
				"HTTP server startup failed | addr: %s | error: %v | time: %s",
				serverAddr,
				err,
				time.Now().Format("2006-01-02 15:04:05"),
			)
			// 启动失败退出程序
			return
		}
	}()
	// 4. 优雅关闭处理（监听系统信号）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logx.Warn("Shutting down HTTP server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logx.Error("HTTP server forced shutdown | error: %v", err)
	} else {
		logx.Info("HTTP server exited normally")
	}
}

// registerRoutes 统一注册所有HTTP路由，便于管理
func registerRoutes() {
	// 1. favicon.ico路由（完善错误处理）
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// 错误不再忽略，记录日志+返回500
		content, err := embedFs.ReadFile("favicon.ico")
		if err != nil {
			logx.Error("Read favicon.ico failed | error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Header().Set("Cache-Control", "public, max-age=86400") // 增加缓存，优化性能
		if _, err = w.Write(content); err != nil {
			logx.Error("Write favicon.ico response failed | error: %v", err)
		}
	})
	// 2. 根路由（增强错误处理+日志）
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 先执行token设置，捕获可能的错误（如果token.SetToken有返回值的话）
		token.SetToken(w, r)
		// 执行模板，完善错误日志（包含请求上下文）
		if err := indexTemplate.Execute(w, nil); err != nil {
			errMsg := fmt.Sprintf("Execute template failed | path: %s | remote_ip: %s | error: %v", r.URL.Path, r.RemoteAddr, err)
			logx.Error(errMsg)
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	})
	handles.BusinessRoutes()
}
