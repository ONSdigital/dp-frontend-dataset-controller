// Code generated by MockGen. DO NOT EDIT.
// Source: clients.go

// Package handlers is a generated GoMock package.
package handlers

import (
	context "context"
	io "io"
	reflect "reflect"

	cantabular "github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	dataset "github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	files "github.com/ONSdigital/dp-api-clients-go/v2/files"
	filter "github.com/ONSdigital/dp-api-clients-go/v2/filter"
	population "github.com/ONSdigital/dp-api-clients-go/v2/population"
	model "github.com/ONSdigital/dp-renderer/v2/model"
	gomock "github.com/golang/mock/gomock"
)

// MockFilterClient is a mock of FilterClient interface.
type MockFilterClient struct {
	ctrl     *gomock.Controller
	recorder *MockFilterClientMockRecorder
}

// MockFilterClientMockRecorder is the mock recorder for MockFilterClient.
type MockFilterClientMockRecorder struct {
	mock *MockFilterClient
}

// NewMockFilterClient creates a new mock instance.
func NewMockFilterClient(ctrl *gomock.Controller) *MockFilterClient {
	mock := &MockFilterClient{ctrl: ctrl}
	mock.recorder = &MockFilterClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFilterClient) EXPECT() *MockFilterClientMockRecorder {
	return m.recorder
}

// CreateBlueprint mocks base method.
func (m *MockFilterClient) CreateBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, names []string) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateBlueprint", ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version, names)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateBlueprint indicates an expected call of CreateBlueprint.
func (mr *MockFilterClientMockRecorder) CreateBlueprint(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version, names interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBlueprint", reflect.TypeOf((*MockFilterClient)(nil).CreateBlueprint), ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version, names)
}

// CreateCustomFilter mocks base method.
func (m *MockFilterClient) CreateCustomFilter(ctx context.Context, userAuthToken, serviceAuthToken, populationType string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCustomFilter", ctx, userAuthToken, serviceAuthToken, populationType)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateCustomFilter indicates an expected call of CreateCustomFilter.
func (mr *MockFilterClientMockRecorder) CreateCustomFilter(ctx, userAuthToken, serviceAuthToken, populationType interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCustomFilter", reflect.TypeOf((*MockFilterClient)(nil).CreateCustomFilter), ctx, userAuthToken, serviceAuthToken, populationType)
}

// CreateFlexibleBlueprint mocks base method.
func (m *MockFilterClient) CreateFlexibleBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, dimensions []filter.ModelDimension, population_type string) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateFlexibleBlueprint", ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version, dimensions, population_type)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateFlexibleBlueprint indicates an expected call of CreateFlexibleBlueprint.
func (mr *MockFilterClientMockRecorder) CreateFlexibleBlueprint(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version, dimensions, population_type interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFlexibleBlueprint", reflect.TypeOf((*MockFilterClient)(nil).CreateFlexibleBlueprint), ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version, dimensions, population_type)
}

// CreateFlexibleBlueprintCustom mocks base method.
func (m *MockFilterClient) CreateFlexibleBlueprintCustom(ctx context.Context, uAuthToken, svcAuthToken, dlServiceToken string, req filter.CreateFlexBlueprintCustomRequest) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateFlexibleBlueprintCustom", ctx, uAuthToken, svcAuthToken, dlServiceToken, req)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateFlexibleBlueprintCustom indicates an expected call of CreateFlexibleBlueprintCustom.
func (mr *MockFilterClientMockRecorder) CreateFlexibleBlueprintCustom(ctx, uAuthToken, svcAuthToken, dlServiceToken, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFlexibleBlueprintCustom", reflect.TypeOf((*MockFilterClient)(nil).CreateFlexibleBlueprintCustom), ctx, uAuthToken, svcAuthToken, dlServiceToken, req)
}

