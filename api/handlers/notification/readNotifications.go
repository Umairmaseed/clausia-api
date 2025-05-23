package notification

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/umairmaseed/clausia-api/api/handlers/errorhandler"
	"github.com/umairmaseed/clausia-api/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReadNotificationsForm struct {
	Id string `form:"id" binding:"required"`
}

func ReadNotifications(c *gin.Context) {
	var form ReadNotificationsForm
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(form.Id)
	if err != nil {
		errorhandler.ReturnError(c, err, "Invalid ID format", http.StatusBadRequest)
		return
	}

	nerr := db.NewNotificationService(db.GetDB().Database()).MarkNotificationAsRead(c.Request.Context(), objectID)
	if nerr != nil {
		errorhandler.ReturnError(c, nerr, "failed to generate notification", http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})

}
