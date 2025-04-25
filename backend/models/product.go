package models

import "gorm.io/gorm"

type Product struct {
    gorm.Model
    SerialNumber  string `gorm:"unique"`
    Manufacturer  string
    ProductModel  string
}