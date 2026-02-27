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
	if !IsValidItemType(itemType) {
		return nil, errors.New("invalid item type")
	}
	if !IsValidCategory(category) {
		return nil, errors.New("invalid category")
	}
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("item title is required")
	}
	if strings.TrimSpace(date) != "" {
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return nil, errors.New("invalid date format")
		}
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
	if item.Status != "open" {
		return errors.New("this item is no longer open for request")
	}
	if strings.TrimSpace(description) == "" {
		return errors.New("request description is required")
	}

	hasActive, err := s.itemRepo.HasActiveClaimByUser(itemID, userID)
	if err != nil {
		return err
	}
	if hasActive {
		return errors.New("you already have a pending or approved request for this post")
	}

	requestType := "claim_request"
	if item.Type == "lost" {
		requestType = "found_match_request"
	}

	claim := &model.Claim{
		ItemID:      itemID,
		UserID:      userID,
		RequestType: requestType,
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

		requestTypeLabel := "Claim Request"
		if claim.RequestType == "found_match_request" {
			requestTypeLabel = "Found Match Request"
		}

		posterContact := fmt.Sprintf("%s (%s / %s)", claim.Item.User.Name, claim.Item.User.StudentID, claim.Item.User.Phone)
		requesterContact := fmt.Sprintf("%s (%s / %s)", claim.User.Name, claim.User.StudentID, claim.User.Phone)

		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: claim.UserID,
			Title:  "Request Approved",
			Message: fmt.Sprintf(
				"Admin approved your %s for \"%s\". Contact post owner: %s.",
				requestTypeLabel, claim.Item.Title, posterContact,
			),
		})
		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: claim.Item.UserID,
			Title:  "Request Approved On Your Post",
			Message: fmt.Sprintf(
				"Admin approved %s on your post \"%s\". Contact requester: %s.",
				requestTypeLabel, claim.Item.Title, requesterContact,
			),
		})
	} else {
		reason := strings.TrimSpace(remarks)
		if reason == "" {
			reason = "No remarks provided"
		}
		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: claim.UserID,
			Title:  "Request Rejected",
			Message: fmt.Sprintf(
				"Admin rejected your request for \"%s\". Remarks: %s",
				claim.Item.Title, reason,
			),
		})
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
	if !IsValidApprovalStatus(status) {
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
	if err := s.itemRepo.Update(item); err != nil {
		return err
	}

	if status == "approved" {
		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: item.UserID,
			Title:  "Post Approved",
			Message: fmt.Sprintf(
				"Admin approved your %s post \"%s\". It is now visible in explore/search.",
				item.Type, item.Title,
			),
		})
	}
	if status == "rejected" {
		reason := strings.TrimSpace(remarks)
		if reason == "" {
			reason = "No remarks provided"
		}
		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: item.UserID,
			Title:  "Post Rejected",
			Message: fmt.Sprintf(
				"Admin rejected your %s post \"%s\". Remarks: %s",
				item.Type, item.Title, reason,
			),
		})
	}

	return nil
}

func (s *ItemService) DeleteItem(itemID uint) error {
	return s.itemRepo.DeleteItem(itemID)
}

func (s *ItemService) HasApprovedRequestForUser(itemID, userID uint) bool {
	ok, err := s.itemRepo.HasApprovedClaimForUser(itemID, userID)
	if err != nil {
		return false
	}
	return ok
}

func (s *ItemService) GetNotificationsByUserID(userID uint) ([]model.Notification, error) {
	return s.itemRepo.FindNotificationsByUserID(userID)
}

func (s *ItemService) MarkNotificationsRead(userID uint) error {
	return s.itemRepo.MarkNotificationsRead(userID)
}

func (s *ItemService) CountUnreadNotifications(userID uint) int64 {
	count, err := s.itemRepo.CountUnreadNotifications(userID)
	if err != nil {
		return 0
	}
	return count
}
