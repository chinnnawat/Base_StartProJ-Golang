package controller

import (
	"math/rand"
	"os"
	"time"

	"PROJ/models"
	"PROJ/util"

	db "PROJ/database"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("PRIV_KEY"))

// CreateUser route registers a User into the database
func CreateUser(c *fiber.Ctx) error {
	u := new(models.User)

	// แปลงข้อมูลที่รับเข้ามาให้อยู่ในรูปแบบของโมเดลผู้ใช้งาน
	if err := c.BodyParser(u); err != nil {
		// กรณีมีข้อผิดพลาดในการแปลงข้อมูลเข้ารูปแบบของโมเดลผู้ใช้งาน
		return c.JSON(fiber.Map{
			"error": true,
			"input": "Please review your input",
		})
	}

	// ตรวจสอบความถูกต้องของข้อมูลโดยเรียกใช้ ValidateRegister จากไฟล์ util/validators.go
	errors := util.ValidationRegister(u)
	if errors.Err {
		return c.JSON(errors)
	}

	// ตรวจสอบว่า email ไม่ซ้ำกัน
	if count := db.DB.Where(&models.User{Email: u.Email}).First(new(models.User)).RowsAffected; count > 0 {
		errors.Err, errors.Email = true, "Email is already registered"
	}

	// ตรวจสอบว่า username ไม่ซ้ำกัน
	if count := db.DB.Where(&models.User{Username: u.Username}).First(new(models.User)).RowsAffected; count > 0 {
		errors.Err, errors.Username = true, "Username is already registered"
	}
	if errors.Err {
		return c.JSON(errors)
	}

	// เข้ารหัสรหัสผ่านด้วยการสุ่ม
	// rand.Intn(bcrypt.MaxCost-bcrypt.MinCost)+bcrypt.MinCost: สร้างค่าเกลือ (salt) โดยการสุ่มตัวเลขในช่วงระหว่าง bcrypt.MinCost ถึง bcrypt.MaxCost ซึ่งเป็นค่าที่กำหนดความคมชัดของการเข้ารหัส การสุ่มค่าเกลือนี้ช่วยให้ผลลัพธ์ของการเข้ารหัสไม่ซ้ำกันแม้ข้อมูลเข้าที่เหมือนกัน
	password := []byte(u.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword(
		password,
		rand.Intn(bcrypt.MaxCost-bcrypt.MinCost)+bcrypt.MinCost,
	)
	if err != nil {
		panic(err)
	}
	u.Password = string(hashedPassword)

	// ลงทะเบียนผู้ใช้ในฐานข้อมูลและสร้างโทเคนการเข้าถึงและการรีเฟรช
	if err := db.DB.Create(&u).Error; err != nil {
		return c.JSON(fiber.Map{
			"error":   true,
			"general": "Something went wrong, please try again later. 😕", // เกิดข้อผิดพลาดบางอย่าง โปรดลองอีกครั้งในภายหลัง
		})
	}

	// ตั้งค่าคุกกี้สำหรับการอนุญาตเข้าถึงและรีเฟรช
	accessToken, refreshToken := util.GenerateTokens(u.UUID.String())
	accessCookie, refreshCookie := util.GetAuthCookies(accessToken, refreshToken)

	c.Cookie(accessCookie)
	c.Cookie(refreshCookie)

	// return token เมื่อทำการ
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// LoginUser route logins a user in the app
func LoginUser(c *fiber.Ctx) error {
	type LoginInput struct {
		Identity string `json:"identity"` // ตัวแปรที่ใช้รับอีเมลหรือชื่อผู้ใช้
		Password string `json:"password"`
	}

	input := new(LoginInput)

	// แปลงข้อมูลที่รับเข้ามาให้อยู่ในรูปแบบของโครงสร้าง LoginInput
	if err := c.BodyParser(input); err != nil {
		return c.JSON(fiber.Map{"error": true, "input": "Please review your input"})
	}

	// ตรวจสอบว่ามีผู้ใช้งานในระบบหรือไม่
	u := new(models.User)
	if res := db.DB.Where(
		&models.User{Email: input.Identity}).Or(
		&models.User{Username: input.Identity},
	).First(&u); res.RowsAffected <= 0 {
		return c.JSON(fiber.Map{"error": true, "general": "Invalid Credentials."}) // ข้อมูลเข้าไม่ถูกต้อง
	}

	// เปรียบเทียบรหัสผ่านกับการเข้ารหัส
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(input.Password)); err != nil {
		return c.JSON(fiber.Map{"error": true, "general": "Invalid Credentials."}) // ข้อมูลเข้าไม่ถูกต้อง
	}

	// ตั้งค่าคุกกี้สำหรับการอนุญาตเข้าถึงและรีเฟรช
	accessToken, refreshToken := util.GenerateTokens(u.UUID.String())
	accessCookie, refreshCookie := util.GetAuthCookies(accessToken, refreshToken)
	c.Cookie(accessCookie)
	c.Cookie(refreshCookie)

	// ส่งโทเคนกลับ
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func GetAccessToken(c *fiber.Ctx) error {
	// ดึงค่า refresh token จากคุกกี้
	refreshToken := c.Cookies("refresh_token")

	// สร้าง Claims ใหม่เพื่อเก็บข้อมูลจาก refresh token
	refreshClaims := new(models.Claims)

	// ทำการตรวจสอบความถูกต้องของ refresh token และดึงข้อมูล Claims ออกมา
	// Claims ช่วยให้เราสามารถเข้าถึงข้อมูลที่อยู่ใน Token ได้ง่ายๆ โดยไม่ต้องกังวลเรื่องการถอดรหัส Token หรือการตรวจสอบความถูกต้องของ Token เอง เพราะฟังก์ชัน jwt.ParseWithClaims() จะทำการทำงานนี้ให้เราอัตโนมัติ
	token, _ := jwt.ParseWithClaims(refreshToken, refreshClaims,
		func(t *jwt.Token) (interface{}, error) {

			// return jwtKey, nil: ใน anonymous function นี้ เราส่งกลับคีย์ที่ใช้ในการเข้ารหัส Token กลับมาเพื่อให้ jwt.ParseWithClaims() ใช้ในการตรวจสอบความถูกต้องของ Token และการถอดรหัส
			return jwtKey, nil
		})
	// ตรวจสอบว่ามี refresh token ในฐานข้อมูลหรือไม่
	if res := db.DB.Where(
		"expires_at = ? AND issued_at = ? AND issuer = ?", refreshClaims.ExpiresAt, refreshClaims.IssuedAt, refreshClaims.Issuer,
	).First(&models.Claims{}); res.RowsAffected <= 0 {
		// ถ้าไม่มี refresh token ในฐานข้อมูล เราจะล้างคุกกี้ access token และ refresh token และส่งค่าผิดพลาดกลับไป
		c.ClearCookie("access_token", "refresh_token")
		return c.SendStatus(fiber.StatusForbidden)
	}

	// ตรวจสอบความถูกต้องของ refresh token
	if token.Valid {
		// ตรวจสอบว่า refresh token หมดอายุหรือไม่
		if refreshClaims.ExpiresAt < time.Now().Unix() {
			// ถ้า refresh token หมดอายุ เราจะล้างคุกกี้ access token และ refresh token และส่งค่าผิดพลาดกลับไป
			c.ClearCookie("access_token", "refresh_token")
			return c.SendStatus(fiber.StatusForbidden)
		}

	} else {
		// ถ้า refresh token มีโครงสร้างไม่ถูกต้อง เราจะล้างคุกกี้ access token และ refresh token และส่งค่าผิดพลาดกลับไป
		c.ClearCookie("access_token", "refresh_token")
		return c.SendStatus(fiber.StatusForbidden)
	}

	// สร้าง access token ใหม่จาก issuer ของ refresh token
	_, accessToken := util.GenerateAccessClaims(refreshClaims.Issuer)

	// ตั้งค่าคุกกี้ access token ให้กับผู้ใช้
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
	})

	// ส่ง access token กลับไปให้ผู้ใช้
	return c.JSON(fiber.Map{"access_token": accessToken})
}

func GetUserData(c *fiber.Ctx) error {
	// รับรหัสผู้ใช้จากข้อมูลพื้นฐานของคำขอ (context)
	id := c.Locals("id")

	// สร้างตัวแปรเพื่อเก็บข้อมูลของผู้ใช้
	u := new(models.User)

	// ค้นหาข้อมูลผู้ใช้ในฐานข้อมูลโดยใช้ UUID
	if res := db.DB.Where("uuid = ?", id).First(u); res.RowsAffected <= 0 {
		// ถ้าไม่พบข้อมูลผู้ใช้ ส่งค่าผิดพลาดกลับไป
		return c.JSON(fiber.Map{"error": true, "general": "Cannot find the User"})
	}

	userInfo := map[string]interface{}{
		"firstName":    u.FirstName,
		"lastName":     u.LastName,
		"email":        u.Email,
		"username":     u.Username,
		"mobileNumber": u.MobileNumber,
		"pictureUrl":   u.PictureUrl,
	}

	// ส่งข้อมูลผู้ใช้กลับไปให้ผู้ใช้
	return c.JSON(userInfo)
}
