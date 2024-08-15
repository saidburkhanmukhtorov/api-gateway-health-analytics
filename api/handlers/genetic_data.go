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

// GeneticDataHandler handles requests related to Genetic Data.
type GeneticDataHandler struct {
	kafkaProducer *kafka.Producer
	service       health.GeneticDataServiceClient
}

// NewGeneticDataHandler creates a new GeneticDataHandler.
func NewGeneticDataHandler(kafkaProducer *kafka.Producer, healthGrpcConn *grpc.ClientConn) *GeneticDataHandler {
	return &GeneticDataHandler{
		kafkaProducer: kafkaProducer,
		service:       health.NewGeneticDataServiceClient(healthGrpcConn),
	}
}

// CreateGeneticData godoc
// @Summary     Create a new Genetic Data record
// @Description Create a new genetic data record with the provided details.
// @Tags        GeneticData
// @Accept      json
// @Produce     json
// @Param       geneticData body     health.GeneticData true "Genetic Data details"
// @Security    ApiKeyAuth
// @Success     202     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/genetic-data [post]
func (h *GeneticDataHandler) CreateGeneticData(c *gin.Context) {
	var geneticData health.GeneticData
	if err := c.ShouldBindJSON(&geneticData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Publish to Kafka
	if err := h.kafkaProducer.ProduceMessage(c.Request.Context(), h.kafkaProducer.Cfg.KafkaGeneticDataTopic, "genetic_data.create", &geneticData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create genetic data " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Genetic data creation request accepted"})
}

// GetGeneticData godoc
// @Summary     Get Genetic Data by ID
// @Description Get details of a genetic data record by its ID.
// @Tags        GeneticData
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Genetic Data ID"
// @Security    ApiKeyAuth
// @Success     200     {object} health.GeneticData
// @Failure     400     {object} map[string]interface{}
// @Failure     404     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/genetic-data/{id} [get]
func (h *GeneticDataHandler) GetGeneticData(c *gin.Context) {
	geneticDataID := c.Param("id")

	// Use gRPC to get the genetic data from the service
	grpcResponse, err := h.service.GetGeneticData(context.Background(), &health.ByIdRequest{Id: geneticDataID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Genetic data not found " + err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get genetic data" + err.Error()})
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
		"id":            grpcResponse.Id,
		"user_id":       grpcResponse.UserId,
		"data_type":     grpcResponse.DataType,
		"data_value":    string(dataValueJSON),
		"analysis_date": grpcResponse.AnalysisDate,
		"created_at":    grpcResponse.CreatedAt,
		"updated_at":    grpcResponse.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateGeneticData godoc
// @Summary     Update Genetic Data
// @Description Update an existing genetic data record.
// @Tags        GeneticData
// @Accept      json
// @Produce     json
// @Param       id           path     string                   true "Genetic Data ID"
// @Param       geneticData body     health.GeneticData true "Updated genetic data"
// @Security    ApiKeyAuth
// @Success     202     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/genetic-data/{id} [put]
func (h *GeneticDataHandler) UpdateGeneticData(c *gin.Context) {
	geneticDataID := c.Param("id")
	var geneticData health.GeneticData
	if err := c.ShouldBindJSON(&geneticData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body " + err.Error()})
		return
	}

	// Ensure the ID in the URL matches the ID in the payload
	if geneticData.Id != geneticDataID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID mismatch"})
		return
	}

	// Publish to Kafka
	if err := h.kafkaProducer.ProduceMessage(c.Request.Context(), h.kafkaProducer.Cfg.KafkaGeneticDataTopic, "genetic_data.update", &geneticData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update genetic data " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Genetic data update request accepted"})
}

// DeleteGeneticData godoc
// @Summary     Delete Genetic Data
// @Description Delete a genetic data record by its ID.
// @Tags        GeneticData
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Genetic Data ID"
// @Security    ApiKeyAuth
// @Success     204     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/genetic-data/{id} [delete]
func (h *GeneticDataHandler) DeleteGeneticData(c *gin.Context) {
	geneticDataID := c.Param("id")

	// Call gRPC service to delete genetic data
	_, err := h.service.DeleteGeneticData(context.Background(), &health.ByIdRequest{Id: geneticDataID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Genetic data not found " + err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete genetic data " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Genetic data deleted successfully"})
}

// ListGeneticData godoc
// @Summary     List Genetic Data
// @Description Get a list of genetic data records.
// @Tags        GeneticData
// @Accept      json
// @Produce     json
// @Param        user_id     query    string false  "Filter by user ID"
// @Param        data_type   query    string false  "Filter by data type"
// @Param        analysis_date query string false  "Filter by analysis date (YYYY-MM-DD)"
// @Security    ApiKeyAuth
// @Success     200     {object} health.ListGeneticDataResponse
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/genetic-data [get]
func (h *GeneticDataHandler) ListGeneticData(c *gin.Context) {
	// Get query parameters for pagination and filtering
	userID := c.Query("user_id")
	dataType := c.Query("data_type")
	analysisDate := c.Query("analysis_date")

	// Use gRPC to get the genetic data from the service
	grpcResponse, err := h.service.ListGeneticData(context.Background(), &health.ListGeneticDataRequest{
		UserId:       userID,
		DataType:     dataType,
		AnalysisDate: analysisDate,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get genetic data " + err.Error()})
		return
	}

	for _, data := range grpcResponse.GeneticData {
		// Convert Any proto message to JSON
		dataValueJSON, err := protojson.Marshal(data.DataValue)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal data value " + err.Error()})
			return
		}
		data.DataValue = &anypb.Any{
			Value: dataValueJSON,
		}
	}

	c.JSON(http.StatusOK, grpcResponse)
}
