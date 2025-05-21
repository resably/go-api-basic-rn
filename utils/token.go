package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/resably/go-rest-api/models"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type JWTConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration //  time.Hour * 24 <--
	RefreshTokenExpiry time.Duration //  time.Hour * 24 * 7 <--
}

var jwtConfig *JWTConfig

func InitJWTConfig() {
	accessSecret := os.Getenv("JWT_SECRET_KEY")
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET_KEY")

	if accessSecret == "" || refreshSecret == "" {
		log.Println("Uyarı: JWT secretler .env dosyasında tanımlanmamış, varsayılan değerler kullanılıyor")
		accessSecret = "default_not_secure_access_secret"
		refreshSecret = "default_not_secure_refresh_secret"
	}

	jwtConfig = &JWTConfig{
		AccessTokenSecret:  accessSecret,
		RefreshTokenSecret: refreshSecret,
		AccessTokenExpiry:  time.Minute * 60,    // 60 dakika
		RefreshTokenExpiry: time.Hour * 24 * 14, // 14 gün
	}
}

// Token oluşturma
func GenerateTokens(user *models.User) (accessToken string, refreshToken string, err error) {
	// Access Token
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(jwtConfig.AccessTokenExpiry).Unix(),
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(jwtConfig.AccessTokenSecret))
	if err != nil {
		return "", "", fmt.Errorf("access token oluşturulamadı: %w", err)
	}

	// Refresh Token
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(jwtConfig.RefreshTokenExpiry).Unix(),
	}

	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(jwtConfig.RefreshTokenSecret))
	if err != nil {
		return "", "", fmt.Errorf("refresh token oluşturulamadı: %w", err)
	}

	return accessToken, refreshToken, nil
}

// Token doğrulama
func ValidateToken(tokenString string, isRefreshToken bool) (*jwt.Token, error) {
	secret := jwtConfig.AccessTokenSecret
	if isRefreshToken {
		secret = jwtConfig.RefreshTokenSecret
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("beklenmeyen imzalama yöntemi: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("geçersiz token")
	}

	return token, nil
}

// Token'dan kullanıcı ID'sini çıkarma
func ExtractUserIDFromToken(token *jwt.Token) (uint, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("geçersiz token claims")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("user_id claim bulunamadı")
	}

	return uint(userID), nil
}

// Refresh token'ı veritabanına kaydetme
func SaveRefreshToken(db *gorm.DB, userID uint, token string) error {
	expiry := time.Now().Add(jwtConfig.RefreshTokenExpiry)
	refreshToken := models.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiry,
		CreatedAt: time.Now(),
	}

	if err := db.Create(&refreshToken).Error; err != nil {
		return fmt.Errorf("refresh token kaydedilemedi: %w", err)
	}

	return nil
}

// Kullanıcı için eski refresh token'ları temizleme
func CleanOldRefreshTokens(db *gorm.DB, userID uint) error {
	if err := db.Where("user_id = ? AND expires_at < ?", userID, time.Now()).Delete(&models.RefreshToken{}).Error; err != nil {
		return fmt.Errorf("eski tokenlar temizlenemedi: %w", err)
	}
	return nil
}

func DeleteAllRefreshTokens(db *gorm.DB, userID uint) error {
	return db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}
