package service

import (
    // "errors"
    "io"
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

func (s *ItemService) SaveImage(file *multipart.FileHeader) (string, error) {
    uploadDir := "static/uploads"
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        return "", err
    }
    
    filename := time.Now().Format("20060102150405") + "_" + file.Filename
    filepath := filepath.Join(uploadDir, filename)
    
    // Open the uploaded file
    src, err := file.Open()
    if err != nil {
        return "", err
    }
    defer src.Close()
    
    // Create destination file
    dst, err := os.Create(filepath)
    if err != nil {
        return "", err
    }
    defer dst.Close()
    
    // Copy the file
    if _, err = io.Copy(dst, src); err != nil {
        return "", err
    }
    
    return "/static/uploads/" + filename, nil
}

func (s *ItemService) CreateItem(userID uint, itemType, title, category, color, brand, location, date, description, imagePath string) (*model.Item, error) {
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

func (s *ItemService) SearchItems(filters map[string]interface{}) ([]model.Item, error) {
    return s.itemRepo.FindAll(filters)
}

func (s *ItemService) GetItemByID(id uint) (*model.Item, error) {
    return s.itemRepo.FindByID(id)
}

func (s *ItemService) CreateClaim(itemID, userID uint, description string) error {
    claim := &model.Claim{
        ItemID:      itemID,
        UserID:      userID,
        Description: description,
        Status:      "pending",
    }
    
    return s.itemRepo.CreateClaim(claim)
}

func (s *ItemService) GetStats() (map[string]int64, error) {
    return s.itemRepo.GetStats()
}

func (s *ItemService) GetAllClaims() ([]model.Claim, error) {
    return s.itemRepo.FindAllClaims()
}

func (s *ItemService) UpdateClaimStatus(claimID uint, status, remarks string) error {
    claim, err := s.itemRepo.FindClaimByID(claimID)
    if err != nil {
        return err
    }
    
    claim.Status = status
    claim.AdminRemarks = remarks
    
    if status == "approved" {
        item, err := s.itemRepo.FindByID(claim.ItemID)
        if err == nil {
            item.Status = "claimed"
            s.itemRepo.Update(item)
        }
    }
    
    return s.itemRepo.UpdateClaim(claim)
}