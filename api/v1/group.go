package v1

import "github.com/gin-gonic/gin"

// Group adds group of routes for APIv1
func Group(route *gin.Engine) {
	group := route.Group("/api/v1")
	{
		group.POST("/review", APIReview)
	}
}
