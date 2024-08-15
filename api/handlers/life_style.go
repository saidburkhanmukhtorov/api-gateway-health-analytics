package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/health-analytics-service/api-gateway-health-analytics/genproto/health"
	"github.com/health-analytics-service/api-gateway-health-analytics/kafka"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
)

// LifestyleDataHandler handles requests related to Lifestyle Data.
type LifestyleDataHandler struct {
	kafkaProducer *kafka.Producer
	service       health.LifestyleDataServiceClient
}

// NewLifestyleDataHandler creates a new LifestyleDataHandler.
func NewLifestyleDataHandler(kafkaProducer *kafka.Producer, healthGrpcConn *grpc.ClientConn) *LifestyleDataHandler {
	return &LifestyleDataHandler{
		kafkaProducer: kafkaProducer,
		service:       health.NewLifestyleDataServiceClient(healthGrpcConn),
	}
}

// CreateLifestyleData godoc
// @Summary     Create a new Lifestyle Data record
// @Description Create a new lifestyle data record with the provided details.
// @Tags        LifestyleData
// @Accept      json
// @Produce     json
// @Param       lifestyleData body     health.LifestyleData true "Lifestyle Data details"
// @Security    ApiKeyAuth
// @Success     202     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/lifestyle-data [post]
func (h *LifestyleDataHandler) CreateLifestyleData(c *gin.Context) {
	var lifestyleData health.LifestyleData
	if err := c.ShouldBindJSON(&lifestyleData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Publish to Kafka
	if err := h.kafkaProducer.ProduceMessage(c.Request.Context(), h.kafkaProducer.Cfg.KafkaLifestyleDataTopic, "lifestyle_data.create", &lifestyleData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lifestyle data " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Lifestyle data creation request accepted"})
}

// GetLifestyleData godoc
// @Summary     Get Lifestyle Data by ID
// @Description Get details of a lifestyle data record by its ID.
// @Tags        LifestyleData
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Lifestyle Data ID"
// @Security    ApiKeyAuth
// @Success     200     {object} health.LifestyleData
// @Failure     400     {object} map[string]interface{}
// @Failure     404     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/lifestyle-data/{id} [get]
func (h *LifestyleDataHandler) GetLifestyleData(c *gin.Context) {
	lifestyleDataID := c.Param("id")

	// Use gRPC to get the lifestyle data from the service
	grpcResponse, err := h.service.GetLifestyleData(context.Background(), &health.ByIdRequest{Id: lifestyleDataID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Lifestyle data not found " + err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get lifestyle data " + err.Error()})
		return
	}

	// Convert Any proto message to JSON
	dataValueJSON, err := protojson.Marshal(grpcResponse.DataValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal data value"})
		return
	}

	// Create a response map with the converted data value
	response := map[string]interface{}{
		"id":            grpcResponse.Id,
		"user_id":       grpcResponse.UserId,
		"data_type":     grpcResponse.DataType,
		"data_value":    string(dataValueJSON),
		"recorded_date": grpcResponse.RecordedDate,
		"created_at":    grpcResponse.CreatedAt,
		"updated_at":    grpcResponse.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateLifestyleData godoc
// @Summary     Update Lifestyle Data
// @Description Update an existing lifestyle data record.
// @Tags        LifestyleData
// @Accept      json
// @Produce     json
// @Param       id           path     string                   true "Lifestyle Data ID"
// @Param       lifestyleData body     health.LifestyleData true "Updated lifestyle data"
// @Security    ApiKeyAuth
// @Success     202     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/lifestyle-data/{id} [put]
func (h *LifestyleDataHandler) UpdateLifestyleData(c *gin.Context) {
	lifestyleDataID := c.Param("id")
	var lifestyleData health.LifestyleData
	if err := c.ShouldBindJSON(&lifestyleData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body " + err.Error()})
		return
	}

	// Ensure the ID in the URL matches the ID in the payload
	if lifestyleData.Id != lifestyleDataID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID mismatch"})
		return
	}

	// Publish to Kafka
	if err := h.kafkaProducer.ProduceMessage(c.Request.Context(), h.kafkaProducer.Cfg.KafkaLifestyleDataTopic, "lifestyle_data.update", &lifestyleData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lifestyle data " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Lifestyle data update request accepted"})
}

// DeleteLifestyleData godoc
// @Summary     Delete Lifestyle Data
// @Description Delete a lifestyle data record by its ID.
// @Tags        LifestyleData
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Lifestyle Data ID"
// @Security    ApiKeyAuth
// @Success     204     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/lifestyle-data/{id} [delete]
func (h *LifestyleDataHandler) DeleteLifestyleData(c *gin.Context) {
	lifestyleDataID := c.Param("id")

	// Call gRPC service to delete lifestyle data
	_, err := h.service.DeleteLifestyleData(context.Background(), &health.ByIdRequest{Id: lifestyleDataID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Lifestyle data not found " + err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete lifestyle data " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Lifestyle data deleted successfully"})
}

// ListLifestyleData godoc
// @Summary     List Lifestyle Data
// @Description Get a list of lifestyle data records.
// @Tags        LifestyleData
// @Accept      json
// @Produce     json
// @Param        user_id     query    string false  "Filter by user ID"
// @Param        data_type   query    string false  "Filter by data type"
// @Param        recorded_date query string false  "Filter by recorded date (YYYY-MM-DD)"
// @Security    ApiKeyAuth
// @Success     200     {object} health.ListLifestyleDataResponse
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/lifestyle-data [get]
func (h *LifestyleDataHandler) ListLifestyleData(c *gin.Context) {
	// Get query parameters for pagination and filtering
	userID := c.Query("user_id")
	dataType := c.Query("data_type")
	recordedDate := c.Query("recorded_date")

	// Use gRPC to get the lifestyle data from the service
	grpcResponse, err := h.service.ListLifestyleData(context.Background(), &health.ListLifestyleDataRequest{
		UserId:       userID,
		DataType:     dataType,
		RecordedDate: recordedDate,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get lifestyle data " + err.Error()})
		return
	}

	for i, data := range grpcResponse.LifestyleData {
		// Convert Any proto message to JSON
		dataValueJSON, err := protojson.Marshal(data.DataValue)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal data value " + err.Error()})
			return
		}
		grpcResponse.LifestyleData[i].DataValue = &anypb.Any{
			Value: dataValueJSON,
		}
	}

	c.JSON(http.StatusOK, grpcResponse)
}
