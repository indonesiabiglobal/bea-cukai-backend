package userService

import (
	"Bea-Cukai/helper"
	"Bea-Cukai/model"
	"Bea-Cukai/repo/userLogRepository"
	"Bea-Cukai/repo/userRepository"
	"errors"
	"fmt" // Add this for debugging

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// ... rest of the code

// In LoginUser function, uncomment the fmt.Printf lines:
// if logErr != nil {
//     fmt.Printf("❌ Error logging failed login: %v\n", logErr)
// }

// This will print errors to console for debugging
