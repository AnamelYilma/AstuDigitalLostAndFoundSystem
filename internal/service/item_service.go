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

	"github.com/gabriel-vasile/mimetype"
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

	// Validate actual MIME type
	head := make([]byte, 512)
	n, _ := src.Read(head)
	mime := mimetype.Detect(head[:n])
	if !(mime.Is("image/jpeg") || mime.Is("image/png")) {
		return "", errors.New("invalid image type")
	}
	if seeker, ok := src.(io.Seeker); ok {
		_, _ = seeker.Seek(0, io.SeekStart)
	}

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

func (s *ItemService) CreateItem(userID uint, itemType, title, category, color, brand, location, date, description string, imagePaths []string) (*model.Item, error) {
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
		if parsed, err := time.Parse("2006-01-02", date); err != nil {
			return nil, errors.New("invalid date format")
		} else if parsed.After(time.Now()) {
			return nil, errors.New("date cannot be in the future")
		}
	}
	primaryImage := ""
	if len(imagePaths) > 0 {
		primaryImage = imagePaths[0]
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
		Image:          primaryImage,
		Status:         "open",
		ApprovalStatus: "pending",
	}

	if err := s.itemRepo.Create(item); err != nil {
		return nil, err
	}
	if err := s.itemRepo.CreateItemImages(item.ID, imagePaths); err != nil {
		return nil, err
	}

	// Notify admins to review the new post (pending).
	_ = s.notifyAdmins(
		"New Post Pending Approval",
		fmt.Sprintf("A %s item was submitted and needs review.\n\nTitle: %s\nCategory: %s\nLocation: %s\nReported by: %d (user id)\n\nOpen Admin > Items to approve/reject.",
			itemType, title, category, location, userID),
	)

	return item, nil
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

	if err := s.itemRepo.CreateClaim(claim); err != nil {
		return err
	}

	// Notify admins of the incoming request so they don't miss approvals.
	requestTypeLabel := "Claim Request"
	if requestType == "found_match_request" {
		requestTypeLabel = "Found Match Request"
	}
	_ = s.notifyAdmins(
		"New Request Pending",
		fmt.Sprintf(
			"%s submitted for \"%s\".\n\nPost owner: %s (%s / %s)\nRequester: %s (%s / %s)\n\nPlease review in Admin > Claims.",
			requestTypeLabel,
			item.Title,
			item.User.Name, item.User.StudentID, item.User.Phone,
			claim.User.Name, claim.User.StudentID, claim.User.Phone,
		),
	)

	return nil
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

		posterName := claim.Item.User.Name
		posterID := claim.Item.User.StudentID
		posterPhone := strings.TrimSpace(claim.Item.User.Phone)
		requesterName := claim.User.Name
		requesterID := claim.User.StudentID
		requesterPhone := strings.TrimSpace(claim.User.Phone)

		requesterBlock := fmt.Sprintf("Contact requester:\nName: %s (%s)\nPhone: %s", requesterName, requesterID, requesterPhone)
		posterBlock := fmt.Sprintf("Contact post owner:\nName: %s (%s)\nPhone: %s", posterName, posterID, posterPhone)
		selfBlockForRequester := fmt.Sprintf("Your details (shared):\nName: %s (%s)\nPhone: %s", requesterName, requesterID, requesterPhone)
		selfBlockForPoster := fmt.Sprintf("Your details (shared):\nName: %s (%s)\nPhone: %s", posterName, posterID, posterPhone)

		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: claim.UserID,
			Title:  "Request Approved",
			Message: fmt.Sprintf(
				"Admin approved your %s for \"%s\".\n\n%s\n\n%s\n\nPlease meet in a safe, public spot on campus and confirm the item details together.",
				requestTypeLabel, claim.Item.Title, posterBlock, selfBlockForRequester,
			),
		})
		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: claim.Item.UserID,
			Title:  "Request Approved On Your Post",
			Message: fmt.Sprintf(
				"Admin approved %s on your post \"%s\".\n\n%s\n\n%s\n\nCoordinate handoff safely and mark the item resolved after exchange.",
				requestTypeLabel, claim.Item.Title, requesterBlock, selfBlockForPoster,
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
				"Admin rejected your request for \"%s\".\n\nRemarks: %s",
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
				"Admin approved your %s post \"%s\". It is now visible in the Report list.",
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
	item, err := s.itemRepo.FindByID(itemID)
	if err != nil {
		return err
	}

	seen := map[string]bool{}
	for _, img := range item.Images {
		_ = deleteUploadedFile(img.Path, seen)
	}
	if item.Image != "" {
		_ = deleteUploadedFile(item.Image, seen)
	}

	_ = s.itemRepo.DeleteItemImages(itemID)
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

func (s *ItemService) notifyAdmins(title, message string) error {
	admins, err := s.itemRepo.FindAdmins()
	if err != nil {
		return err
	}
	for _, admin := range admins {
		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID:  admin.ID,
			Title:   title,
			Message: message,
		})
	}
	return nil
}

func deleteUploadedFile(path string, seen map[string]bool) error {
	clean := strings.TrimPrefix(path, "/")
	clean = filepath.Clean(clean)
	if !strings.HasPrefix(clean, "static/uploads") {
		return nil
	}
	if seen[clean] {
		return nil
	}
	seen[clean] = true
	return os.Remove(clean)
}
