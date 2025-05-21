package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/resably/go-rest-api/models"
	"gorm.io/gorm"
)

type ProductController struct {
	DB *gorm.DB
}

func NewProductController(db *gorm.DB) *ProductController {
	return &ProductController{
		DB: db,
	}
}

// addProduct func

func (pc *ProductController) AddProduct(c *gin.Context) {
	var input models.Products

	// gelen veriyi doğrula
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Geçersiz istek formatı",
			"details": err.Error(),
		})
		return
	}

	// ürün ekle
	if err := pc.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ürün eklenirken hata oluştu",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Ürün başarıyla eklendi",
		"product": gin.H{
			"id":          input.ID,
			"name":        input.Name,
			"description": input.Description,
			"price":       input.Price,
			"stock":       input.Stock,
			"category":    input.Category,
			"created_at":  input.CreatedAt,
			"updated_at":  input.UpdatedAt,
		},
	})
}

// getProducts func
func (pc *ProductController) GetProducts(c *gin.Context) {
	var products []models.Products

	// ürünleri al
	if err := pc.DB.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ürünler alınırken hata oluştu",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
	})
}

// getProductByID func
func (pc *ProductController) GetProductByID(c *gin.Context) {
	id := c.Param("id")
	var product models.Products

	// ürünü al
	if err := pc.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Ürün bulunamadı",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"product": product,
	})
}

// updateProduct func
func (pc *ProductController) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var input models.Products

	// gelen veriyi doğrula
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Geçersiz istek formatı",
		})
		return
	}

	var product models.Products

	// ürünü al
	if err := pc.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Ürün bulunamadı",
		})
		return
	}

	// ürünü güncelle
	if err := pc.DB.Model(&product).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ürün güncellenirken hata oluştu",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ürün başarıyla güncellendi",
		"product": gin.H{
			"id":          product.ID,
			"name":        product.Name,
			"description": product.Description,
			"price":       product.Price,
			"stock":       product.Stock,
			"category":    product.Category,
			"created_at":  product.CreatedAt,
			"updated_at":  product.UpdatedAt,
		},
	})
}

// deleteProduct func
func (pc *ProductController) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Products

	// ürünü al
	if err := pc.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Ürün bulunamadı",
		})
		return
	}

	// ürünü sil
	if err := pc.DB.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ürün silinirken hata oluştu",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ürün başarıyla silindi",
		"product": gin.H{
			"id": product.ID,
		},
	})
}
