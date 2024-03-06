package util

import (
	"os"
	"time"

	db "PROJ/database"
	"PROJ/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

var jwtKey = []byte(os.Getenv("PRIV_KEY")) // กำหนด jwtKey ด้วยคีย์ส่วนตัวจากตัวแปรสภาพแวดล้อม

// GenerateTokens returns the access and refresh tokens
func GenerateTokens(uuid string) (string, string) {
	claim, accessToken := GenerateAccessClaims(uuid)
	refreshToken := GenerateRefreshClaims(claim)

	return accessToken, refreshToken
}

// GenerateAccessClaims returns a claim and a acess_token string
func GenerateAccessClaims(uuid string) (*models.Claims, string) {
	// ดึงเวลาปัจจุบัน
	t := time.Now()

	// สร้าง claim โดยกำหนด Issuer เป็น uuid และกำหนดเวลาหมดอายุและเวลาออก token
	claim := &models.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    uuid,
			ExpiresAt: t.Add(15 * time.Minute).Unix(),
			Subject:   "access_token",
			IssuedAt:  t.Unix(),
		},
	}

	// สร้าง token โดยใช้ claim และเลือกวิธีการเซ็นต์ข้อมูลด้วย HS256 algorithm
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	// เซ็นต์ token เป็น string โดยใช้คีย์ที่กำหนดไว้
	tokenString, err := token.SignedString(jwtKey)
	// ตรวจสอบข้อผิดพลาด
	if err != nil {
		panic(err)
	}
	// คืนค่า claim และ access token string กลับไป
	return claim, tokenString
}

// GenerateRefreshClaims returns refresh_token
func GenerateRefreshClaims(cl *models.Claims) string {
	// ค้นหาข้อมูล refresh token ที่เกี่ยวข้องกับ issuer ในฐานข้อมูล
	result := db.DB.Where(&models.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer: cl.Issuer,
		},
	}).Find(&models.Claims{})

	// ตรวจสอบจำนวนของ refresh token ที่เก็บ
	// หากจำนวนมากกว่า 3 ให้ลบ refresh token ทั้งหมดและเพิ่มเฉพาะใหม่
	if result.RowsAffected > 3 {
		db.DB.Where(&models.Claims{
			StandardClaims: jwt.StandardClaims{Issuer: cl.Issuer},
		}).Delete(&models.Claims{})
	}

	// ดึงเวลาปัจจุบัน
	t := time.Now()
	// สร้าง claim ใหม่สำหรับ refresh token
	refreshClaim := &models.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    cl.Issuer,                         // ผู้ออก token
			ExpiresAt: t.Add(30 * 24 * time.Hour).Unix(), // เวลาหมดอายุของ token
			Subject:   "refresh_token",                   // วัตถุประสงค์ของ token
			IssuedAt:  t.Unix(),                          // เวลาที่ออก token
		},
	}
	// สร้าง claim ในฐานข้อมูล
	db.DB.Create(&refreshClaim) // สร้าง claim ในฐานข้อมูล

	// สร้าง refresh token จาก claim ที่สร้างขึ้น
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaim)

	// เซ็นต์ refresh token เป็น string โดยใช้คีย์ที่กำหนดไว้
	refreshTokenString, err := refreshToken.SignedString(jwtKey)

	if err != nil {
		panic(err)
	}

	// คืนค่า refresh token string กลับไป
	return refreshTokenString
}

// SecureAuth คืนค่า middleware ที่ใช้ป้องกันเส้นทางเอกสารทั้งหมด
func SecureAuth() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// ดึง access token จากคุกกี้ของคำขอ
		accessToken := c.Cookies("access_token")
		// สร้างตัวแปร claims เพื่อเก็บข้อมูลจาก token
		claims := new(models.Claims)

		// ตรวจสอบความถูกต้องของ token และดึงข้อมูลที่ถูกเซ็นต์ไว้ใน claims
		token, err := jwt.ParseWithClaims(accessToken, claims,
			func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})

		// ตรวจสอบความถูกต้องของ token
		if token.Valid {
			// ตรวจสอบว่า token หมดอายุหรือไม่
			if claims.ExpiresAt < time.Now().Unix() {
				// หาก token หมดอายุ ส่งค่าข้อผิดพลาดแบบไม่มีสิทธิ์
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"general": "Token Expired",
				})
			}
		} else if ve, ok := err.(*jwt.ValidationError); ok {
			// กรณีเกิดข้อผิดพลาดของ token
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				// หาก token ไม่ถูกต้อง ลบคุกกี้ทั้งหมดที่เกี่ยวข้อง
				c.ClearCookie("access_token", "refresh_token")
				// ส่งค่าข้อผิดพลาดแบบไม่มีสิทธิ์
				return c.SendStatus(fiber.StatusForbidden)
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				// กรณี token หมดอายุหรือยังไม่ได้เริ่มใช้งาน
				// ส่งค่าข้อผิดพลาดแบบไม่มีสิทธิ์
				return c.SendStatus(fiber.StatusUnauthorized)
			} else {
				// กรณีไม่สามารถจัดการ token ได้
				// ลบคุกกี้ทั้งหมดที่เกี่ยวข้อง
				c.ClearCookie("access_token", "refresh_token")
				// ส่งค่าข้อผิดพลาดแบบไม่มีสิทธิ์
				return c.SendStatus(fiber.StatusForbidden)
			}
		}

		// นำค่า issuer ของ token เข้าไปในตัวแปร locals เพื่อให้สามารถเข้าถึงได้ใน middleware ถัดไป
		c.Locals("id", claims.Issuer)
		// ส่งคำขอต่อไป
		return c.Next()
	}
}

// GetAuthCookies ส่งคุกกี้สองชนิด คือ access_token และ refresh_token
func GetAuthCookies(accessToken, refreshToken string) (*fiber.Cookie, *fiber.Cookie) {
	// สร้างคุกกี้สำหรับ access_token
	accessCookie := &fiber.Cookie{
		Name:     "access_token",                 // ตั้งชื่อคุกกี้
		Value:    accessToken,                    // กำหนดค่าของคุกกี้ด้วย access_token ที่รับมา
		Expires:  time.Now().Add(24 * time.Hour), // กำหนดเวลาหมดอายุของคุกกี้เป็น 24 ชั่วโมง
		HTTPOnly: true,                           // ตั้งค่า HTTPOnly เพื่อป้องกันการเข้าถึงคุกกี้ผ่าน JavaScript
		Secure:   true,                           // ตั้งค่า Secure เพื่อให้คุกกี้ถูกส่งผ่านการเชื่อมต่อที่ปลอดภัยเท่านั้น
	}

	// สร้างคุกกี้สำหรับ refresh_token
	refreshCookie := &fiber.Cookie{
		Name:     "refresh_token",                     // ตั้งชื่อคุกกี้
		Value:    refreshToken,                        // กำหนดค่าของคุกกี้ด้วย refresh_token ที่รับมา
		Expires:  time.Now().Add(10 * 24 * time.Hour), // กำหนดเวลาหมดอายุของคุกกี้เป็น 10 วัน
		HTTPOnly: true,                                // ตั้งค่า HTTPOnly เพื่อป้องกันการเข้าถึงคุกกี้ผ่าน JavaScript
		Secure:   true,                                // ตั้งค่า Secure เพื่อให้คุกกี้ถูกส่งผ่านการเชื่อมต่อที่ปลอดภัยเท่านั้น
	}

	// คืนค่าคุกกี้ access_token และ refresh_token
	return accessCookie, refreshCookie
}
