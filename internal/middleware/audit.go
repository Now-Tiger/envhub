package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Now-Tiger/envhub/internal/repository"
)

type AuditLogger struct {
	repo repository.Querier
}

func NewAuditLogger(repo repository.Querier) *AuditLogger {
	return &AuditLogger{repo: repo}
}

func (a *AuditLogger) LogSecretAccess(ctx context.Context, userID uuid.UUID, envID uuid.UUID, action string) {
	_, _ = a.repo.CreateAccessLog(ctx, repository.CreateAccessLogParams{
		UserID:       pgtype.UUID{Bytes: userID, Valid: true},
		ResourceType: "secret",
		ResourceID:   envID,
		Action:       repository.AccessAction(action),
		Success:      true,
	})
}

func (a *AuditLogger) LogFailedDecryption(ctx context.Context, userID, envID uuid.UUID, secretKey, errorMsg string) {
	errMsg := errorMsg
	_, _ = a.repo.CreateAccessLog(ctx, repository.CreateAccessLogParams{
		UserID:       pgtype.UUID{Bytes: userID, Valid: true},
		ResourceType: "secret",
		ResourceID:   envID,
		Action:       "decrypt_failed",
		Success:      false,
		ErrorMessage: &errMsg,
	})
}
