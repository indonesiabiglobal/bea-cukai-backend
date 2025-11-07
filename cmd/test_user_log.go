package main

import (
	"Bea-Cukai/database"
	"Bea-Cukai/model"
	"Bea-Cukai/repo/userLogRepository"
	"fmt"
	"log"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fmt.Println("Testing User Log System...")
	fmt.Println("==========================")

	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate to ensure table exists
	fmt.Println("\n1. Auto-migrating user_log table...")
	err = db.AutoMigrate(&model.UserLog{})
	if err != nil {
		log.Fatalf("Failed to migrate user_log: %v", err)
	}
	fmt.Println("✓ Migration successful")

	// Also migrate user table for new columns
	fmt.Println("\n2. Auto-migrating user table...")
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatalf("Failed to migrate user: %v", err)
	}
	fmt.Println("✓ Migration successful")

	// Create repository
	userLogRepo := userLogRepository.NewUserLogRepository(db)

	// Test insert
	fmt.Println("\n3. Testing insert to user_log...")
	logEntry := model.UserLogRequest{
		UserId:    "TEST001",
		Username:  "test_user",
		Action:    "login",
		IpAddress: "127.0.0.1",
		UserAgent: "Test Agent v1.0",
		Status:    "success",
		Message:   "Test login from Go script",
	}

	createdLog, err := userLogRepo.CreateLog(logEntry)
	if err != nil {
		log.Fatalf("Failed to create log: %v", err)
	}

	fmt.Printf("✓ Log created successfully with ID: %d\n", createdLog.Id)
	fmt.Printf("  User ID: %s\n", createdLog.UserId)
	fmt.Printf("  Username: %s\n", createdLog.Username)
	fmt.Printf("  Action: %s\n", createdLog.Action)
	fmt.Printf("  IP: %s\n", createdLog.IpAddress)
	fmt.Printf("  Status: %s\n", createdLog.Status)
	fmt.Printf("  Created At: %s\n", createdLog.CreatedAt)

	// Get all logs
	fmt.Println("\n4. Retrieving all logs...")
	logs, total, err := userLogRepo.GetAll(model.UserLogListRequest{
		Page:  1,
		Limit: 10,
	})
	if err != nil {
		log.Fatalf("Failed to get logs: %v", err)
	}

	fmt.Printf("✓ Total logs in database: %d\n", total)
	if total > 0 {
		fmt.Println("\nLatest logs:")
		for i, l := range logs {
			fmt.Printf("  %d. [%s] %s - %s by %s (%s) at %s\n",
				i+1,
				l.Status,
				l.Action,
				l.Message,
				l.Username,
				l.IpAddress,
				l.CreatedAt.Format("2006-01-02 15:04:05"),
			)
		}
	}

	// Check user table structure
	fmt.Println("\n5. Checking user table for new columns...")
	var users []model.User
	result := db.Limit(3).Find(&users)
	if result.Error != nil {
		log.Fatalf("Failed to query users: %v", result.Error)
	}

	fmt.Printf("✓ Found %d users\n", len(users))
	for i, u := range users {
		fmt.Printf("  %d. %s (%s) - Login count: %d, Last login: %v, Last IP: %s\n",
			i+1,
			u.Username,
			u.Id,
			u.LoginCount,
			u.LastLoginAt,
			u.LastLoginIp,
		)
	}

	fmt.Println("\n==========================")
	fmt.Println("Test completed successfully!")
	fmt.Println("\nIf you see this message, the user_log system is working correctly.")
	fmt.Println("If login logs are not being saved, check:")
	fmt.Println("1. Make sure the backend is running with the latest code")
	fmt.Println("2. Check if there are any errors in the terminal/logs")
	fmt.Println("3. Verify the database connection is working")
}
