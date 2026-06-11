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

func (cc *ChallanController) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/eChallan/v1")
	{
		v1.POST("/_create", cc.Create)
		v1.POST("/_search", cc.Search)
		v1.POST("/_update", cc.Update)
		v1.POST("/_count", cc.Count)
		v1.POST("/_test", cc.Test)
	}
}

func (cc *ChallanController) Create(c *gin.Context) {
	var req domain.ChallanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:    "INVALID_REQUEST",
			Message: "Failed to parse JSON payload",
		})
		return
	}

	challan, err := cc.challanService.Create(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:    "CREATE_ERROR",
			Message: err.Error(),
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
			Code:    "INVALID_REQUEST",
			Message: "Failed to parse JSON payload",
		})
		return
	}

	var criteria domain.SearchCriteria
	if err := c.ShouldBindQuery(&criteria); err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:    "INVALID_QUERY",
			Message: "Failed to parse query parameters",
		})
		return
	}

	if criteria.TenantId == "" {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:    "INVALID_SEARCH",
			Message: "tenantId is mandatory for searching",
		})
		return
	}

	challans, totalCount, err := cc.challanService.Search(criteria, req.RequestInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:    "SEARCH_ERROR",
			Message: err.Error(),
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
			Code:    "INVALID_REQUEST",
			Message: "Failed to parse JSON payload",
		})
		return
	}

	challan, err := cc.challanService.Update(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:    "UPDATE_ERROR",
			Message: err.Error(),
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
	var reqInfo domain.RequestInfo
	if err := c.ShouldBindJSON(&reqInfo); err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:    "INVALID_REQUEST",
			Message: "Failed to parse JSON payload",
		})
		return
	}

	countRes, err := cc.challanService.Count(tenantId, &reqInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Error{
			Code:    "COUNT_ERROR",
			Message: err.Error(),
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
