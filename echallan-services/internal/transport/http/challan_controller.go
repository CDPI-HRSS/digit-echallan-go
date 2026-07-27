package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/CDPI-HRSS/echallan-services/internal/service"
	"github.com/CDPI-HRSS/echallan-services/internal/transport/kafka"
)

type ChallanController struct {
	challanService service.ChallanService
	producer *kafka.Producer
}

func NewChallanController(service service.ChallanService, producer *kafka.Producer) *ChallanController {
	return &ChallanController{challanService: service, producer: producer}
}

func (cc *ChallanController) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// New RESTful routes
	api := router.Group("/api/v1/challans")
	api.Use(authMiddleware)
	{
		api.POST("", cc.Create)
		api.POST("/search", cc.Search)
		api.PUT("/:id", cc.Update)
		api.POST("/count", cc.Count)
	}
	// Legacy backward-compatible aliases
	legacy := router.Group("/eChallan/v1")
	{
		legacy.POST("/_create", cc.Create)
		legacy.POST("/_search", cc.Search)
		legacy.POST("/_update", cc.Update)
		legacy.POST("/_count", cc.Count)
	}
}

func (cc *ChallanController) Create(c *gin.Context) {
	var req domain.ChallanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:        "VALIDATION_ERROR",
			Message:     "Invalid request payload",
			Description: err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	challan, err := cc.challanService.Create(ctx, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:        "CREATE_ERROR",
			Message:     "Failed to create challan",
			Description: err.Error(),
		})
		return
	}

	res := domain.ChallanResponse{
		ResponseInfo: createResponseInfo(req.RequestInfo, "200 OK"),
		Challans:     []*domain.Challan{challan},
	}
	c.JSON(http.StatusOK, res)
}

func (cc *ChallanController) Search(c *gin.Context) {
	var req domain.RequestInfoWrapper
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:        "VALIDATION_ERROR",
			Message:     "Invalid request payload",
			Description: err.Error(),
		})
		return
	}

	var criteria domain.SearchCriteria
	if err := c.ShouldBindQuery(&criteria); err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:        "VALIDATION_ERROR",
			Message:     "Failed to parse query parameters",
			Description: err.Error(),
		})
		return
	}

	if criteria.TenantId == "" {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:        "INVALID_SEARCH",
			Message:     "tenantId is mandatory for searching",
			Description: "tenantId query parameter is empty",
		})
		return
	}

	ctx := c.Request.Context()
	challans, totalCount, err := cc.challanService.Search(ctx, criteria, req.RequestInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:        "SEARCH_ERROR",
			Message:     "Failed to search challans",
			Description: err.Error(),
		})
		return
	}

	res := domain.ChallanResponse{
		ResponseInfo: createResponseInfo(req.RequestInfo, "200 OK"),
		Challans:     challans,
		TotalCount:   totalCount,
	}
	c.JSON(http.StatusOK, res)
}

func (cc *ChallanController) Update(c *gin.Context) {
	var req domain.ChallanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:        "VALIDATION_ERROR",
			Message:     "Invalid request payload",
			Description: err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	challan, err := cc.challanService.Update(ctx, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:        "UPDATE_ERROR",
			Message:     "Failed to update challan",
			Description: err.Error(),
		})
		return
	}

	res := domain.ChallanResponse{
		ResponseInfo: createResponseInfo(req.RequestInfo, "200 OK"),
		Challans:     []*domain.Challan{challan},
	}
	c.JSON(http.StatusOK, res)
}

func (cc *ChallanController) Count(c *gin.Context) {
	tenantId := c.Query("tenantId")
	var req domain.RequestInfoWrapper
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:        "VALIDATION_ERROR",
			Message:     "Invalid request payload",
			Description: err.Error(),
		})
		return
	}
	var reqInfo domain.RequestInfo
	if req.RequestInfo != nil {
		reqInfo = *req.RequestInfo
	}

	ctx := c.Request.Context()
	countRes, err := cc.challanService.Count(ctx, tenantId, &reqInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:        "COUNT_ERROR",
			Message:     "Failed to count challans",
			Description: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, countRes)
}

func createResponseInfo(reqInfo *domain.RequestInfo, status string) *domain.ResponseInfo {
	if reqInfo == nil {
		return &domain.ResponseInfo{Status: status}
	}
	return &domain.ResponseInfo{
		ApiId:  reqInfo.ApiId,
		Ver:    reqInfo.Ver,
		Ts:     reqInfo.Ts,
		MsgId:  reqInfo.MsgId,
		Status: status,
	}
}

func (cc *ChallanController) Test(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON"})
		return
	}
	if err := cc.producer.Push("update-challan", req); err != nil {
		c.JSON(500, gin.H{"error": "Failed to push to Kafka"})
		return
	}
	c.JSON(200, gin.H{"message": "Successfully pushed to update-challan topic"})
}
