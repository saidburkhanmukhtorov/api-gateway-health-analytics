package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/health-analytics-service/api-gateway-health-analytics/genproto/health"
	"github.com/health-analytics-service/api-gateway-health-analytics/helper"
	"github.com/health-analytics-service/api-gateway-health-analytics/kafka"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HealthRecommendationHandler handles requests related to Health Recommendations.
type HealthRecommendationHandler struct {
	kafkaProducer *kafka.Producer
	service       health.HealthRecommendationServiceClient
}

// NewHealthRecommendationHandler creates a new HealthRecommendationHandler.
func NewHealthRecommendationHandler(kafkaProducer *kafka.Producer, healthGrpcConn *grpc.ClientConn) *HealthRecommendationHandler {
	return &HealthRecommendationHandler{
		kafkaProducer: kafkaProducer,
		service:       health.NewHealthRecommendationServiceClient(healthGrpcConn),
	}
}

// CreateHealthRecommendation godoc
// @Summary     Create a new Health Recommendation record
// @Description Create a new health recommendation record with the provided details.
// @Tags        HealthRecommendations
// @Accept      json
// @Produce     json
// @Param       healthRecommendation body     health.HealthRecommendation true "Health Recommendation details"
// @Security    ApiKeyAuth
// @Success     202     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/health-recommendations [post]
func (h *HealthRecommendationHandler) CreateHealthRecommendation(c *gin.Context) {
	var healthRecommendation health.HealthRecommendation
	if err := c.ShouldBindJSON(&healthRecommendation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body " + err.Error()})
		return
	}

	// Publish to Kafka
	if err := h.kafkaProducer.ProduceMessage(c.Request.Context(), h.kafkaProducer.Cfg.KafkaHealthRecommendationTopic, "health_recommendation.create", &healthRecommendation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create health recommendation " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Health recommendation creation request accepted"})
}

// GetHealthRecommendation godoc
// @Summary     Get Health Recommendation by ID
// @Description Get details of a health recommendation record by its ID.
// @Tags        HealthRecommendations
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Health Recommendation ID"
// @Security    ApiKeyAuth
// @Success     200     {object} health.HealthRecommendation
// @Failure     400     {object} map[string]interface{}
// @Failure     404     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/health-recommendations/{id} [get]
func (h *HealthRecommendationHandler) GetHealthRecommendation(c *gin.Context) {
	healthRecommendationID := c.Param("id")

	// Use gRPC to get the health recommendation from the service
	grpcResponse, err := h.service.GetHealthRecommendation(context.Background(), &health.ByIdRequest{Id: healthRecommendationID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Health recommendation not found " + err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get health recommendation " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, grpcResponse)
}

// UpdateHealthRecommendation godoc
// @Summary     Update Health Recommendation
// @Description Update an existing health recommendation record.
// @Tags        HealthRecommendations
// @Accept      json
// @Produce     json
// @Param       id                   path     string                             true "Health Recommendation ID"
// @Param       healthRecommendation body     health.HealthRecommendation true "Updated health recommendation"
// @Security    ApiKeyAuth
// @Success     202     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/health-recommendations/{id} [put]
func (h *HealthRecommendationHandler) UpdateHealthRecommendation(c *gin.Context) {
	healthRecommendationID := c.Param("id")
	var healthRecommendation health.HealthRecommendation
	if err := c.ShouldBindJSON(&healthRecommendation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body " + err.Error()})
		return
	}

	// Ensure the ID in the URL matches the ID in the payload
	if healthRecommendation.Id != healthRecommendationID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID mismatch"})
		return
	}

	// Publish to Kafka
	if err := h.kafkaProducer.ProduceMessage(c.Request.Context(), h.kafkaProducer.Cfg.KafkaHealthRecommendationTopic, "health_recommendation.update", &healthRecommendation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update health recommendation " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Health recommendation update request accepted"})
}

// DeleteHealthRecommendation godoc
// @Summary     Delete Health Recommendation
// @Description Delete a health recommendation record by its ID.
// @Tags        HealthRecommendations
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Health Recommendation ID"
// @Security    ApiKeyAuth
// @Success     204     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/health-recommendations/{id} [delete]
func (h *HealthRecommendationHandler) DeleteHealthRecommendation(c *gin.Context) {
	healthRecommendationID := c.Param("id")

	// Call gRPC service to delete health recommendation
	_, err := h.service.DeleteHealthRecommendation(context.Background(), &health.ByIdRequest{Id: healthRecommendationID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Health recommendation not found " + err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete health recommendation " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Health recommendation deleted successfully"})
}

// ListHealthRecommendations godoc
// @Summary     List Health Recommendations
// @Description Get a list of health recommendation records.
// @Tags        HealthRecommendations
// @Accept      json
// @Produce     json
// @Param        user_id             query    string false  "Filter by user ID"
// @Param        recommendation_type query    string false  "Filter by recommendation type"
// @Param        priority            query    int32   false  "Filter by priority"
// @Security    ApiKeyAuth
// @Success     200     {object} health.ListHealthRecommendationsResponse
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/health-recommendations [get]
func (h *HealthRecommendationHandler) ListHealthRecommendations(c *gin.Context) {
	// Get query parameters for pagination and filtering
	userID := c.Query("user_id")
	recommendationType := c.Query("recommendation_type")
	priority := helper.StringToInt(c.Query("priority"))

	// Use gRPC to get the health recommendations from the service
	grpcResponse, err := h.service.ListHealthRecommendations(context.Background(), &health.ListHealthRecommendationsRequest{
		UserId:             userID,
		RecommendationType: recommendationType,
		Priority:           priority,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get health recommendations " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, grpcResponse)
}