// GetDimensionOptions mocks base method.
func (m *MockFilterClient) GetDimensionOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, q *filter.QueryParams) (filter.DimensionOptions, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDimensionOptions", ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, q)
	ret0, _ := ret[0].(filter.DimensionOptions)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetDimensionOptions indicates an expected call of GetDimensionOptions.
func (mr *MockFilterClientMockRecorder) GetDimensionOptions(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, q interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDimensionOptions", reflect.TypeOf((*MockFilterClient)(nil).GetDimensionOptions), ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, q)
}

// GetOutput mocks base method.
func (m *MockFilterClient) GetOutput(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID string) (filter.Model, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOutput", ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID)
	ret0, _ := ret[0].(filter.Model)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOutput indicates an expected call of GetOutput.
func (mr *MockFilterClientMockRecorder) GetOutput(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOutput", reflect.TypeOf((*MockFilterClient)(nil).GetOutput), ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID)
}

// MockDatasetClient is a mock of DatasetClient interface.
type MockDatasetClient struct {
	ctrl     *gomock.Controller
	recorder *MockDatasetClientMockRecorder
}

// MockDatasetClientMockRecorder is the mock recorder for MockDatasetClient.
type MockDatasetClientMockRecorder struct {
	mock *MockDatasetClient
}

// NewMockDatasetClient creates a new mock instance.
func NewMockDatasetClient(ctrl *gomock.Controller) *MockDatasetClient {
	mock := &MockDatasetClient{ctrl: ctrl}
	mock.recorder = &MockDatasetClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDatasetClient) EXPECT() *MockDatasetClientMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockDatasetClient) Get(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (dataset.DatasetDetails, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, userAuthToken, serviceAuthToken, collectionID, datasetID)
	ret0, _ := ret[0].(dataset.DatasetDetails)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockDatasetClientMockRecorder) Get(ctx, userAuthToken, serviceAuthToken, collectionID, datasetID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockDatasetClient)(nil).Get), ctx, userAuthToken, serviceAuthToken, collectionID, datasetID)
}

