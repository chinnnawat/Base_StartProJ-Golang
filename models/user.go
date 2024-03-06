package models

import (
	"github.com/dgrijalva/jwt-go"
)

type User struct {
	Base
	FirstName    *string `json:"firstName" validate:"required"`
	LastName     *string `json:"lastName" validate:"required"`
	Email        string  `json:"email" gorm:"unique"`
	Username     string  `json:"username" gorm:"unique"`
	Password     string  `json:"password"`
	MobileNumber *string `json:"mobileNumber" validate:"omitempty,len=10"`
	PictureUrl   *string `json:"pictureUrl"`
	// Post         []*Post `json:"posts"`
}

// UserErrors แทนรูปแบบของข้อผิดพลาดสำหรับเส้นทางของผู้ใช้
type UserErrors struct {
	Err      bool   `json:"error"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims แทนโครงสร้างของโทเค็น JWT
type Claims struct {
	jwt.StandardClaims
	ID uint `gorm:"primaryKey"`
}
