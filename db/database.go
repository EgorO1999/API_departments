package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Department struct {
	ID        int    `gorm:"primaryKey"`
	Name      string `gorm:"size:200;not null"`
	ParentID  *int
	Parent    *Department  `gorm:"foreignKey:ParentID"`
	Children  []Department `gorm:"foreignKey:ParentID"`
	Employees []Employee
	CreatedAt time.Time
}

type Employee struct {
	ID           int    `gorm:"primaryKey"`
	DepartmentID int    `gorm:"not null"`
	FullName     string `gorm:"size:200;not null"`
	Position     string `gorm:"size:200;not null"`
	HiredAt      *time.Time
	CreatedAt    time.Time
}

func ConnectGORM() *gorm.DB {
	dsn := "host=db user=postgres password=postgres dbname=departments_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД через GORM: %v", err)
	}
	log.Println("GORM подключен и модели мигрированы")
	return db
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func RunMigrations() {
	host := getEnv("DB_HOST", "db")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "departments_db")
	port := getEnv("DB_PORT", "5432")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer conn.Close()

	if err := goose.Up(conn, "./db/migrations"); err != nil {
		log.Fatalf("Ошибка применения миграций: %v", err)
	}

	log.Println("Миграции успешно применены")
}
