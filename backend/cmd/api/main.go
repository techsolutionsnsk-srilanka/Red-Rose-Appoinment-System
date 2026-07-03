package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"redrose/backend/internal/domain"
	"redrose/backend/internal/email"
	handlerhttp "redrose/backend/internal/handler/http"
	"redrose/backend/internal/middleware"
	"redrose/backend/internal/repository/mysql"
	"redrose/backend/internal/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// ── Database ────────────────────────────────────────────────
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC&time_zone=%%27%%2B00%%3A00%%27",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_NAME", "redrose_db"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	log.Println("✅ Connected to MySQL")

	if err := mysql.EnsureSchema(context.Background(), db); err != nil {
		log.Fatalf("Failed to ensure schema: %v", err)
	}
	log.Println("🧱 Schema ensured (appointments)")

	// ── Layers (Clean Architecture) ─────────────────────────────
	appointmentRepo := mysql.NewAppointmentRepository(db)

	var notifier domain.AppointmentNotifier
	if key := getEnv("RESEND_API_KEY", ""); key != "" {
		notifier = email.NewResendNotifier(
			key,
			getEnv("FROM_EMAIL", "RedRose <onboarding@resend.dev>"),
			getEnv("ADMIN_EMAIL", ""),
			getEnv("ADMIN_URL", "http://localhost:3001"),
		)
		log.Println("📧 Resend email notifier enabled")
	} else {
		log.Println("⚠️  RESEND_API_KEY not set — email notifications disabled")
	}

	appointmentUC := usecase.NewAppointmentUseCase(appointmentRepo, notifier)
	appointmentHandler := handlerhttp.NewAppointmentHandler(appointmentUC)

	// ── Router ──────────────────────────────────────────────────
	router := gin.Default()

	allowedOrigins := []string{
		"http://localhost:3000", // public web
		"http://localhost:3001", // admin
	}
	if origin := getEnv("ALLOWED_ORIGIN", ""); origin != "" {
		allowedOrigins = append(allowedOrigins, origin)
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "redrose-backend", "time": time.Now().Format(time.RFC3339)})
	})

	auth := middleware.AuthMiddleware()

	api := router.Group("/api")
	{
		appts := api.Group("/appointments")
		{
			// Public: customers on the web app can book without an admin login.
			appts.POST("", appointmentHandler.CreateAppointment)

			// Protected: admin-only management endpoints.
			admin := appts.Group("")
			admin.Use(auth)
			{
				admin.GET("", appointmentHandler.ListAppointments)
				admin.GET("/stats", appointmentHandler.GetStats)
				admin.GET("/:id", appointmentHandler.GetAppointment)
				admin.PUT("/:id/status", appointmentHandler.UpdateStatus)
				admin.PUT("/:id/notes", appointmentHandler.UpdateAdminNotes)
				admin.DELETE("/:id", appointmentHandler.DeleteAppointment)
			}
		}
	}

	port := getEnv("PORT", "8080")
	log.Printf("🚀 RedRose API starting on port %s", port)
	log.Printf("📍 Health: http://localhost:%s/health", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
