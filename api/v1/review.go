package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tczekajlo/kir/policy"
	"github.com/tczekajlo/kir/types"
)

// APIReview is handler to make image review
func APIReview(c *gin.Context) {
	var json types.ImageReview

	err := c.BindJSON(&json)
	if err == nil {
		c.JSON(http.StatusOK, policy.Review(&json))
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s", err)})
	}
}
