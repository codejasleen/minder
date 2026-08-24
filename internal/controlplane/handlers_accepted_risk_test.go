package controlplane

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	mockdb "github.com/mindersec/minder/database/mock"
	acceptedriskmock "github.com/mindersec/minder/internal/acceptedrisks/service/mock"
	"github.com/mindersec/minder/internal/db"
	"github.com/mindersec/minder/internal/engine/engcontext"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func acceptedRiskContext(t *testing.T, projectID uuid.UUID) context.Context {
	t.Helper()
	ctx := context.Background()
	return engcontext.WithEntityContext(ctx, &engcontext.EntityContext{
		Project:  engcontext.Project{ID: projectID},
		Provider: engcontext.Provider{Name: "testing"},
	})
}

func setupAcceptedRiskServer(
	t *testing.T,
	projectID uuid.UUID,
	mockStore *mockdb.MockStore,
	mockService *acceptedriskmock.MockAcceptedRisksService,
) *Server {
	t.Helper()

	mockStore.EXPECT().
		GetProjectByID(gomock.Any(), projectID).
		Return(db.Project{ID: projectID}, nil).
		AnyTimes()

	srv := newDefaultServer(t, mockStore, nil, nil, nil)
	srv.acceptedRisksService = mockService
	return srv
}

func TestCreateAcceptedRisk(t *testing.T) {
	t.Parallel()

	projectID := uuid.New()
	risk := &minderv1.AcceptedRisk{
		EntityId:   uuid.New().String(),
		RuleTypeId: uuid.New().String(),
		ExpiresAt:  timestamppb.Now(),
	}

	tests := []struct {
		name              string
		request           *minderv1.CreateAcceptedRiskRequest
		setupMocks        func(*acceptedriskmock.MockAcceptedRisksService)
		expectedErrorCode codes.Code
		expectedResponse  *minderv1.CreateAcceptedRiskResponse
	}{
		{
			name:    "success",
			request: &minderv1.CreateAcceptedRiskRequest{AcceptedRisk: risk},
			setupMocks: func(svc *acceptedriskmock.MockAcceptedRisksService) {
				svc.EXPECT().
					Create(gomock.Any(), projectID, risk).
					Return(risk, nil)
			},
			expectedErrorCode: codes.OK,
			expectedResponse: &minderv1.CreateAcceptedRiskResponse{
				AcceptedRisk: risk,
			},
		},
		{
			name:              "missing accepted risk",
			request:           &minderv1.CreateAcceptedRiskRequest{},
			expectedErrorCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mockdb.NewMockStore(ctrl)
			mockService := acceptedriskmock.NewMockAcceptedRisksService(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			srv := setupAcceptedRiskServer(t, projectID, mockStore, mockService)
			resp, err := srv.CreateAcceptedRisk(
				acceptedRiskContext(t, projectID),
				tt.request,
			)

			if tt.expectedErrorCode != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.expectedErrorCode, st.Code())
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectedResponse, resp)
		})
	}
}

func TestListAcceptedRisks(t *testing.T) {
	t.Parallel()

	projectID := uuid.New()
	risk := &minderv1.AcceptedRisk{
		Id:         uuid.New().String(),
		EntityId:   uuid.New().String(),
		RuleTypeId: uuid.New().String(),
		ExpiresAt:  timestamppb.Now(),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockdb.NewMockStore(ctrl)
	mockService := acceptedriskmock.NewMockAcceptedRisksService(ctrl)

	mockService.EXPECT().
		List(gomock.Any(), projectID).
		Return([]*minderv1.AcceptedRisk{risk}, nil)

	srv := setupAcceptedRiskServer(t, projectID, mockStore, mockService)

	resp, err := srv.ListAcceptedRisks(
		acceptedRiskContext(t, projectID),
		&minderv1.ListAcceptedRisksRequest{},
	)

	require.NoError(t, err)
	require.Equal(t, []*minderv1.AcceptedRisk{risk}, resp.AcceptedRisks)
}

func TestDeleteAcceptedRisk(t *testing.T) {
	t.Parallel()

	projectID := uuid.New()
	riskID := uuid.New()

	tests := []struct {
		name              string
		request           *minderv1.DeleteAcceptedRiskRequest
		setupMocks        func(*acceptedriskmock.MockAcceptedRisksService)
		expectedErrorCode codes.Code
	}{
		{
			name:    "success",
			request: &minderv1.DeleteAcceptedRiskRequest{Id: riskID.String()},
			setupMocks: func(svc *acceptedriskmock.MockAcceptedRisksService) {
				svc.EXPECT().
					Delete(gomock.Any(), projectID, riskID).
					Return(nil)
			},
			expectedErrorCode: codes.OK,
		},
		{
			name:              "missing ID",
			request:           &minderv1.DeleteAcceptedRiskRequest{},
			expectedErrorCode: codes.InvalidArgument,
		},
		{
			name:              "invalid ID",
			request:           &minderv1.DeleteAcceptedRiskRequest{Id: "not-a-uuid"},
			expectedErrorCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mockdb.NewMockStore(ctrl)
			mockService := acceptedriskmock.NewMockAcceptedRisksService(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			srv := setupAcceptedRiskServer(t, projectID, mockStore, mockService)
			resp, err := srv.DeleteAcceptedRisk(
				acceptedRiskContext(t, projectID),
				tt.request,
			)

			if tt.expectedErrorCode != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.expectedErrorCode, st.Code())
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}
