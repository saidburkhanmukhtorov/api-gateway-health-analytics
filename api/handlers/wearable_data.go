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
)

// WearableDataHandler handles requests related to Wearable Data.
type WearableDataHandler struct {
	kafkaProducer *kafka.Producer
	service       health.WearableDataServiceClient
}

// NewWearableDataHandler creates a new WearableDataHandler.
func NewWearableDataHandler(kafkaProducer *kafka.Producer, healthGrpcConn *grpc.ClientConn) *WearableDataHandler {
	return &WearableDataHandler{
		kafkaProducer: kafkaProducer,
		service:       health.NewWearableDataServiceClient(healthGrpcConn),
	}
}

// CreateWearableData godoc
// @Summary     Create a new Wearable Data record
// @Description Create a new wearable data record with the provided details.
// @Tags        WearableData
// @Accept      json
// @Produce     json
// @Param       wearableData body     health.WearableData true "Wearable Data details"
// @Security    ApiKeyAuth
// @Success     202     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/wearable-data [post]
func (h *WearableDataHandler) CreateWearableData(c *gin.Context) {
	var wearableData health.WearableData
	if err := c.ShouldBindJSON(&wearableData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body " + err.Error()})
		return
	}

	// Publish to Kafka
	if err := h.kafkaProducer.ProduceMessage(c.Request.Context(), h.kafkaProducer.Cfg.KafkaWearableDataTopic, "wearable_data.create", &wearableData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create wearable data " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Wearable data creation request accepted"})
}

// GetWearableData godoc
// @Summary     Get Wearable Data by ID
// @Description Get details of a wearable data record by its ID.
// @Tags        WearableData
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Wearable Data ID"
// @Security    ApiKeyAuth
// @Success     200     {object} health.WearableData
// @Failure     400     {object} map[string]interface{}
// @Failure     404     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/wearable-data/{id} [get]
func (h *WearableDataHandler) GetWearableData(c *gin.Context) {
	wearableDataID := c.Param("id")

	// Use gRPC to get the wearable data from the service
	grpcResponse, err := h.service.GetWearableData(context.Background(), &health.ByIdRequest{Id: wearableDataID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Wearable data not found " + err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wearable data " + err.Error()})
		return
	}

	// Convert Any proto message to JSON
	dataValueJSON, err := protojson.Marshal(grpcResponse.DataValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal data value " + err.Error()})
		return
	}

	// Create a response map with the converted data value
	response := map[string]interface{}{
		"id":                 grpcResponse.Id,
		"user_id":            grpcResponse.UserId,
		"device_type":        grpcResponse.DeviceType,
		"data_type":          grpcResponse.DataType,
		"data_value":         string(dataValueJSON),
		"recorded_timestamp": grpcResponse.RecordedTimestamp,
		"created_at":         grpcResponse.CreatedAt,
		"updated_at":         grpcResponse.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateWearableData godoc
// @Summary     Update Wearable Data
// @Description Update an existing wearable data record.
// @Tags        WearableData
// @Accept      json
// @Produce     json
// @Param       id           path     string                   true "Wearable Data ID"
// @Param       wearableData body     health.WearableData true "Updated wearable data"
// @Security    ApiKeyAuth
// @Success     202     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/wearable-data/{id} [put]
func (h *WearableDataHandler) UpdateWearableData(c *gin.Context) {
	wearableDataID := c.Param("id")
	var wearableData health.WearableData
	if err := c.ShouldBindJSON(&wearableData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body " + err.Error()})
		return
	}

	// Ensure the ID in the URL matches the ID in the payload
	if wearableData.Id != wearableDataID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID mismatch"})
		return
	}

	// Publish to Kafka
	if err := h.kafkaProducer.ProduceMessage(c.Request.Context(), h.kafkaProducer.Cfg.KafkaWearableDataTopic, "wearable_data.update", &wearableData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wearable data " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Wearable data update request accepted"})
}

// DeleteWearableData godoc
// @Summary     Delete Wearable Data
// @Description Delete a wearable data record by its ID.
// @Tags        WearableData
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Wearable Data ID"
// @Security    ApiKeyAuth
// @Success     204     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/wearable-data/{id} [delete]
func (h *WearableDataHandler) DeleteWearableData(c *gin.Context) {
	wearableDataID := c.Param("id")

	// Call gRPC service to delete wearable data
	_, err := h.service.DeleteWearableData(context.Background(), &health.ByIdRequest{Id: wearableDataID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Wearable data not found " + err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete wearable data " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Wearable data deleted successfully " + err.Error()})
}

// ListWearableData godoc
// @Summary     List Wearable Data
// @Description Get a list of wearable data records.
// @Tags        WearableData
// @Accept      json
// @Produce     json
// @Param        user_id            query    string false  "Filter by user ID"
// @Param        device_type        query    string false  "Filter by device type"
// @Param        data_type          query    string false  "Filter by data type"
// @Param        recorded_timestamp query string false  "Filter by recorded timestamp (RFC3339 format)"
// @Security    ApiKeyAuth
// @Success     200     {object} health.ListWearableDataResponse
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/wearable-data [get]
func (h *WearableDataHandler) ListWearableData(c *gin.Context) {
	// Get query parameters for pagination and filtering
	userID := c.Query("user_id")
	deviceType := c.Query("device_type")
	dataType := c.Query("data_type")
	recordedTimestamp := c.Query("recorded_timestamp")

	// Use gRPC to get the wearable data from the service
	grpcResponse, err := h.service.ListWearableData(context.Background(), &health.ListWearableDataRequest{
		UserId:            userID,
		DeviceType:        deviceType,
		DataType:          dataType,
		RecordedTimestamp: recordedTimestamp,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wearable data " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, grpcResponse)
}
