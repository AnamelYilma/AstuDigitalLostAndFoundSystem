package model

import (
    "time"
    "gorm.io/gorm"
)

type Item struct {
    ID          uint           `gorm:"primarykey" json:"id"`
    UserID      uint           `gorm:"not null" json:"user_id"`
    Type        string         `gorm:"size:10;not null" json:"type"`
    Title       string         `gorm:"size:200;not null" json:"title"`
    Category    string         `gorm:"size:50;not null" json:"category"`
    Color       string         `gorm:"size:50" json:"color"`
    Brand       string         `gorm:"size:100" json:"brand"`
    Location    string         `gorm:"size:200;not null" json:"location"`
    Date        string         `gorm:"size:20" json:"date"`
    Description string         `gorm:"type:text" json:"description"`
    Image       string         `gorm:"size:500" json:"image"`
    Status      string         `gorm:"size:20;default:'open'" json:"status"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
    
    User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type Claim struct {
    ID          uint           `gorm:"primarykey" json:"id"`
    ItemID      uint           `gorm:"not null" json:"item_id"`
    UserID      uint           `gorm:"not null" json:"user_id"`
    Description string         `gorm:"type:text" json:"description"`
    Status      string         `gorm:"size:20;default:'pending'" json:"status"`
    AdminRemarks string        `gorm:"type:text" json:"admin_remarks"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
    
    Item Item `gorm:"foreignKey:ItemID" json:"item"`
    User User `gorm:"foreignKey:UserID" json:"user"`
}

func (Item) TableName() string {
    return "items"
}

func (Claim) TableName() string {
    return "claims"
}