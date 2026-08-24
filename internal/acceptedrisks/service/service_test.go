// SPDX-FileCopyrightText: Copyright 2026 The Minder Authors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"

	mockdb "github.com/mindersec/minder/database/mock"
	"github.com/mindersec/minder/internal/db"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	projectID := uuid.New()
	entityID := uuid.New()
	ruleTypeID := uuid.New()
	riskID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour).Round(0)
	createdAt := time.Now().Round(0)

	tests := []struct {
		name    string
		risk    *minderv1.AcceptedRisk
		setup   func(*mockdb.MockStore)
		want    *minderv1.AcceptedRisk
		wantErr bool
	}{
		{
			name: "success",
			risk: &minderv1.AcceptedRisk{
				EntityId:   entityID.String(),
				RuleTypeId: ruleTypeID.String(),
				ExpiresAt:  timestamppb.New(expiresAt),
			},
			setup: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteExpiredAcceptedRisk(
						gomock.Any(),
						db.DeleteExpiredAcceptedRiskParams{
							ProjectID:  projectID,
							EntityID:   entityID,
							RuleTypeID: ruleTypeID,
						},
					).
					Return(nil)

				store.EXPECT().
					CreateAcceptedRisk(
						gomock.Any(),
						gomock.Cond(func(params db.CreateAcceptedRiskParams) bool {
							return params.ProjectID == projectID &&
								params.EntityID == entityID &&
								params.RuleTypeID == ruleTypeID &&
								params.ExpiresAt.Equal(expiresAt)
						}),
					).
					Return(db.AcceptedRisk{
						ID:         riskID,
						ProjectID:  projectID,
						EntityID:   entityID,
						RuleTypeID: ruleTypeID,
						ExpiresAt:  expiresAt,
						CreatedAt:  createdAt,
					}, nil)
			},
			want: &minderv1.AcceptedRisk{
				Id:         riskID.String(),
				EntityId:   entityID.String(),
				RuleTypeId: ruleTypeID.String(),
				ExpiresAt:  timestamppb.New(expiresAt),
				CreatedAt:  timestamppb.New(createdAt),
			},
		},
		{
			name: "invalid entity ID",
			risk: &minderv1.AcceptedRisk{
				EntityId:   "invalid",
				RuleTypeId: ruleTypeID.String(),
				ExpiresAt:  timestamppb.New(expiresAt),
			},
			setup:   func(_ *mockdb.MockStore) {},
			wantErr: true,
		},
		{
			name: "invalid rule type ID",
			risk: &minderv1.AcceptedRisk{
				EntityId:   entityID.String(),
				RuleTypeId: "invalid",
				ExpiresAt:  timestamppb.New(expiresAt),
			},
			setup:   func(_ *mockdb.MockStore) {},
			wantErr: true,
		},
		{
			name: "failed to delete expired risk",
			risk: &minderv1.AcceptedRisk{
				EntityId:   entityID.String(),
				RuleTypeId: ruleTypeID.String(),
				ExpiresAt:  timestamppb.New(expiresAt),
			},
			setup: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteExpiredAcceptedRisk(gomock.Any(), gomock.Any()).
					Return(sql.ErrConnDone)
			},
			wantErr: true,
		},
		{
			name: "failed to create accepted risk",
			risk: &minderv1.AcceptedRisk{
				EntityId:   entityID.String(),
				RuleTypeId: ruleTypeID.String(),
				ExpiresAt:  timestamppb.New(expiresAt),
			},
			setup: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteExpiredAcceptedRisk(gomock.Any(), gomock.Any()).
					Return(nil)

				store.EXPECT().
					CreateAcceptedRisk(gomock.Any(), gomock.Any()).
					Return(db.AcceptedRisk{}, sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)
			tt.setup(store)

			svc := NewAcceptedRisksService(store)

			got, err := svc.Create(context.Background(), projectID, tt.risk)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	projectID := uuid.New()
	entityID := uuid.New()
	ruleTypeID := uuid.New()
	riskID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)
	createdAt := time.Now()

	tests := []struct {
		name    string
		setup   func(*mockdb.MockStore)
		want    []*minderv1.AcceptedRisk
		wantErr bool
	}{
		{
			name: "success",
			setup: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAcceptedRisksByProjectID(gomock.Any(), projectID).
					Return([]db.AcceptedRisk{
						{
							ID:         riskID,
							ProjectID:  projectID,
							EntityID:   entityID,
							RuleTypeID: ruleTypeID,
							ExpiresAt:  expiresAt,
							CreatedAt:  createdAt,
						},
					}, nil)
			},
			want: []*minderv1.AcceptedRisk{
				{
					Id:         riskID.String(),
					EntityId:   entityID.String(),
					RuleTypeId: ruleTypeID.String(),
					ExpiresAt:  timestamppb.New(expiresAt),
					CreatedAt:  timestamppb.New(createdAt),
				},
			},
		},
		{
			name: "empty",
			setup: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAcceptedRisksByProjectID(gomock.Any(), projectID).
					Return([]db.AcceptedRisk{}, nil)
			},
			want: []*minderv1.AcceptedRisk{},
		},
		{
			name: "database error",
			setup: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAcceptedRisksByProjectID(gomock.Any(), projectID).
					Return(nil, sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)
			tt.setup(store)

			svc := NewAcceptedRisksService(store)

			got, err := svc.List(context.Background(), projectID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	projectID := uuid.New()
	riskID := uuid.New()

	tests := []struct {
		name    string
		setup   func(*mockdb.MockStore)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAcceptedRisk(
						gomock.Any(),
						db.DeleteAcceptedRiskParams{
							ID:        riskID,
							ProjectID: projectID,
						},
					).
					Return(nil)
			},
		},
		{
			name: "database error",
			setup: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAcceptedRisk(gomock.Any(), gomock.Any()).
					Return(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)
			tt.setup(store)

			svc := NewAcceptedRisksService(store)

			err := svc.Delete(context.Background(), projectID, riskID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCreateMissingExpiresAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mockdb.NewMockStore(ctrl)

	service := NewAcceptedRisksService(store)

	_, err := service.Create(
		context.Background(),
		uuid.New(),
		&minderv1.AcceptedRisk{
			EntityId:   uuid.New().String(),
			RuleTypeId: uuid.New().String(),
		},
	)

	require.Error(t, err)
	require.EqualError(t, err, "expires_at is required")
}
