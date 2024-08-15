package handlers

import (
	"google.golang.org/grpc"

	"github.com/health-analytics-service/api-gateway-health-analytics/config"
	"github.com/health-analytics-service/api-gateway-health-analytics/kafka"
)

// Handler struct holds all the individual entity handlers.
type Handler struct {
	// Health service handlers.
	GeneticDataHandler          *GeneticDataHandler
	HealthRecommendationHandler *HealthRecommendationHandler
	LifestyleDataHandler        *LifestyleDataHandler
	MedicalRecordHandler        *MedicalRecordHandler
	WearableDataHandler         *WearableDataHandler
	HealthMonitoringHandler     *HealthMonitoringHandler
}

// It accepts gRPC connections and initializes Kafka producers within each handler.
func NewHandler(healthGrpcConn *grpc.ClientConn, cfg *config.Config) *Handler {
	// Create Kafka producer
	kafkaProducer := kafka.NewProducer(*cfg)

	return &Handler{
		// Health service handlers.
		GeneticDataHandler:          NewGeneticDataHandler(kafkaProducer, healthGrpcConn),
		HealthRecommendationHandler: NewHealthRecommendationHandler(kafkaProducer, healthGrpcConn),
		LifestyleDataHandler:        NewLifestyleDataHandler(kafkaProducer, healthGrpcConn),
		MedicalRecordHandler:        NewMedicalRecordHandler(kafkaProducer, healthGrpcConn),
		WearableDataHandler:         NewWearableDataHandler(kafkaProducer, healthGrpcConn),
		HealthMonitoringHandler:     NewHealthMonitoringHandler(healthGrpcConn),
	}
}
