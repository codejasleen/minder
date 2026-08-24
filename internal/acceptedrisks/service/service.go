// SPDX-FileCopyrightText: Copyright 2026 The Minder Authors
// SPDX-License-Identifier: Apache-2.0

// Package service contains the business logic for accepted risks.
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/mindersec/minder/internal/db"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

// AcceptedRisksService provides operations for accepted risks.
type AcceptedRisksService interface {
	Create(
		ctx context.Context,
		projectID uuid.UUID,
		acceptedRisk *minderv1.AcceptedRisk,
	) (*minderv1.AcceptedRisk, error)

	List(
		ctx context.Context,
		projectID uuid.UUID,
	) ([]*minderv1.AcceptedRisk, error)

	Delete(
		ctx context.Context,
		projectID uuid.UUID,
		id uuid.UUID,
	) error
}

type acceptedRisksService struct {
	store db.Store
}

// NewAcceptedRisksService creates a new accepted risks service.
func NewAcceptedRisksService(store db.Store) AcceptedRisksService {
	return &acceptedRisksService{
		store: store,
	}
}

func (s *acceptedRisksService) Create(
	ctx context.Context,
	projectID uuid.UUID,
	acceptedRisk *minderv1.AcceptedRisk,
) (*minderv1.AcceptedRisk, error) {
	entityID, err := uuid.Parse(acceptedRisk.GetEntityId())
	if err != nil {
		return nil, fmt.Errorf("invalid entity ID: %w", err)
	}

	ruleTypeID, err := uuid.Parse(acceptedRisk.GetRuleTypeId())
	if err != nil {
		return nil, fmt.Errorf("invalid rule type ID: %w", err)
	}

	if acceptedRisk.GetExpiresAt() == nil {
		return nil, fmt.Errorf("expires_at is required")
	}

	expiresAt := acceptedRisk.GetExpiresAt().AsTime()

	// Remove an expired record first so that the unique constraint
	// does not prevent creating a new accepted risk.
	if err := s.store.DeleteExpiredAcceptedRisk(ctx, db.DeleteExpiredAcceptedRiskParams{
		ProjectID:  projectID,
		EntityID:   entityID,
		RuleTypeID: ruleTypeID,
	}); err != nil {
		return nil, fmt.Errorf("failed to remove expired accepted risk: %w", err)
	}

	record, err := s.store.CreateAcceptedRisk(ctx, db.CreateAcceptedRiskParams{
		ProjectID:  projectID,
		EntityID:   entityID,
		RuleTypeID: ruleTypeID,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create accepted risk: %w", err)
	}

	return acceptedRiskFromDB(record), nil
}

func (s *acceptedRisksService) List(
	ctx context.Context,
	projectID uuid.UUID,
) ([]*minderv1.AcceptedRisk, error) {
	risks, err := s.store.ListAcceptedRisksByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list accepted risks: %w", err)
	}

	result := make([]*minderv1.AcceptedRisk, 0, len(risks))
	for _, risk := range risks {
		result = append(result, acceptedRiskFromDB(risk))
	}

	return result, nil
}

func (s *acceptedRisksService) Delete(
	ctx context.Context,
	projectID uuid.UUID,
	id uuid.UUID,
) error {
	return s.store.DeleteAcceptedRisk(ctx, db.DeleteAcceptedRiskParams{
		ID:        id,
		ProjectID: projectID,
	})
}

func acceptedRiskFromDB(risk db.AcceptedRisk) *minderv1.AcceptedRisk {
	return &minderv1.AcceptedRisk{
		Id:         risk.ID.String(),
		EntityId:   risk.EntityID.String(),
		RuleTypeId: risk.RuleTypeID.String(),
		ExpiresAt:  timestamppb.New(risk.ExpiresAt),
		CreatedAt:  timestamppb.New(risk.CreatedAt),
	}
}
