package service

import (
	"errors"
	"fmt"
	"io"
	"lostfound/internal/model"
	"lostfound/internal/repository"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
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

	if file.Size > 5*1024*1024 {
		return "", errors.New("image is too large (max 5MB)")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", errors.New("only JPG and PNG images are allowed")
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
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
	if itemType != "lost" && itemType != "found" {
		return nil, errors.New("invalid item type")
	}
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("item title is required")
	}

	item := &model.Item{
		UserID:         userID,
		Type:           itemType,
		Title:          title,
		Category:       category,
		Color:          color,
		Brand:          brand,
		Location:       location,
		Date:           date,
		Description:    description,
		Image:          imagePath,
		Status:         "open",
		ApprovalStatus: "pending",
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
	item, err := s.itemRepo.FindByID(itemID)
	if err != nil {
		return err
	}
	if item.UserID == userID {
		return errors.New("you cannot claim your own post")
	}
	if item.ApprovalStatus != "approved" {
		return errors.New("item is not approved by admin yet")
	}
	if item.Type != "found" {
		return errors.New("you can only claim found items")
	}
	if item.Status != "open" {
		return errors.New("this item is no longer available for claim")
	}
	if strings.TrimSpace(description) == "" {
		return errors.New("claim description is required")
	}

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
	if status != "approved" && status != "rejected" {
		return errors.New("invalid claim status")
	}

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
			_ = s.itemRepo.Update(item)
		}
	}

	return s.itemRepo.UpdateClaim(claim)
}

func (s *ItemService) GetItemsByUserID(userID uint) ([]model.Item, error) {
	return s.itemRepo.FindByUserID(userID)
}

func (s *ItemService) GetAllItemsForAdmin() ([]model.Item, error) {
	return s.itemRepo.FindAllItemsForAdmin()
}

func (s *ItemService) UpdateItemApproval(itemID uint, status, remarks string) error {
	if status != "approved" && status != "rejected" && status != "pending" {
		return errors.New("invalid approval status")
	}

	item, err := s.itemRepo.FindByID(itemID)
	if err != nil {
		return err
	}

	item.ApprovalStatus = status
	item.AdminRemarks = remarks
	if status != "approved" {
		item.Status = "open"
	}
	return s.itemRepo.Update(item)
}

func (s *ItemService) DeleteItem(itemID uint) error {
	return s.itemRepo.DeleteItem(itemID)
}
