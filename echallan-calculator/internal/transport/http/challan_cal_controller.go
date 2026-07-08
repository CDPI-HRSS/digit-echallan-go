package controllers

import (
	"net/http"

	"github.com/CDPI-HRSS/calci_sp/internal/domain"
	"github.com/CDPI-HRSS/calci_sp/internal/service"
	"github.com/gin-gonic/gin"
)

type ChallanCalController struct {
	calcService *service.CalculationService
}

func NewChallanCalController(calcService *service.CalculationService) *ChallanCalController {
	return &ChallanCalController{
		calcService: calcService,
	}
}

func (ctrl *ChallanCalController) RegisterRoutes(r *gin.Engine, contextPath string) {
	group := r.Group(contextPath)
	{
		group.POST("/_calculate", ctrl.Calculate)
		group.POST("/_calculate/:servicename", ctrl.Calculate)
	}
}

func (ctrl *ChallanCalController) Calculate(c *gin.Context) {
	var req domain.CalculationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorRes{
			Errors: []domain.Error{
				{
					Code:        "INVALID_REQUEST",
					Message:     "Failed to parse request JSON",
					Description: err.Error(),
				},
			},
		})
		return
	}

	calculations, err := ctrl.calcService.GetCalculation(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorRes{
			Errors: []domain.Error{
				{
					Code:    "CALCULATION_ERROR",
					Message: err.Error(),
				},
			},
		})
		return
	}

	var resInfo *domain.ResponseInfo
	if req.RequestInfo != nil {
		resInfo = &domain.ResponseInfo{
			APIId:  req.RequestInfo.APIId,
			Ver:    req.RequestInfo.Ver,
			Ts:     req.RequestInfo.Ts,
			MsgId:  req.RequestInfo.MsgId,
			Status: "successful",
		}
	}

	res := domain.CalculationRes{
		ResponseInfo: resInfo,
		Calculations: calculations,
	}

	c.JSON(http.StatusOK, res)
}
