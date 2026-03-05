package main

import (
	"fmt"
	"log"
	"os"

	"ybg-backend-go/internal/delivery/http"
	"ybg-backend-go/internal/delivery/http/middleware" // Import middleware Anda
	"ybg-backend-go/internal/repository"
	"ybg-backend-go/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	// 2. Konfigurasi Database Connection
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Fatal("DB_URL is not set in .env")
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	fmt.Println("Successfully connected to Supabase via Transaction Pooler!")

	// 3. Seeding & Initialitation
	repository.SeedAdmin(db)

	productRepo := repository.NewProductRepository(db)
	productUC := usecase.NewProductUsecase(productRepo)
	productHandler := http.NewProductHandler(productUC)

	userRepo := repository.NewUserRepository(db)
	pointRepo := repository.NewPointRepository(db)
	userUC := usecase.NewUserUsecase(userRepo, pointRepo)
	userHandler := http.NewUserHandler(userUC)

	newsRepo := repository.NewNewsRepository(db)
	newsUC := usecase.NewNewsUsecase(newsRepo)
	newsHandler := http.NewNewsHandler(newsUC)

	brandRepo := repository.NewBrandRepository(db)
	brandUC := usecase.NewBrandUsecase(brandRepo)
	brandHandler := http.NewBrandHandler(brandUC)

	categoryRepo := repository.NewCategoryRepository(db)
	categoryUC := usecase.NewCategoryUsecase(categoryRepo)
	categoryHandler := http.NewCategoryHandler(categoryUC)

	pRepo := repository.NewPointRepository(db)
	pUC := usecase.NewPointUsecase(pRepo)
	pHandler := http.NewPointHandler(pUC)

	// 4. Setup Gin Router
	r := gin.Default()

	// --- 5. REGISTRASI ROUTES ---

	// A. Public Routes (Tanpa Login)
	r.POST("/register", userHandler.Create)
	r.POST("/login", userHandler.Login) // Fungsi Login harus ada di UserHandler
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP", "database": "connected"})
	})

	// B. Protected Routes (Harus Login)
	api := r.Group("/api")
	api.GET("/news", newsHandler.GetAll)
	api.GET("/products", productHandler.GetAll)
	api.GET("/products/:id", productHandler.GetByID)
	api.GET("/brand", brandHandler.GetAll)
	api.GET("/category", categoryHandler.GetAll)
	api.Use(middleware.AuthMiddleware()) // Cek token JWT
	{
		// --- FITUR PRODUCT ---
		// Semua user (Admin & Customer) bisa Lihat
		brandAdmin := api.Group("/brand") // Group ini sudah prefix /api/brand
		brandAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			// Gunakan 'brandAdmin', JANGAN 'api'
			brandAdmin.POST("/", brandHandler.Create)      // POST untuk Create
			brandAdmin.DELETE("/:id", brandHandler.Delete) // Tambahkan :id untuk Delete
		}

		// --- FITUR CATEGORY ---
		categoryAdmin := api.Group("/category") // Group ini sudah prefix /api/category
		categoryAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			// Gunakan 'categoryAdmin', JANGAN 'api'
			categoryAdmin.POST("/", categoryHandler.Create)      // POST untuk Create
			categoryAdmin.DELETE("/:id", categoryHandler.Delete) // Tambahkan :id untuk Delete
		}

		// Hanya Admin yang bisa Create, Update, Delete Product
		productAdmin := api.Group("/products")
		productAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			productAdmin.POST("/", productHandler.Create)
			productAdmin.PUT("/:id", productHandler.Update)
			productAdmin.DELETE("/:id", productHandler.Delete)
		}
		// Group Point
		points := api.Group("/points")
		{
			// Semua yang Login bisa lihat history
			points.GET("/history", pHandler.GetHistory)

			// Khusus Admin
			// Kita langsung pasang Middleware di baris rutenya saja supaya lebih clear
			points.POST("/", middleware.RoleMiddleware("admin"), pHandler.CreatePoint)
			points.GET("/all", middleware.RoleMiddleware("admin"), pHandler.GetAllSummaries)
		}

		newsAdmin := api.Group("/news")
		newsAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			newsAdmin.POST("/", newsHandler.Create)
			newsAdmin.PUT("/:id", newsHandler.Update)
			newsAdmin.DELETE("/:id", newsHandler.Delete)
		}

		// --- FITUR USER / PROFILE ---
		// Admin bisa lihat semua list user
		api.GET("/users", middleware.RoleMiddleware("admin"), userHandler.GetAll)

		// Customer & Admin bisa lihat/update profil sendiri
		api.GET("/profile", userHandler.GetByID)
		api.PUT("/profile", userHandler.Update)

		// Di dalam group admi

		// --- FITUR POINT (Read Only) ---
		// api.GET("/points/total", userHandler.GetPointTotal)
		// api.GET("/points/history", userHandler.GetPointHistory)
	}

	// 6. Jalankan Server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on port %s...\n", port)
	r.Run(":" + port)
}
