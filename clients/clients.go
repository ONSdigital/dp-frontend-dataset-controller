package clients

// This package defines interfaces for all API clients used within this service.
// Each interface is defined in its own file (e.g. topic.go for TopicAPIClient) to allow for easier maintenance.
//
// Mock implementations are automatically generated in mock_clients.go.
//
//go:generate mockgen -destination=mock_clients.go -package=clients github.com/ONSdigital/dp-frontend-dataset-controller/clients FilterClient,APIClientsGoDatasetClient,DatasetAPISdkClient,TopicAPIClient,PopulationClient,ZebedeeClient,RenderClient,FilesAPIClient,ClientError
