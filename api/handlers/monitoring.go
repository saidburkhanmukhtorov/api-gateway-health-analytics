package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/health-analytics-service/api-gateway-health-analytics/genproto/health"
	"google.golang.org/grpc"
)

// HealthMonitoringHandler handles requests related to Health Monitoring.
type HealthMonitoringHandler struct {
	service health.HealthMonitoringServiceClient
}

// NewHealthMonitoringHandler creates a new HealthMonitoringHandler.
func NewHealthMonitoringHandler(healthGrpcConn *grpc.ClientConn) *HealthMonitoringHandler {
	return &HealthMonitoringHandler{
		service: health.NewHealthMonitoringServiceClient(healthGrpcConn),
	}
}

// GetDailySummary godoc
// @Summary     Get Daily Summary
// @Description Get a daily summary of health data for a user.
// @Tags        HealthMonitoring
// @Accept      json
// @Produce     json
// @Param       user_id path     string true "User ID"
// @Param       date    query    string true "Date (YYYY-MM-DD)"
// @Security    ApiKeyAuth
// @Success     200     {object} health.SummaryResponse
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/health-monitoring/daily-summary/{user_id} [get]
func (h *HealthMonitoringHandler) GetDailySummary(c *gin.Context) {
	userID := c.Param("user_id")
	date := c.Query("date")

	// Use gRPC to get the daily summary from the service
	grpcResponse, err := h.service.GetDailySummary(context.Background(), &health.DailySummaryRequest{
		UserId: userID,
		Date:   date,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get daily summary " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, grpcResponse)
}

// GetWeeklySummary godoc
// @Summary     Get Weekly Summary
// @Description Get a weekly summary of health data for a user.
// @Tags        HealthMonitoring
// @Accept      json
// @Produce     json
// @Param       user_id   path     string true "User ID"
// @Param       start_date query    string true "Start Date (YYYY-MM-DD)"
// @Param       end_date   query    string true "End Date (YYYY-MM-DD)"
// @Security    ApiKeyAuth
// @Success     200     {object} health.SummaryResponse
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/health-monitoring/weekly-summary/{user_id} [get]
func (h *HealthMonitoringHandler) GetWeeklySummary(c *gin.Context) {
	userID := c.Param("user_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Use gRPC to get the weekly summary from the service
	grpcResponse, err := h.service.GetWeeklySummary(context.Background(), &health.WeeklySummaryRequest{
		UserId:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get weekly summary " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, grpcResponse)
}
