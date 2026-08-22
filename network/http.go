package network

import (
	"fmt"
	"time"
	"strconv"
	"context"
	"net/http"
	"encoding/json"

	"github.com/wengerbinning/srv/log"
	"github.com/wengerbinning/srv/config"
	"github.com/wengerbinning/srv/service"
	"github.com/wengerbinning/srv/database"
)

const (
	// HTTP 接口默认超时
	defReadTimeout  = 15 * time.Second
	defWriteTimeout = 15 * time.Second
)

// UserView 是 HTTP 接口对外暴露的用户数据传输对象（DTO），
// 不直接泄漏内部 User 结构/密码策略等实现细节。
type UserView struct {
	Uuid     int64  `json:"uuid"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Server 封装 HTTP 监听服务的配置与运行时状态。
type Server struct {
	conf  *config.Conf
	log   log.Interface
	http  *http.Server
	srv   *service.Context
	users *service.UserContext
}

// Addr 返回 "listen:port" 形式的监听地址。
func (s *Server) Addr() string {
	listen := s.conf.Network.Http.Listen
	if listen == "" {
		listen = "0.0.0.0"
	}
	port := s.conf.Network.Http.Port
	if port == "" {
		port = "8080"
	}
	return listen + ":" + port
}

// parseTimeout 解析配置中的超时时长字符串。
func (s *Server) parseTimeout(raw string, def time.Duration) time.Duration {
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		s.log.Warning("network: invalid timeout %q, use default %v", raw, def)
		return def
	}
	return d
}

// writeJSON 输出 JSON 响应。
func (s *Server) writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.log.Error("network: encode response: %v", err)
	}
}

// writeError 输出统一的错误响应体。
func (s *Server) writeError(w http.ResponseWriter, code int, err error) {
	s.writeJSON(w, code, map[string]string{
		"error": err.Error(),
	})
}

// parseUuid 从 URL 路径中解析用户 uuid。
func (s *Server) parseUuid(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest,
			fmt.Errorf("invalid user uuid: %q", r.PathValue("id")))
		return 0, false
	}
	return id, true
}

// ---- User HTTP handlers ----

// getUsers 处理 GET /api/v1/users，支持按查询参数搜索，全部参数为 * 时返回所有用户。
// 查询参数: username / email / password。
func (s *Server) getUsers(w http.ResponseWriter, r *http.Request) {
	param := database.NewUserParam(
		r.URL.Query().Get("username"),
		r.URL.Query().Get("password"),
		r.URL.Query().Get("email"),
	)
	users, err := s.users.Search(param)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	views := make([]UserView, 0, len(users))
	for _, u := range users {
		views = append(views, UserView{
			Uuid:     u.Uuid,
			Username: u.Username,
			Email:    u.Email,
		})
	}
	s.writeJSON(w, http.StatusOK, views)
}

// getUser 处理 GET /api/v1/users/{id}。
func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	id, ok := s.parseUuid(w, r)
	if !ok {
		return
	}
	user, err := s.users.Fetch(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}
	s.writeJSON(w, http.StatusOK,
		UserView{Uuid: user.Uuid, Username: user.Username, Email: user.Email})
}

// postUser 处理 POST /api/v1/users，创建用户。
func (s *Server) postUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest,
			fmt.Errorf("invalid request body: %w", err))
		return
	}
	user, err := s.users.Registrate(req.Username, req.Email, req.Password)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	s.writeJSON(w, http.StatusCreated,
		UserView{Uuid: user.Uuid, Username: user.Username, Email: user.Email})
}

// putUser 处理 PUT /api/v1/users/{id}，更新用户。
func (s *Server) putUser(w http.ResponseWriter, r *http.Request) {
	id, ok := s.parseUuid(w, r)
	if !ok {
		return
	}
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest,
			fmt.Errorf("invalid request body: %w", err))
		return
	}
	user, err := s.users.Changerate(id, req.Username, req.Email, req.Password)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	s.writeJSON(w, http.StatusOK,
		UserView{Uuid: user.Uuid, Username: user.Username, Email: user.Email})
}

// deleteUser 处理 DELETE /api/v1/users/{id}。
func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	id, ok := s.parseUuid(w, r)
	if !ok {
		return
	}
	if err := s.users.Cancellate(id); err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ---- Server lifecycle ----

// newMux 构建路由。
func (s *Server) newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/users", s.getUsers)
	mux.HandleFunc("POST /api/v1/users", s.postUser)
	mux.HandleFunc("GET /api/v1/users/{id}", s.getUser)
	mux.HandleFunc("PUT /api/v1/users/{id}", s.putUser)
	mux.HandleFunc("DELETE /api/v1/users/{id}", s.deleteUser)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.writeError(w, http.StatusNotFound, fmt.Errorf("not found: %s", r.URL.Path))
	})
	return mux
}

// CreateServer 创建一个 HTTP 监听服务器。
func CreateServer(conf *config.Conf, srv *service.Context) (*Server, error) {
	if conf == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if srv == nil {
		return nil, fmt.Errorf("service context is nil")
	}
	return &Server{
		conf:  conf,
		log:   srv.LogIface(),
		srv:   srv,
		users: srv.CreateUserContext(),
		http:  &http.Server{},
	}, nil
}

// Listen 开始监听并服务 HTTP 请求，阻塞直到服务器关闭。
func (s *Server) Listen() error {
	if s.http == nil {
		return fmt.Errorf("http server is nil")
	}
	s.http.Addr = s.Addr()
	s.http.Handler = s.newMux()
	s.http.ReadTimeout = s.parseTimeout(s.conf.Network.Http.ReadTimeout, defReadTimeout)
	s.http.WriteTimeout = s.parseTimeout(s.conf.Network.Http.WriteTimeout, defWriteTimeout)
	s.log.Message("network: listening on %s", s.Addr())
	err := s.http.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http listen failed: %w", err)
	}
	return nil
}

// Serve 在后台异步监听，返回立即。
func (s *Server) Serve() {
	go func() {
		if err := s.Listen(); err != nil {
			s.log.Error("network: %v", err)
		}
	}()
}

// Shutdown 优雅关闭 HTTP 服务器。
func (s *Server) Shutdown() error {
	if s.http == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), defWriteTimeout)
	defer cancel()
	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("http shutdown failed: %w", err)
	}
	s.log.Message("network: shutdown finished")
	return nil
}
