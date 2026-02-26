// internal/service/item_service.go
package service

import (
    "errors"
    "lostfound/internal/model"
    "lostfound/internal/repository"
    "mime/multipart"
    "os"
    "path/filepath"
    "time"
)

type ItemService struct {
    itemRepo *repository.ItemRepository
}

func NewItemService() *ItemService {
    return &ItemService{
        itemRepo: repository.NewItemRepository(),
    }
}

// Save uploaded image
func (s *ItemService) SaveImage(file *multipart.FileHeader) (string, error) {
    // Create uploads directory if not exists
    uploadDir := "static/uploads"
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        return "", err
    }
    
    // Generate unique filename
    filename := time.Now().Format("20060102150405") + "_" + file.Filename
    filepath := filepath.Join(uploadDir, filename)
    
    // Save file
    if err := c.SaveUploadedFile(file, filepath); err != nil {
        return "", err
    }
    
    return "/static/uploads/" + filename, nil
}

// Create item
func (s *ItemService) CreateItem(userID uint, itemType, title, category, color, brand, location, date, description string, imagePath string) (*model.Item, error) {
    item := &model.Item{
        UserID:      userID,
        Type:        itemType,
        Title:       title,
        Category:    category,
        Color:       color,
        Brand:       brand,
        Location:    location,
        Date:        date,
        Description: description,
        Image:       imagePath,
        Status:      "open",
    }
    
    err := s.itemRepo.Create(item)
    return item, err
}

// Search items
func (s *ItemService) SearchItems(filters map[string]interface{}) ([]model.Item, error) {
    return s.itemRepo.FindAll(filters)
}

// Get item by ID
func (s *ItemService) GetItemByID(id uint) (*model.Item, error) {
    return s.itemRepo.FindByID(id)
}

// Create claim
func (s *ItemService) CreateClaim(itemID, userID uint, description string) error {
    // Check if already claimed
    // This is simplified - you might want to check existing claims
    
    claim := &model.Claim{
        ItemID:      itemID,
        UserID:      userID,
        Description: description,
        Status:      "pending",
    }
    
    return s.itemRepo.CreateClaim(claim)
}

// Get stats
func (s *ItemService) GetStats() (map[string]int64, error) {
    return s.itemRepo.GetStats()
}

// Admin: Get all claims
func (s *ItemService) GetAllClaims() ([]model.Claim, error) {
    return s.itemRepo.FindAllClaims()
}

// Admin: Update claim status
func (s *ItemService) UpdateClaimStatus(claimID uint, status, remarks string) error {
    claim, err := s.itemRepo.FindClaimByID(claimID)
    if err != nil {
        return err
    }
    
    claim.Status = status
    claim.AdminRemarks = remarks
    
    // If approved, update item status
    if status == "approved" {
        item, err := s.itemRepo.FindByID(claim.ItemID)
        if err == nil {
            item.Status = "claimed"
            s.itemRepo.Update(item)
        }
    }
    
    return s.itemRepo.UpdateClaim(claim)
}