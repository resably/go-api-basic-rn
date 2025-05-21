package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/resably/go-rest-api/models"
	"github.com/resably/go-rest-api/utils"

	"gorm.io/gorm"
)

type AuthController struct {
	DB     *gorm.DB
	Hasher utils.PasswordHasher
}

func NewAuthController(db *gorm.DB, hasher utils.PasswordHasher) *AuthController {
	return &AuthController{
		DB:     db,
		Hasher: hasher,
	}
}

// register func
func (ac *AuthController) Register(c *gin.Context) {
	var input models.RegisterRequest

	// gelen veriyi doğrula
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Geçersiz istek formatı",
		})
		return
	}

	// kullanıcı var mı kontrol et
	var existingUser models.User
	if err := ac.DB.Where("email = ? OR username = ?", input.Email, input.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Kullanıcı zaten mevcut",
			"conflicts": map[string]string{
				"email":    "Bu email kullanımda",
				"username": "Bu kullanıcı adı alınmış",
			},
		})
		return
	}

	// şifreyi hashle
	hashedPassword, err := ac.Hasher.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Şifre işlenirken hata oluştu",
		})
		return
	}

	// transaction başlat
	tx := ac.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// kullanıcıyı oluştur
	user := models.User{
		Username:  input.Username,
		Email:     input.Email,
		Password:  hashedPassword,
		Name:      input.Name,
		Surname:   input.Surname,
		CreatedAt: time.Now(),
	}

	// kullanıcıyı veritabanına ekle
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Kullanıcı oluşturulamadı",
		})
		return
	}

	// token oluştur
	accessToken, refreshToken, err := utils.GenerateTokens(&user)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Token oluşturulamadı",
		})
		return
	}

	// refresh token'ı veritabanına kaydet
	if err := utils.SaveRefreshToken(tx, user.ID, refreshToken); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Oturum bilgisi kaydedilemedi",
		})
		return
	}

	// transaction'ı onayla
	tx.Commit()

	// kullanıcıyı ve token'ları döndür
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"name":     user.Name,
				"surname":  user.Surname,
			},
		},
	})
}

// login func
func (ac *AuthController) Login(c *gin.Context) {
	var input models.LoginRequest

	// gelen veriyi doğrula
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Geçersiz istek formatı",
		})
		return
	}

	// kullanıcıyı bul
	var user models.User
	if err := ac.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Kullanıcı bulunamadı",
		})
		return
	}

	// şifreyi kontrol et
	if err := ac.Hasher.CheckPasswordHash(input.Password, user.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Geçersiz şifre",
		})
		return
	}

	// token oluştur
	accessToken, refreshToken, err := utils.GenerateTokens(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Token oluşturulamadı",
		})
		return
	}

	// refresh token'ı veritabanına kaydet
	if err := utils.SaveRefreshToken(ac.DB, user.ID, refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Oturum bilgisi kaydedilemedi",
		})
		return
	}

	// kullanıcıyı ve token'ları döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"name":     user.Name,
				"surname":  user.Surname,
			},
		},
	})
}

// autologin func
func (ac *AuthController) AutoLogin(c *gin.Context) {
	var input models.RefreshTokenRequest

	// giriş doğrulama
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Geçersiz istek formatı",
		})
		return
	}

	// token'ı doğrula
	token, err := utils.ValidateToken(input.RefreshToken, true)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Geçersiz token",
		})
		return
	}

	// kullanıcı ID'sini al
	userID, err := utils.ExtractUserIDFromToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token içeriği geçersiz",
		})
		return
	}

	// veritabanında token'ı kontrol et
	var refreshToken models.RefreshToken
	if err := ac.DB.Where("token = ? AND user_id = ?", input.RefreshToken, userID).First(&refreshToken).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Oturum bulunamadı",
		})
		return
	}

	// token'ın süresi dolmuş mu kontrol et
	if time.Now().After(refreshToken.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Oturum süresi dolmuş",
		})
		return
	}

	// kullanıcıyı bul
	var user models.User
	if err := ac.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Kullanıcı bulunamadı",
		})
		return
	}

	// yeni token oluştur
	newAccessToken, _, err := utils.GenerateTokens(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Token oluşturulamadı",
		})
		return
	}

	// yanıtı döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  newAccessToken,
			"refresh_token": refreshToken.Token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"name":     user.Name,
				"surname":  user.Surname,
			},
		},
	})
}

// logout func
func (ac *AuthController) Logout(c *gin.Context) {
	// kullanıcı ID'sini token'dan al
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme bilgisi bulunamadı"})
		return
	}

	// transaction başlat
	tx := ac.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// kullanıcının tüm refresh token'larını sil
	if err := tx.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Oturum kapatılamadı"})
		return
	}

	// transaction'ı onayla
	tx.Commit()

	// response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Başarıyla çıkış yapıldı",
	})
}
