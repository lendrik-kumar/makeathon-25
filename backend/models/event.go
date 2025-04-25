package models

import "gorm.io/gorm"

type Event struct {
    gorm.Model
    ProductID         uint
    EventType        string 
    EventData        string 
    PreviousEventHash string
    EventHash        string
    CreatedBy        uint 
}