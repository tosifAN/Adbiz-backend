package handlers

import (
	"adbiz_backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FavDealRequest represents the request structure for favorite operations
type FavDealRequest struct {
	CurrentUserMobile string `json:"current_user_mobile" binding:"required"`
	TargetUserMobile  string `json:"target_user_mobile" binding:"required"`
}

// HandleFav processes when a user favorites another user
// It updates both Fav1 (following) and Fav2 (followers) tables
func (h *FavHandler) HandleFav(c *gin.Context) {
	<-h.rateLimit.C

	var req FavDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user (the one who is doing the favoriting)
	var currentUser models.User
	if err := h.db.Where("mobile_number = ?", req.CurrentUserMobile).First(&currentUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Current user not found"})
		return
	}

	// Get target user (the one being favorited)
	var targetUser models.User
	if err := h.db.Where("mobile_number = ?", req.TargetUserMobile).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target user not found"})
		return
	}

	// Begin transaction
	tx := h.db.Begin()

	// Update Fav1 (following) for current user
	if err := h.updateFav1(tx, currentUser.ID, req.TargetUserMobile); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update following: " + err.Error()})
		return
	}

	// Update Fav2 (followers) for target user
	if err := h.updateFav2(tx, targetUser.ID, req.CurrentUserMobile); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update followers: " + err.Error()})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Favorite updated successfully"})
}

// updateFav1 updates the Fav1 (following) record for a user
func (h *FavHandler) updateFav1(tx *gorm.DB, userID uint, targetMobile string) error {
	// Check if Fav1 record exists for the user
	var fav1 models.Fav1
	result := tx.Where("user_id = ?", userID).First(&fav1)

	if result.Error != nil {
		// Create new Fav1 record if it doesn't exist
		fav1 = models.Fav1{
			Fav:     1,
			FavList: []string{targetMobile},
			UserID:  userID,
		}
		return tx.Create(&fav1).Error
	}

	// Check if target mobile is already in the FavList
	for _, mobile := range fav1.FavList {
		if mobile == targetMobile {
			// Already favorited, no need to update
			return nil
		}
	}

	// Update existing Fav1 record
	fav1.Fav++
	fav1.FavList = append(fav1.FavList, targetMobile)
	return tx.Save(&fav1).Error
}

// updateFav2 updates the Fav2 (followers) record for a user
func (h *FavHandler) updateFav2(tx *gorm.DB, userID uint, followerMobile string) error {
	// Check if Fav2 record exists for the user
	var fav2 models.Fav2
	result := tx.Where("user_id = ?", userID).First(&fav2)

	if result.Error != nil {
		// Create new Fav2 record if it doesn't exist
		fav2 = models.Fav2{
			Fav:     1,
			FavList: []string{followerMobile},
			UserID:  userID,
		}
		return tx.Create(&fav2).Error
	}

	// Check if follower mobile is already in the FavList
	for _, mobile := range fav2.FavList {
		if mobile == followerMobile {
			// Already in followers, no need to update
			return nil
		}
	}

	// Update existing Fav2 record
	fav2.Fav++
	fav2.FavList = append(fav2.FavList, followerMobile)
	return tx.Save(&fav2).Error
}
