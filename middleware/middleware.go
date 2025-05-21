package middleware

import (
	"net/http"
	"strings"

	"github.com/resably/go-rest-api/utils"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Authorization header'ını al
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Yetkilendirme gerekli",
			})
			return
		}

		// 2. "Bearer token" şeklinde mi kontrol et
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Geçersiz yetkilendirme formatı",
			})
			return
		}

		// 3. Token'ı doğrula (Access Token)
		token, err := utils.ValidateToken(parts[1], false) // false → access token
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Geçersiz veya süresi dolmuş token",
			})
			return
		}

		// 4. Token'dan userID'yi çek
		userID, err := utils.ExtractUserIDFromToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token'dan kullanıcı bilgisi alınamadı",
			})
			return
		}

		// 5. userID'yi context'e ekle
		c.Set("userID", userID)

		// 6. İşleme devam et
		c.Next()
	}
}
