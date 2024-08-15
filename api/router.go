package api

import (
	"github.com/gin-gonic/gin"

	"github.com/health-analytics-service/api-gateway-health-analytics/api/auth"
	_ "github.com/health-analytics-service/api-gateway-health-analytics/api/docs"
	"github.com/health-analytics-service/api-gateway-health-analytics/api/handlers"
	"github.com/health-analytics-service/api-gateway-health-analytics/config"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
)

// @title           Swagger Example API
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @BasePath  /v1
// @description					Description for what is this security definition being used
func NewRouter(healthGrpcConn *grpc.ClientConn) *gin.Engine {
	cfg := config.Load()
	router := gin.Default()

	handler := handlers.NewHandler(healthGrpcConn, &cfg)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API versioning
	v1 := router.Group("/v1")
	v1.Use(auth.AuthMiddleware(&cfg))
	{
		// Genetic Data routes
		geneticData := v1.Group("/genetic-data")
		{
			geneticData.POST("", handler.GeneticDataHandler.CreateGeneticData)
			geneticData.GET(":id", handler.GeneticDataHandler.GetGeneticData)
			geneticData.PUT(":id", handler.GeneticDataHandler.UpdateGeneticData)
			geneticData.DELETE(":id", handler.GeneticDataHandler.DeleteGeneticData)
			geneticData.GET("", handler.GeneticDataHandler.ListGeneticData)
		}

		// Health Recommendation routes
		healthRecommendations := v1.Group("/health-recommendations")
		{
			healthRecommendations.POST("", handler.HealthRecommendationHandler.CreateHealthRecommendation)
			healthRecommendations.GET(":id", handler.HealthRecommendationHandler.GetHealthRecommendation)
			healthRecommendations.PUT(":id", handler.HealthRecommendationHandler.UpdateHealthRecommendation)
			healthRecommendations.DELETE(":id", handler.HealthRecommendationHandler.DeleteHealthRecommendation)
			healthRecommendations.GET("", handler.HealthRecommendationHandler.ListHealthRecommendations)
		}

		// Lifestyle Data routes
		lifestyleData := v1.Group("/lifestyle-data")
		{
			lifestyleData.POST("", handler.LifestyleDataHandler.CreateLifestyleData)
			lifestyleData.GET(":id", handler.LifestyleDataHandler.GetLifestyleData)
			lifestyleData.PUT(":id", handler.LifestyleDataHandler.UpdateLifestyleData)
			lifestyleData.DELETE(":id", handler.LifestyleDataHandler.DeleteLifestyleData)
			lifestyleData.GET("", handler.LifestyleDataHandler.ListLifestyleData)
		}

		// Medical Record routes
		medicalRecords := v1.Group("/medical-records")
		{
			medicalRecords.POST("", handler.MedicalRecordHandler.CreateMedicalRecord)
			medicalRecords.GET(":id", handler.MedicalRecordHandler.GetMedicalRecord)
			medicalRecords.PUT(":id", handler.MedicalRecordHandler.UpdateMedicalRecord)
			medicalRecords.DELETE(":id", handler.MedicalRecordHandler.DeleteMedicalRecord)
			medicalRecords.GET("", handler.MedicalRecordHandler.ListMedicalRecords)
		}

		// Wearable Data routes
		wearableData := v1.Group("/wearable-data")
		{
			wearableData.POST("", handler.WearableDataHandler.CreateWearableData)
			wearableData.GET(":id", handler.WearableDataHandler.GetWearableData)
			wearableData.PUT(":id", handler.WearableDataHandler.UpdateWearableData)
			wearableData.DELETE(":id", handler.WearableDataHandler.DeleteWearableData)
			wearableData.GET("", handler.WearableDataHandler.ListWearableData)
		}

		// Health Monitoring routes
		healthMonitoring := v1.Group("/health-monitoring")
		{
			healthMonitoring.GET("daily-summary/:user_id", handler.HealthMonitoringHandler.GetDailySummary)
			healthMonitoring.GET("weekly-summary/:user_id", handler.HealthMonitoringHandler.GetWeeklySummary)
		}
	}

	return router
}

/*1. Authentication and Authorization:

Implementation: You have placeholders for auth.AuthMiddleware and auth.AuthorizationMiddleware in your router. You need to implement these middleware functions based on your chosen authentication and authorization mechanisms (e.g., JWT, OAuth, API keys).
Security Considerations: Review and implement appropriate security measures to protect your API gateway and backend services. This might include rate limiting, input validation, and protection against common web vulnerabilities.
2. Error Handling:

Centralized Error Handling: Consider implementing a centralized error handling mechanism to gracefully handle errors and provide consistent error responses to clients.
Logging: Implement proper logging to track errors, requests, and other relevant information for debugging and monitoring.
3. Testing:

Unit Tests: Write unit tests for your handlers, Kafka producers and consumers, and other components to ensure their correctness.
Integration Tests: Perform integration tests to verify the interactions between your API gateway, Kafka, and backend services.
4. Deployment:

Containerization: Consider containerizing your API gateway using Docker for easier deployment and scalability.
Orchestration: If you're deploying multiple instances of your API gateway, use an orchestration tool like Kubernetes to manage and scale your deployment.
5. Monitoring and Observability:

Metrics: Collect metrics on API gateway performance, request latency, error rates, and other relevant indicators.
Tracing: Implement distributed tracing to track requests across your microservices, including the API gateway, Kafka, and backend services.
6. Documentation:

API Documentation: Ensure that your Swagger documentation is complete and up-to-date.
Internal Documentation: Document your API gateway's architecture, deployment process, and other relevant information for maintainability and knowledge sharing.*/
