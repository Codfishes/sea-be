package service

import (
	"context"
	"strings"
	"time"

	"sea-catering-backend/internal/api/testimonials"
	"sea-catering-backend/internal/api/testimonials/repository"
	"sea-catering-backend/internal/entity"
	"sea-catering-backend/pkg/logger"
	"sea-catering-backend/pkg/utils"
)

type TestimonialService interface {
	CreateTestimonial(ctx context.Context, req testimonials.CreateTestimonialRequest) (*testimonials.TestimonialResponse, error)
	GetApprovedTestimonials(ctx context.Context) ([]testimonials.TestimonialResponse, error)
	GetAllTestimonials(ctx context.Context) ([]testimonials.TestimonialResponse, error)
	ApproveTestimonial(ctx context.Context, id string) error
	RejectTestimonial(ctx context.Context, id string) error
	DeleteTestimonial(ctx context.Context, id string) error
}

type testimonialService struct {
	repo         repository.TestimonialRepository
	utilsService utils.Interface
	logger       *logger.Logger
}

func NewTestimonialService(
	repo repository.TestimonialRepository,
	utilsService utils.Interface,
	logger *logger.Logger,
) TestimonialService {
	return &testimonialService{
		repo:         repo,
		utilsService: utilsService,
		logger:       logger,
	}
}

func (s *testimonialService) CreateTestimonial(ctx context.Context, req testimonials.CreateTestimonialRequest) (*testimonials.TestimonialResponse, error) {

	if req.Rating < 1 || req.Rating > 5 {
		return nil, testimonials.ErrInvalidRating
	}

	testimonial := &entity.Testimonial{
		ID:           s.utilsService.GenerateULID(),
		CustomerName: strings.TrimSpace(req.CustomerName),
		Message:      strings.TrimSpace(req.Message),
		Rating:       req.Rating,
		IsApproved:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, testimonial); err != nil {
		s.logger.Error("Failed to create testimonial", logger.Fields{
			"error":         err.Error(),
			"customer_name": req.CustomerName,
		})
		return nil, err
	}

	s.logger.Info("Testimonial created successfully", logger.Fields{
		"id":            testimonial.ID,
		"customer_name": testimonial.CustomerName,
		"rating":        testimonial.Rating,
	})

	return s.entityToResponse(testimonial), nil
}

func (s *testimonialService) GetApprovedTestimonials(ctx context.Context) ([]testimonials.TestimonialResponse, error) {
	testimonialList, err := s.repo.GetApproved(ctx)
	if err != nil {
		s.logger.Error("Failed to get approved testimonials", logger.Fields{
			"error": err.Error(),
		})
		return nil, err
	}

	responses := make([]testimonials.TestimonialResponse, len(testimonialList))
	for i, testimonial := range testimonialList {
		responses[i] = *s.entityToResponse(&testimonial)
	}

	return responses, nil
}

func (s *testimonialService) GetAllTestimonials(ctx context.Context) ([]testimonials.TestimonialResponse, error) {
	testimonialList, err := s.repo.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get all testimonials", logger.Fields{
			"error": err.Error(),
		})
		return nil, err
	}

	responses := make([]testimonials.TestimonialResponse, len(testimonialList))
	for i, testimonial := range testimonialList {
		responses[i] = *s.entityToResponse(&testimonial)
	}

	return responses, nil
}

func (s *testimonialService) ApproveTestimonial(ctx context.Context, id string) error {
	testimonial, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if testimonial.IsApproved {
		return nil
	}

	testimonial.IsApproved = true
	testimonial.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, testimonial); err != nil {
		s.logger.Error("Failed to approve testimonial", logger.Fields{
			"error": err.Error(),
			"id":    id,
		})
		return err
	}

	s.logger.Info("Testimonial approved successfully", logger.Fields{
		"id":            id,
		"customer_name": testimonial.CustomerName,
	})

	return nil
}

func (s *testimonialService) RejectTestimonial(ctx context.Context, id string) error {
	testimonial, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !testimonial.IsApproved {
		return nil
	}

	testimonial.IsApproved = false
	testimonial.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, testimonial); err != nil {
		s.logger.Error("Failed to reject testimonial", logger.Fields{
			"error": err.Error(),
			"id":    id,
		})
		return err
	}

	s.logger.Info("Testimonial rejected successfully", logger.Fields{
		"id":            id,
		"customer_name": testimonial.CustomerName,
	})

	return nil
}

func (s *testimonialService) DeleteTestimonial(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete testimonial", logger.Fields{
			"error": err.Error(),
			"id":    id,
		})
		return err
	}

	s.logger.Info("Testimonial deleted successfully", logger.Fields{
		"id": id,
	})

	return nil
}

func (s *testimonialService) entityToResponse(testimonial *entity.Testimonial) *testimonials.TestimonialResponse {
	return &testimonials.TestimonialResponse{
		ID:           testimonial.ID,
		CustomerName: testimonial.CustomerName,
		Message:      testimonial.Message,
		Rating:       testimonial.Rating,
		IsApproved:   testimonial.IsApproved,
		CreatedAt:    testimonial.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
