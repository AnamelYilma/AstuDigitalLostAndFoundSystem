// internal/repository/item_repository.go
package repository

import (
    "lostfound/internal/model"
    "lostfound/pkg/database"
)

type ItemRepository struct{}

func NewItemRepository() *ItemRepository {
    return &ItemRepository{}
}

// Create item
func (r *ItemRepository) Create(item *model.Item) error {
    return database.DB.Create(item).Error
}

// Find all items with filters
func (r *ItemRepository) FindAll(filters map[string]interface{}) ([]model.Item, error) {
    var items []model.Item
    query := database.DB.Preload("User")
    
    // Apply filters
    if category, ok := filters["category"]; ok && category != "" {
        query = query.Where("category = ?", category)
    }
    if location, ok := filters["location"]; ok && location != "" {
        query = query.Where("location ILIKE ?", "%"+location.(string)+"%")
    }
    if status, ok := filters["status"]; ok && status != "" {
        query = query.Where("status = ?", status)
    }
    if itemType, ok := filters["type"]; ok && itemType != "" {
        query = query.Where("type = ?", itemType)
    }
    
    err := query.Order("created_at DESC").Find(&items).Error
    return items, err
}

// Find by ID
func (r *ItemRepository) FindByID(id uint) (*model.Item, error) {
    var item model.Item
    err := database.DB.Preload("User").First(&item, id).Error
    return &item, err
}

// Update
func (r *ItemRepository) Update(item *model.Item) error {
    return database.DB.Save(item).Error
}

// Delete
func (r *ItemRepository) Delete(id uint) error {
    return database.DB.Delete(&model.Item{}, id).Error
}

// Get statistics
func (r *ItemRepository) GetStats() (map[string]int64, error) {
    stats := make(map[string]int64)
    
    database.DB.Model(&model.Item{}).Where("type = ?", "lost").Count(&stats["total_lost"])
    database.DB.Model(&model.Item{}).Where("type = ?", "found").Count(&stats["total_found"])
    database.DB.Model(&model.Claim{}).Count(&stats["total_claims"])
    database.DB.Model(&model.Claim{}).Where("status = ?", "pending").Count(&stats["pending_claims"])
    
    return stats, nil
}

// Create claim
func (r *ItemRepository) CreateClaim(claim *model.Claim) error {
    return database.DB.Create(claim).Error
}

// Find claims (for admin)
func (r *ItemRepository) FindAllClaims() ([]model.Claim, error) {
    var claims []model.Claim
    err := database.DB.Preload("Item").Preload("User").Order("created_at DESC").Find(&claims).Error
    return claims, err
}

// Update claim
func (r *ItemRepository) UpdateClaim(claim *model.Claim) error {
    return database.DB.Save(claim).Error
}

// Find claim by ID
func (r *ItemRepository) FindClaimByID(id uint) (*model.Claim, error) {
    var claim model.Claim
    err := database.DB.Preload("Item").Preload("User").First(&claim, id).Error
    return &claim, err
}