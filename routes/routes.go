package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/resably/go-rest-api/controllers"
	"github.com/resably/go-rest-api/middleware"
	"github.com/resably/go-rest-api/utils"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// bağımlılıkları oluştur
	hasher, _ := utils.NewBcryptHasher(12)
	authController := controllers.NewAuthController(db, hasher)
	productController := controllers.NewProductController(db)

	// API Versiyonlama
	api := r.Group("/api")
	{
		// V1 Routes
		v1 := api.Group("/v1")
		{
			setupAuthRoutes(v1, authController)
			setupProductRoutes(v1, productController)
		}
	}

	return r
}

// authRoutes
func setupAuthRoutes(group *gin.RouterGroup, authController *controllers.AuthController) {
	authGroup := group.Group("/auth")
	{
		authGroup.POST("/register", authController.Register)
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/autologin", authController.AutoLogin)
		authGroup.POST("/logout", middleware.JWTAuthMiddleware(), authController.Logout)
	}
}

// productRoutes
func setupProductRoutes(group *gin.RouterGroup, productController *controllers.ProductController) {
	productGroup := group.Group("/products", middleware.JWTAuthMiddleware())
	{
		productGroup.POST("/", productController.AddProduct)
		productGroup.GET("/", productController.GetProducts)
		productGroup.GET("/:id", productController.GetProductByID)
		productGroup.PUT("/:id", productController.UpdateProduct)
		productGroup.DELETE("/:id", productController.DeleteProduct)
	}
}
