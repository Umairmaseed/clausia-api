package notification

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type deleteNotification struct {
	Id string `form:"id" binding:"required"`
}

func DeleteNotification(c *gin.Context) {
	var form deleteNotification
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(form.Id)
	if err != nil {
		errorhandler.ReturnError(c, err, "Invalid ID format", http.StatusBadRequest)
		return
	}

	nerr := db.NewNotificationService(db.GetDB().Database()).DeleteNotification(c.Request.Context(), objectID)
	if nerr != nil {
		errorhandler.ReturnError(c, nerr, "failed to generate notification", http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification deleted successfully"})

}
