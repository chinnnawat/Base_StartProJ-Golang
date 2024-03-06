package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GenerateISOString สร้างสตริงเวลาที่เป็นรูปแบบของ "2006-01-02T15:04:05.999Z07:00" ซึ่งเป็นรูปแบบที่เหมือนกับ JavaScript Date.now().toISOString()
func GenerateISOString() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.999Z07:00")
}

// Base เก็บคอลัมน์ที่เป็นร่วมกันสำหรับตารางทั้งหมด
type Base struct {
	ID        uint      `gorm:"primaryKey"`
	UUID      uuid.UUID `json:"_id" gorm:"primaryKey;autoIncrement:false"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

// BeforeCreate จะเรียกก่อนทุกครั้งที่มีการเพิ่มข้อมูลใหม่ลงในฐานข้อมูล มีหน้าที่สร้าง UUID ใหม่สำหรับแถวใหม่ และกำหนดค่า timestamp สำหรับ CreatedAt และ UpdatedAt
func (base *Base) BeforeCreate(tx *gorm.DB) error {
	base.UUID = uuid.New()
	t := GenerateISOString()
	base.CreatedAt, base.UpdatedAt = t, t
	return nil
}

// // AfterUpdate จะถูกเรียกหลังจากทุกครั้งที่มีการอัปเดตข้อมูลในฐานข้อมูล มีหน้าที่อัปเดตค่า timestamp สำหรับ UpdatedAt เท่านั้น
func (base *Base) AfterUpdate(tx *gorm.DB) error {
	base.UpdatedAt = GenerateISOString()
	return nil
}
