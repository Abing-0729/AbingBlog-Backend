package handler

import (
	"net/http"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/response"
	"abingblog-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "用户名和密码不能为空")
		return
	}
	token, expiresIn, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"token": token, "expires_in": expiresIn})
}
