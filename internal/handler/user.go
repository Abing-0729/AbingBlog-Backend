package handler

import (
	"net/http"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/response"
	"abingblog-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler 登录接口层
type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Login 登录：校验账号密码，成功返回 JWT token。
// 参数绑定这段是样板；调 svc.Login 拿 token 的链路你已在 service 里实现。
func (h *UserHandler) Login(c *gin.Context) {
	var req service.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	token, err := h.svc.Login(req)
	if err != nil {
		response.Error(c, err) // CodeLoginFail 已映射为 401
		return
	}
	response.OK(c, gin.H{"token": token})
}
