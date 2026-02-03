package service

import (
	"context"
	"log/slog"
	"subscription-service/internal/domain"
	"subscription-service/internal/repository"
)

type SubscriptionService struct {
	repo *repository.Repository
	log  *slog.Logger
}

func NewService(repo *repository.Repository, log *slog.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, log: log}
}

func (s *SubscriptionService) Create(ctx context.Context, sub domain.Subscription) (int, error) {
	s.log.Info("creating subscription", "user_id", sub.UserID)
	return s.repo.Create(ctx, sub)
}

func (s *SubscriptionService) GetTotalCost(ctx context.Context, userID string) (int, error) {
	subs, err := s.repo.GetAllByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, sub := range subs {
		total += sub.Price
	}
	return total, nil
}

func (s *SubscriptionService) Update(ctx context.Context, sub domain.Subscription) error {
	s.log.Info("updating subscription", "id", sub.ID)
	return s.repo.Update(ctx, sub)
}

func (s *SubscriptionService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
