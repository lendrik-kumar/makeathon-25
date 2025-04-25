package models

import "gorm.io/gorm"

type PendignTransfer struct {
	gorm.Model
	ProductID  uint
	NewOwnerID uint
}
