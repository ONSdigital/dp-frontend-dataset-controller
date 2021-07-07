// Code generated by MockGen. DO NOT EDIT.
// Source: handlers.go

// Package handlers is a generated GoMock package.
package handlers

import (
	context "context"
	io "io"
	reflect "reflect"

	dataset "github.com/ONSdigital/dp-api-clients-go/dataset"
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
func (m *MockDatasetClient) GetVersions(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition string) ([]dataset.Version, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVersions", ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition)
	ret0, _ := ret[0].([]dataset.Version)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVersions indicates an expected call of GetVersions.
func (mr *MockDatasetClientMockRecorder) GetVersions(ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVersions", reflect.TypeOf((*MockDatasetClient)(nil).GetVersions), ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition)
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

// Page mocks base method.
func (m *MockRenderClient) Page(w io.Writer, page interface{}, templateName string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Page", w, page, templateName)
}

// Page indicates an expected call of Page.
func (mr *MockRenderClientMockRecorder) Page(w, page, templateName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Page", reflect.TypeOf((*MockRenderClient)(nil).Page), w, page, templateName)
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
