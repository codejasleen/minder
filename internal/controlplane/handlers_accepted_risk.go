// SPDX-FileCopyrightText: Copyright 2026 The Minder Authors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/mindersec/minder/internal/engine/engcontext"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

// CreateAcceptedRisk creates an accepted risk.
func (s *Server) CreateAcceptedRisk(
	ctx context.Context,
	in *minderv1.CreateAcceptedRiskRequest,
) (*minderv1.CreateAcceptedRiskResponse, error) {
	entityCtx := engcontext.EntityFromContext(ctx)

	if err := entityCtx.ValidateProject(ctx, s.store); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error in entity context: %v", err)
	}

	acceptedRisk := in.GetAcceptedRisk()
	if acceptedRisk == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing accepted risk")
	}

	ret, err := s.acceptedRisksService.Create(
		ctx,
		entityCtx.Project.ID,
		acceptedRisk,
	)
	if err != nil {
		return nil, err
	}

	return &minderv1.CreateAcceptedRiskResponse{
		AcceptedRisk: ret,
	}, nil
}

// ListAcceptedRisks lists the active accepted risks for the project.
func (s *Server) ListAcceptedRisks(
	ctx context.Context,
	_ *minderv1.ListAcceptedRisksRequest,
) (*minderv1.ListAcceptedRisksResponse, error) {
	entityCtx := engcontext.EntityFromContext(ctx)

	if err := entityCtx.ValidateProject(ctx, s.store); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error in entity context: %v", err)
	}

	ret, err := s.acceptedRisksService.List(ctx, entityCtx.Project.ID)
	if err != nil {
		return nil, err
	}

	return &minderv1.ListAcceptedRisksResponse{
		AcceptedRisks: ret,
	}, nil
}

// DeleteAcceptedRisk deletes an accepted risk.
func (s *Server) DeleteAcceptedRisk(
	ctx context.Context,
	in *minderv1.DeleteAcceptedRiskRequest,
) (*minderv1.DeleteAcceptedRiskResponse, error) {
	entityCtx := engcontext.EntityFromContext(ctx)

	if err := entityCtx.ValidateProject(ctx, s.store); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error in entity context: %v", err)
	}

	idStr := in.GetId()
	if idStr == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing accepted risk ID")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid accepted risk ID: %v", err)
	}

	if err := s.acceptedRisksService.Delete(ctx, entityCtx.Project.ID, id); err != nil {
		return nil, err
	}

	return &minderv1.DeleteAcceptedRiskResponse{}, nil
}
