package repository

import (
    "lostfound/internal/model"
    "lostfound/pkg/database"
)

type ItemRepository struct{}

func NewItemRepository() *ItemRepository {
    return &ItemRepository{}
}

func (r *ItemRepository) Create(item *model.Item) error {
    return database.DB.Create(item).Error
}

func (r *ItemRepository) FindAll(filters map[string]interface{}) ([]model.Item, error) {
    var items []model.Item
    query := database.DB.Preload("User")
    
    if category, ok := filters["category"]; ok && category != "" {
        query = query.Where("category = ?", category)
    }
    if location, ok := filters["location"]; ok && location != "" {
        query = query.Where("location ILIKE ?", "%"+location.(string)+"%")
    }
    if itemType, ok := filters["type"]; ok && itemType != "" {
        query = query.Where("type = ?", itemType)
    }
    
    err := query.Order("created_at DESC").Find(&items).Error
    return items, err
}

func (r *ItemRepository) FindByID(id uint) (*model.Item, error) {
    var item model.Item
    err := database.DB.Preload("User").First(&item, id).Error
    return &item, err
}

func (r *ItemRepository) Update(item *model.Item) error {
    return database.DB.Save(item).Error
}

func (r *ItemRepository) GetStats() (map[string]int64, error) {
    stats := make(map[string]int64)
    
    var totalLost int64
    var totalFound int64
    var totalClaims int64
    var pendingClaims int64
    
    database.DB.Model(&model.Item{}).Where("type = ?", "lost").Count(&totalLost)
    database.DB.Model(&model.Item{}).Where("type = ?", "found").Count(&totalFound)
    database.DB.Model(&model.Claim{}).Count(&totalClaims)
    database.DB.Model(&model.Claim{}).Where("status = ?", "pending").Count(&pendingClaims)
    
    stats["total_lost"] = totalLost
    stats["total_found"] = totalFound
    stats["total_claims"] = totalClaims
    stats["pending_claims"] = pendingClaims
    
    return stats, nil
}


func (r *ItemRepository) CreateClaim(claim *model.Claim) error {
    return database.DB.Create(claim).Error
}

func (r *ItemRepository) FindAllClaims() ([]model.Claim, error) {
    var claims []model.Claim
    err := database.DB.Preload("Item").Preload("User").Order("created_at DESC").Find(&claims).Error
    return claims, err
}

func (r *ItemRepository) UpdateClaim(claim *model.Claim) error {
    return database.DB.Save(claim).Error
}

func (r *ItemRepository) FindClaimByID(id uint) (*model.Claim, error) {
    var claim model.Claim
    err := database.DB.Preload("Item").Preload("User").First(&claim, id).Error
    return &claim, err
}