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
)

// MedicalRecordHandler handles requests related to Medical Records.
type MedicalRecordHandler struct {
	kafkaProducer *kafka.Producer
	service       health.MedicalRecordServiceClient
}

// NewMedicalRecordHandler creates a new MedicalRecordHandler.
func NewMedicalRecordHandler(kafkaProducer *kafka.Producer, healthGrpcConn *grpc.ClientConn) *MedicalRecordHandler {
	return &MedicalRecordHandler{
		kafkaProducer: kafkaProducer,
		service:       health.NewMedicalRecordServiceClient(healthGrpcConn),
	}
}

// CreateMedicalRecord godoc
// @Summary     Create a new Medical Record
// @Description Create a new medical record with the provided details.
// @Tags        MedicalRecords
// @Accept      json
// @Produce     json
// @Param       medicalRecord body     health.MedicalRecord true "Medical Record details"
// @Security    ApiKeyAuth
// @Success     202     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/medical-records [post]
func (h *MedicalRecordHandler) CreateMedicalRecord(c *gin.Context) {
	var medicalRecord health.MedicalRecord
	if err := c.ShouldBindJSON(&medicalRecord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body " + err.Error()})
		return
	}

	// Publish to Kafka
	if err := h.kafkaProducer.ProduceMessage(c.Request.Context(), h.kafkaProducer.Cfg.KafkaMedicalRecordTopic, "medical_record.create", &medicalRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create medical record " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Medical record creation request accepted"})
}

// GetMedicalRecord godoc
// @Summary     Get Medical Record by ID
// @Description Get details of a medical record by its ID.
// @Tags        MedicalRecords
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Medical Record ID"
// @Security    ApiKeyAuth
// @Success     200     {object} health.MedicalRecord
// @Failure     400     {object} map[string]interface{}
// @Failure     404     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/medical-records/{id} [get]
func (h *MedicalRecordHandler) GetMedicalRecord(c *gin.Context) {
	medicalRecordID := c.Param("id")

	// Use gRPC to get the medical record from the service
	grpcResponse, err := h.service.GetMedicalRecord(context.Background(), &health.ByIdRequest{Id: medicalRecordID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Medical record not found" + err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get medical record " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, grpcResponse)
}

// UpdateMedicalRecord godoc
// @Summary     Update Medical Record
// @Description Update an existing medical record.
// @Tags        MedicalRecords
// @Accept      json
// @Produce     json
// @Param       id           path     string                   true "Medical Record ID"
// @Param       medicalRecord body     health.MedicalRecord true "Updated medical record"
// @Security    ApiKeyAuth
// @Success     202     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/medical-records/{id} [put]
func (h *MedicalRecordHandler) UpdateMedicalRecord(c *gin.Context) {
	medicalRecordID := c.Param("id")
	var medicalRecord health.MedicalRecord
	if err := c.ShouldBindJSON(&medicalRecord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body " + err.Error()})
		return
	}

	// Ensure the ID in the URL matches the ID in the payload
	if medicalRecord.Id != medicalRecordID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID mismatch"})
		return
	}

	// Publish to Kafka
	if err := h.kafkaProducer.ProduceMessage(c.Request.Context(), h.kafkaProducer.Cfg.KafkaMedicalRecordTopic, "medical_record.update", &medicalRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update medical record " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Medical record update request accepted"})
}

// DeleteMedicalRecord godoc
// @Summary     Delete Medical Record
// @Description Delete a medical record by its ID.
// @Tags        MedicalRecords
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Medical Record ID"
// @Security    ApiKeyAuth
// @Success     204     {object} map[string]interface{}
// @Failure     400     {object} map[string]interface{}
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/medical-records/{id} [delete]
func (h *MedicalRecordHandler) DeleteMedicalRecord(c *gin.Context) {
	medicalRecordID := c.Param("id")

	// Call gRPC service to delete medical record
	_, err := h.service.DeleteMedicalRecord(context.Background(), &health.ByIdRequest{Id: medicalRecordID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Medical record not found " + err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete medical record " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Medical record deleted successfully"})
}

// ListMedicalRecords godoc
// @Summary     List Medical Records
// @Description Get a list of medical records.
// @Tags        MedicalRecords
// @Accept      json
// @Produce     json
// @Param        user_id     query    string false  "Filter by user ID"
// @Param        record_type query    string false  "Filter by record type"
// @Param        record_date query    string false  "Filter by record date (YYYY-MM-DD)"
// @Param        description query    string false  "Filter by description"
// @Param        doctor_id   query    string false  "Filter by doctor ID"
// @Security    ApiKeyAuth
// @Success     200     {object} health.ListMedicalRecordsResponse
// @Failure     500     {object} map[string]interface{}
// @Router      /v1/medical-records [get]
func (h *MedicalRecordHandler) ListMedicalRecords(c *gin.Context) {
	// Get query parameters for pagination and filtering
	userID := c.Query("user_id")
	recordType := c.Query("record_type")
	recordDate := c.Query("record_date")
	description := c.Query("description")
	doctorID := c.Query("doctor_id")

	// Use gRPC to get the medical records from the service
	grpcResponse, err := h.service.ListMedicalRecords(context.Background(), &health.ListMedicalRecordsRequest{
		UserId:      userID,
		RecordType:  recordType,
		RecordDate:  recordDate,
		Description: description,
		DoctorId:    doctorID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get medical records " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, grpcResponse)
}
