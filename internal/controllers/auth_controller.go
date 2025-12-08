package controllers

import (
	"net/http"
	"os"
	"simdokpol/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service       services.AuthService
	configService services.ConfigService
	currentVersion string
}

func NewAuthController(service services.AuthService, configService services.ConfigService, version string) *AuthController {
	return &AuthController{
		service:       service,
		configService: configService,
		currentVersion: version,
	}
}

type LoginRequest struct {
	NRP      string `json:"nrp" binding:"required" example:"12345"`
	Password string `json:"password" binding:"required" example:"password123"`
}

func (c *AuthController) ShowLoginPage(ctx *gin.Context) {
	isSetup, _ := c.configService.IsSetupComplete()
	if !isSetup {
		ctx.Redirect(http.StatusFound, "/setup")
		return
	}
    // Menggunakan RenderHTML agar version dan changelog masuk
	RenderHTML(ctx, "login.html", gin.H{
		"Title":      "Login Masuk",
	})
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		APIError(ctx, http.StatusBadRequest, "NRP dan Kata Sandi diperlukan")
		return
	}

	token, err := c.service.Login(req.NRP, req.Password)
	if err != nil {
		APIError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	config, _ := c.configService.GetConfig()
	timeoutSeconds := 28800 // Default 8 jam
	if config != nil && config.SessionTimeout > 0 {
		timeoutSeconds = config.SessionTimeout * 60
	}

	isProduction := os.Getenv("APP_ENV") == "production"
	ctx.SetCookie("token", token, timeoutSeconds, "/", "", isProduction, true)

	APIResponse(ctx, http.StatusOK, "Login berhasil", nil)
}

func (c *AuthController) Logout(ctx *gin.Context) {
	isProduction := os.Getenv("APP_ENV") == "production"
	ctx.SetCookie("token", "", -1, "/", "", isProduction, true)
	APIResponse(ctx, http.StatusOK, "Logout berhasil", nil)
}