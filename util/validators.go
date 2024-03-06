// validators = การตรวจสอบความถูกต้อง

package util

import (
	"regexp"

	"PROJ/models"

	valid "github.com/asaskevich/govalidator"
)

// ตรวจสอบว่าสตริงว่างหรือไม่
func IsEmpty(str string) (bool, string) {
	// not empty
	if valid.HasWhitespaceOnly(str) && str != "" {
		return true, "Must not be empty"
	}

	return false, ""
}

// ตรวจสอบข้อมูลของผู้ใช้งานสำหรับการลงทะเบียน
func ValidationRegister(u *models.User) *models.UserErrors {
	e := &models.UserErrors{}
	e.Err, e.Username = IsEmpty(u.Username) // ตรวจสอบว่าชื่อผู้ใช้งานว่างหรือไม่

	// ฟังก์ชัน IsEmail นี้จะคืนค่า true ถ้าค่าที่รับเข้ามาเป็นอีเมลที่ถูกต้อง และคืนค่า false ถ้าไม่ถูกต้อง
	// ถ้าเงื่อนไข !valid.IsEmail(u.Email) เป็นจริง (ค่าอีเมลไม่ถูกต้อง), ก็จะทำการกำหนดค่า e.Err เป็น true เพื่อบ่งชี้ว่ามีข้อผิดพลาดเกิดขึ้น
	if !valid.IsEmail(u.Email) {
		e.Err, e.Email = true, "Must be a valid email"
	}

	// สร้างรูปแบบการตรวจสอบ regex เพื่อตรวจสอบว่ามีตัวเลขอย่างน้อยหนึ่งตัวในสตริง
	re := regexp.MustCompile(`\d`)
	if !(len(u.Password) >= 8 && valid.HasLowerCase(u.Password) && valid.HasUpperCase(u.Password) && re.MatchString(u.Password)) {
		e.Err, e.Password = true, "Length of password should be atleast 8 and it must be a combination of uppercase letters, lowercase letters and numbers"
	}

	return e
}
