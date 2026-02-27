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
		query = query.Where("location = ?", location)
	}
	if colors, ok := filters["colors"]; ok {
		if colorList, ok := colors.([]string); ok && len(colorList) > 0 {
			knownColors := []string{"red", "green", "blue", "yellow", "black", "white", "gray", "brown", "orange", "purple", "pink", "gold", "silver"}
			hasOther := false
			standardSelected := make([]string, 0, len(colorList))
			for _, c := range colorList {
				if c == "other" {
					hasOther = true
					continue
				}
				standardSelected = append(standardSelected, c)
			}

			switch {
			case hasOther && len(standardSelected) > 0:
				query = query.Where("(LOWER(color) IN ? OR (LOWER(color) NOT IN ? AND TRIM(color) <> ''))", standardSelected, knownColors)
			case hasOther:
				query = query.Where("LOWER(color) NOT IN ? AND TRIM(color) <> ''", knownColors)
			default:
				query = query.Where("LOWER(color) IN ?", standardSelected)
			}
		}
	}
	if itemType, ok := filters["type"]; ok && itemType != "" {
		query = query.Where("type = ?", itemType)
	}
	if approvalStatus, ok := filters["approval_status"]; ok && approvalStatus != "" {
		query = query.Where("approval_status = ?", approvalStatus)
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
	var pendingItems int64

	database.DB.Model(&model.Item{}).Where("type = ? AND approval_status = ?", "lost", "approved").Count(&totalLost)
	database.DB.Model(&model.Item{}).Where("type = ? AND approval_status = ?", "found", "approved").Count(&totalFound)
	database.DB.Model(&model.Claim{}).Count(&totalClaims)
	database.DB.Model(&model.Claim{}).Where("status = ?", "pending").Count(&pendingClaims)
	database.DB.Model(&model.Item{}).Where("approval_status = ?", "pending").Count(&pendingItems)

	stats["total_lost"] = totalLost
	stats["total_found"] = totalFound
	stats["total_claims"] = totalClaims
	stats["pending_claims"] = pendingClaims
	stats["pending_items"] = pendingItems

	return stats, nil
}

func (r *ItemRepository) CreateClaim(claim *model.Claim) error {
	return database.DB.Create(claim).Error
}

func (r *ItemRepository) FindAllClaims() ([]model.Claim, error) {
	var claims []model.Claim
	err := database.DB.Preload("Item").Preload("Item.User").Preload("User").Order("created_at DESC").Find(&claims).Error
	return claims, err
}

func (r *ItemRepository) UpdateClaim(claim *model.Claim) error {
	return database.DB.Save(claim).Error
}

func (r *ItemRepository) FindClaimByID(id uint) (*model.Claim, error) {
	var claim model.Claim
	err := database.DB.Preload("Item").Preload("Item.User").Preload("User").First(&claim, id).Error
	return &claim, err
}

func (r *ItemRepository) FindByUserID(userID uint) ([]model.Item, error) {
	var items []model.Item
	err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *ItemRepository) FindAllItemsForAdmin() ([]model.Item, error) {
	var items []model.Item
	err := database.DB.Preload("User").Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *ItemRepository) DeleteItem(itemID uint) error {
	return database.DB.Delete(&model.Item{}, itemID).Error
}