// GetByPath mocks base method.
func (m *MockDatasetClient) GetByPath(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, path string) (dataset.DatasetDetails, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByPath", ctx, userAuthToken, serviceAuthToken, collectionID, path)
	ret0, _ := ret[0].(dataset.DatasetDetails)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByPath indicates an expected call of GetByPath.
func (mr *MockDatasetClientMockRecorder) GetByPath(ctx, userAuthToken, serviceAuthToken, collectionID, path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByPath", reflect.TypeOf((*MockDatasetClient)(nil).GetByPath), ctx, userAuthToken, serviceAuthToken, collectionID, path)
}

// GetEdition mocks base method.
func (m *MockDatasetClient) GetEdition(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition string) (dataset.Edition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEdition", ctx, userAuthToken, serviceAuthToken, collectionID, datasetID, edition)
	ret0, _ := ret[0].(dataset.Edition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEdition indicates an expected call of GetEdition.
func (mr *MockDatasetClientMockRecorder) GetEdition(ctx, userAuthToken, serviceAuthToken, collectionID, datasetID, edition interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEdition", reflect.TypeOf((*MockDatasetClient)(nil).GetEdition), ctx, userAuthToken, serviceAuthToken, collectionID, datasetID, edition)
}

// GetEditions mocks base method.
func (m *MockDatasetClient) GetEditions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) ([]dataset.Edition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEditions", ctx, userAuthToken, serviceAuthToken, collectionID, datasetID)
	ret0, _ := ret[0].([]dataset.Edition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEditions indicates an expected call of GetEditions.
func (mr *MockDatasetClientMockRecorder) GetEditions(ctx, userAuthToken, serviceAuthToken, collectionID, datasetID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEditions", reflect.TypeOf((*MockDatasetClient)(nil).GetEditions), ctx, userAuthToken, serviceAuthToken, collectionID, datasetID)
}

// GetOptions mocks base method.
func (m *MockDatasetClient) GetOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, q *dataset.QueryParams) (dataset.Options, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOptions", ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension, q)
	ret0, _ := ret[0].(dataset.Options)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOptions indicates an expected call of GetOptions.
func (mr *MockDatasetClientMockRecorder) GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension, q interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOptions", reflect.TypeOf((*MockDatasetClient)(nil).GetOptions), ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension, q)
}

// GetVersion mocks base method.
func (m *MockDatasetClient) GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (dataset.Version, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVersion", ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version)
	ret0, _ := ret[0].(dataset.Version)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVersion indicates an expected call of GetVersion.
func (mr *MockDatasetClientMockRecorder) GetVersion(ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVersion", reflect.TypeOf((*MockDatasetClient)(nil).GetVersion), ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version)
}

// GetVersionDimensions mocks base method.
func (m *MockDatasetClient) GetVersionDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (dataset.VersionDimensions, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVersionDimensions", ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version)
	ret0, _ := ret[0].(dataset.VersionDimensions)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVersionDimensions indicates an expected call of GetVersionDimensions.
func (mr *MockDatasetClientMockRecorder) GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVersionDimensions", reflect.TypeOf((*MockDatasetClient)(nil).GetVersionDimensions), ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version)
}

// GetVersionMetadata mocks base method.
func (m *MockDatasetClient) GetVersionMetadata(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (dataset.Metadata, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVersionMetadata", ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version)
	ret0, _ := ret[0].(dataset.Metadata)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVersionMetadata indicates an expected call of GetVersionMetadata.
func (mr *MockDatasetClientMockRecorder) GetVersionMetadata(ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVersionMetadata", reflect.TypeOf((*MockDatasetClient)(nil).GetVersionMetadata), ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version)
}

// GetVersions mocks base method.
func (m *MockDatasetClient) GetVersions(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition string, q *dataset.QueryParams) (dataset.VersionsList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVersions", ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, q)
	ret0, _ := ret[0].(dataset.VersionsList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVersions indicates an expected call of GetVersions.
func (mr *MockDatasetClientMockRecorder) GetVersions(ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, q interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVersions", reflect.TypeOf((*MockDatasetClient)(nil).GetVersions), ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, q)
}

// MockRenderClient is a mock of RenderClient interface.
type MockRenderClient struct {
	ctrl     *gomock.Controller
	recorder *MockRenderClientMockRecorder
}

// MockRenderClientMockRecorder is the mock recorder for MockRenderClient.
type MockRenderClientMockRecorder struct {
	mock *MockRenderClient
}

// NewMockRenderClient creates a new mock instance.
func NewMockRenderClient(ctrl *gomock.Controller) *MockRenderClient {
	mock := &MockRenderClient{ctrl: ctrl}
	mock.recorder = &MockRenderClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRenderClient) EXPECT() *MockRenderClientMockRecorder {
	return m.recorder
}

// BuildPage mocks base method.
func (m *MockRenderClient) BuildPage(w io.Writer, pageModel interface{}, templateName string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "BuildPage", w, pageModel, templateName)
}

// BuildPage indicates an expected call of BuildPage.
func (mr *MockRenderClientMockRecorder) BuildPage(w, pageModel, templateName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildPage", reflect.TypeOf((*MockRenderClient)(nil).BuildPage), w, pageModel, templateName)
}

// NewBasePageModel mocks base method.
func (m *MockRenderClient) NewBasePageModel() model.Page {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewBasePageModel")
	ret0, _ := ret[0].(model.Page)
	return ret0
}

// NewBasePageModel indicates an expected call of NewBasePageModel.
func (mr *MockRenderClientMockRecorder) NewBasePageModel() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewBasePageModel", reflect.TypeOf((*MockRenderClient)(nil).NewBasePageModel))
}

// MockFilesAPIClient is a mock of FilesAPIClient interface.
type MockFilesAPIClient struct {
	ctrl     *gomock.Controller
	recorder *MockFilesAPIClientMockRecorder
}

// MockFilesAPIClientMockRecorder is the mock recorder for MockFilesAPIClient.
type MockFilesAPIClientMockRecorder struct {
	mock *MockFilesAPIClient
}

// NewMockFilesAPIClient creates a new mock instance.
func NewMockFilesAPIClient(ctrl *gomock.Controller) *MockFilesAPIClient {
	mock := &MockFilesAPIClient{ctrl: ctrl}
	mock.recorder = &MockFilesAPIClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFilesAPIClient) EXPECT() *MockFilesAPIClientMockRecorder {
	return m.recorder
}

// GetFile mocks base method.
func (m *MockFilesAPIClient) GetFile(ctx context.Context, path, authToken string) (files.FileMetaData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFile", ctx, path, authToken)
	ret0, _ := ret[0].(files.FileMetaData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFile indicates an expected call of GetFile.
func (mr *MockFilesAPIClientMockRecorder) GetFile(ctx, path, authToken interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFile", reflect.TypeOf((*MockFilesAPIClient)(nil).GetFile), ctx, path, authToken)
}

// MockPopulationClient is a mock of PopulationClient interface.
type MockPopulationClient struct {
	ctrl     *gomock.Controller
	recorder *MockPopulationClientMockRecorder
}

// MockPopulationClientMockRecorder is the mock recorder for MockPopulationClient.
type MockPopulationClientMockRecorder struct {
	mock *MockPopulationClient
}

// NewMockPopulationClient creates a new mock instance.
func NewMockPopulationClient(ctrl *gomock.Controller) *MockPopulationClient {
	mock := &MockPopulationClient{ctrl: ctrl}
	mock.recorder = &MockPopulationClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPopulationClient) EXPECT() *MockPopulationClientMockRecorder {
	return m.recorder
}

// GetArea mocks base method.
func (m *MockPopulationClient) GetArea(ctx context.Context, input population.GetAreaInput) (population.GetAreaResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetArea", ctx, input)
	ret0, _ := ret[0].(population.GetAreaResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetArea indicates an expected call of GetArea.
func (mr *MockPopulationClientMockRecorder) GetArea(ctx, input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetArea", reflect.TypeOf((*MockPopulationClient)(nil).GetArea), ctx, input)
}

// GetAreas mocks base method.
func (m *MockPopulationClient) GetAreas(ctx context.Context, input population.GetAreasInput) (population.GetAreasResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAreas", ctx, input)
	ret0, _ := ret[0].(population.GetAreasResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAreas indicates an expected call of GetAreas.
func (mr *MockPopulationClientMockRecorder) GetAreas(ctx, input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAreas", reflect.TypeOf((*MockPopulationClient)(nil).GetAreas), ctx, input)
}

// GetBlockedAreaCount mocks base method.
func (m *MockPopulationClient) GetBlockedAreaCount(ctx context.Context, input population.GetBlockedAreaCountInput) (*cantabular.GetBlockedAreaCountResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockedAreaCount", ctx, input)
	ret0, _ := ret[0].(*cantabular.GetBlockedAreaCountResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockedAreaCount indicates an expected call of GetBlockedAreaCount.
func (mr *MockPopulationClientMockRecorder) GetBlockedAreaCount(ctx, input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockedAreaCount", reflect.TypeOf((*MockPopulationClient)(nil).GetBlockedAreaCount), ctx, input)
}

// GetCategorisations mocks base method.
func (m *MockPopulationClient) GetCategorisations(ctx context.Context, input population.GetCategorisationsInput) (population.GetCategorisationsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCategorisations", ctx, input)
	ret0, _ := ret[0].(population.GetCategorisationsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCategorisations indicates an expected call of GetCategorisations.
func (mr *MockPopulationClientMockRecorder) GetCategorisations(ctx, input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCategorisations", reflect.TypeOf((*MockPopulationClient)(nil).GetCategorisations), ctx, input)
}

// GetDimensionCategories mocks base method.
func (m *MockPopulationClient) GetDimensionCategories(ctx context.Context, input population.GetDimensionCategoryInput) (population.GetDimensionCategoriesResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDimensionCategories", ctx, input)
	ret0, _ := ret[0].(population.GetDimensionCategoriesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDimensionCategories indicates an expected call of GetDimensionCategories.
func (mr *MockPopulationClientMockRecorder) GetDimensionCategories(ctx, input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDimensionCategories", reflect.TypeOf((*MockPopulationClient)(nil).GetDimensionCategories), ctx, input)
}

// GetDimensionsDescription mocks base method.
func (m *MockPopulationClient) GetDimensionsDescription(ctx context.Context, input population.GetDimensionsDescriptionInput) (population.GetDimensionsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDimensionsDescription", ctx, input)
	ret0, _ := ret[0].(population.GetDimensionsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDimensionsDescription indicates an expected call of GetDimensionsDescription.
func (mr *MockPopulationClientMockRecorder) GetDimensionsDescription(ctx, input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDimensionsDescription", reflect.TypeOf((*MockPopulationClient)(nil).GetDimensionsDescription), ctx, input)
}

// GetPopulationType mocks base method.
func (m *MockPopulationClient) GetPopulationType(ctx context.Context, input population.GetPopulationTypeInput) (population.GetPopulationTypeResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPopulationType", ctx, input)
	ret0, _ := ret[0].(population.GetPopulationTypeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPopulationType indicates an expected call of GetPopulationType.
func (mr *MockPopulationClientMockRecorder) GetPopulationType(ctx, input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPopulationType", reflect.TypeOf((*MockPopulationClient)(nil).GetPopulationType), ctx, input)
}

// GetPopulationTypeMetadata mocks base method.
func (m *MockPopulationClient) GetPopulationTypeMetadata(ctx context.Context, input population.GetPopulationTypeMetadataInput) (population.GetPopulationTypeMetadataResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPopulationTypeMetadata", ctx, input)
	ret0, _ := ret[0].(population.GetPopulationTypeMetadataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPopulationTypeMetadata indicates an expected call of GetPopulationTypeMetadata.
func (mr *MockPopulationClientMockRecorder) GetPopulationTypeMetadata(ctx, input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPopulationTypeMetadata", reflect.TypeOf((*MockPopulationClient)(nil).GetPopulationTypeMetadata), ctx, input)
}

// GetPopulationTypes mocks base method.
func (m *MockPopulationClient) GetPopulationTypes(ctx context.Context, input population.GetPopulationTypesInput) (population.GetPopulationTypesResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPopulationTypes", ctx, input)
	ret0, _ := ret[0].(population.GetPopulationTypesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPopulationTypes indicates an expected call of GetPopulationTypes.
func (mr *MockPopulationClientMockRecorder) GetPopulationTypes(ctx, input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPopulationTypes", reflect.TypeOf((*MockPopulationClient)(nil).GetPopulationTypes), ctx, input)
}

// MockClientError is a mock of ClientError interface.
type MockClientError struct {
	ctrl     *gomock.Controller
	recorder *MockClientErrorMockRecorder
}

// MockClientErrorMockRecorder is the mock recorder for MockClientError.
type MockClientErrorMockRecorder struct {
	mock *MockClientError
}

// NewMockClientError creates a new mock instance.
func NewMockClientError(ctrl *gomock.Controller) *MockClientError {
	mock := &MockClientError{ctrl: ctrl}
	mock.recorder = &MockClientErrorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClientError) EXPECT() *MockClientErrorMockRecorder {
	return m.recorder
}

// Code mocks base method.
func (m *MockClientError) Code() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Code")
	ret0, _ := ret[0].(int)
	return ret0
}

// Code indicates an expected call of Code.
func (mr *MockClientErrorMockRecorder) Code() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Code", reflect.TypeOf((*MockClientError)(nil).Code))
}

// Error mocks base method.
func (m *MockClientError) Error() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Error")
	ret0, _ := ret[0].(string)
	return ret0
}

// Error indicates an expected call of Error.
func (mr *MockClientErrorMockRecorder) Error() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockClientError)(nil).Error))
}
