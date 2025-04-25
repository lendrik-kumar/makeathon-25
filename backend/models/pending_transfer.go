package models

import "gorm.io/gorm"

type PendingTransfer struct {
	gorm.Model
	ProductID  uint
	NewOwnerID uint
}
