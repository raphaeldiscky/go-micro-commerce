// Package service provides business logic for fulfillment operations.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/repository"
)

// FulfillmentServiceInterface defines the interface for fulfillment business operations.
type FulfillmentServiceInterface interface {
	// CreateFulfillment creates a new fulfillment record from order information
	CreateFulfillment(
		ctx context.Context,
		req *dto.CreateFulfillmentRequest,
	) (*dto.FulfillmentResponse, error)
	// UpdateFulfillmentStatus updates the status of a fulfillment
	UpdateFulfillmentStatus(
		ctx context.Context,
		fulfillmentID uuid.UUID,
		req dto.UpdateFulfillmentStatusRequest,
	) (*dto.FulfillmentResponse, error)
	// SetCarrierInfo sets carrier and shipping label information
	SetCarrierInfo(
		ctx context.Context,
		fulfillmentID uuid.UUID,
		req dto.SetCarrierInfoRequest,
	) (*dto.FulfillmentResponse, error)
	// SetDimensions sets package dimensions
	SetDimensions(
		ctx context.Context,
		fulfillmentID uuid.UUID,
		req dto.SetDimensionsRequest,
	) (*dto.FulfillmentResponse, error)
	// SetActualDelivery sets the actual delivery time
	SetActualDelivery(
		ctx context.Context,
		fulfillmentID uuid.UUID,
		req dto.SetActualDeliveryRequest,
	) (*dto.FulfillmentResponse, error)
	// GetFulfillmentByOrderID retrieves fulfillment by order ID
	GetFulfillmentByOrderID(
		ctx context.Context,
		orderID uuid.UUID,
	) (*dto.FulfillmentResponse, error)
	// GetFulfillmentByTrackingNumber retrieves fulfillment by tracking number
	GetFulfillmentByTrackingNumber(
		ctx context.Context,
		trackingNumber string,
	) (*dto.FulfillmentResponse, error)
	// HandleOrderFulfillmentRequested handles fulfillment requests from order service
	HandleOrderFulfillmentRequested(
		ctx context.Context,
		orderID uuid.UUID,
		trackingNumber string,
		shippingCost decimal.Decimal,
	) error
	// GetShippingRates retrieves shipping rates for a fulfillment request
	GetShippingRates(
		ctx context.Context,
		req *dto.GetShippingRatesRequest,
	) ([]dto.ShippingRateResponse, error)
	// CreateShipment creates a shipment with carrier and generates tracking number
	CreateShipment(
		ctx context.Context,
		fulfillmentID uuid.UUID,
		req *dto.CreateShipmentRequest,
	) (*dto.FulfillmentResponse, error)
	// UpdateTrackingStatus updates fulfillment status based on carrier tracking
	UpdateTrackingStatus(
		ctx context.Context,
		trackingNumber string,
	) (*dto.FulfillmentResponse, error)
}

// FulfillmentService implements the FulfillmentServiceInterface.
type FulfillmentService struct {
	dataStore                    repository.DataStore
	logger                       logger.Logger
	fulfillmentLifecycleProducer kafka.ProducerInterface
	carrierClient                client.CarrierClientInterface
}

// NewFulfillmentService creates a new instance of FulfillmentService.
func NewFulfillmentService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	fulfillmentLifecycleProducer kafka.ProducerInterface,
	carrierClient client.CarrierClientInterface,
) FulfillmentServiceInterface {
	return &FulfillmentService{
		dataStore:                    dataStore,
		logger:                       appLogger,
		fulfillmentLifecycleProducer: fulfillmentLifecycleProducer,
		carrierClient:                carrierClient,
	}
}

// CreateFulfillment creates a new fulfillment record from order information.
func (s *FulfillmentService) CreateFulfillment(
	ctx context.Context,
	req *dto.CreateFulfillmentRequest,
) (*dto.FulfillmentResponse, error) {
	res := new(dto.FulfillmentResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		fulfillmentRepo := ds.FulfillmentRepository()
		outboxRepo := ds.OutboxRepository()

		// Check if fulfillment already exists for this order
		existingFulfillment, err := fulfillmentRepo.FindByOrderID(ctx, req.OrderID)
		if err != nil {
			return httperror.NewInternalServerError("failed to check existing fulfillment")
		}

		if existingFulfillment != nil {
			// Fulfillment already exists, return existing fulfillment
			res = mapper.MapToFulfillmentResponse(existingFulfillment)

			return nil
		}

		// Create new fulfillment entity
		fulfillment, err := entity.NewFulfillment(
			req.OrderID,
			req.TrackingNumber,
			req.Currency,
			req.ShippingCost,
			req.WeightKG,
			req.EstimatedDeliveryAt,
		)
		if err != nil {
			return httperror.NewBadRequestError(
				fmt.Sprintf("failed to create fulfillment: %v", err),
			)
		}

		// Save fulfillment
		savedFulfillment, err := fulfillmentRepo.Create(ctx, fulfillment)
		if err != nil {
			return httperror.NewInternalServerError("failed to save fulfillment")
		}

		// Publish fulfillment created event
		evt := mq.NewFulfillmentLifecycleEvent(
			savedFulfillment,
		)

		payload, err := json.Marshal(evt)
		if err != nil {
			return httperror.NewInternalServerError("failed to marshal fulfillment event")
		}

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "fulfillment",
			AggregateID:   savedFulfillment.ID,
			EventType:     kafka.FulfillmentCreatedEventType,
			Topic:         kafka.FulfillmentLifecycleTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err := outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create outbox event")
		}

		res = mapper.MapToFulfillmentResponse(savedFulfillment)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// UpdateFulfillmentStatus updates the status of a fulfillment.
func (s *FulfillmentService) UpdateFulfillmentStatus(
	ctx context.Context,
	fulfillmentID uuid.UUID,
	req dto.UpdateFulfillmentStatusRequest,
) (*dto.FulfillmentResponse, error) {
	res := new(dto.FulfillmentResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		fulfillmentRepo := ds.FulfillmentRepository()
		outboxRepo := ds.OutboxRepository()

		// Get fulfillment
		fulfillment, err := fulfillmentRepo.FindByID(ctx, fulfillmentID)
		if err != nil {
			return httperror.NewInternalServerError("failed to get fulfillment")
		}

		if fulfillment == nil {
			return httperror.NewFulfillmentNotFoundError()
		}

		// Update status
		if err := fulfillment.UpdateStatus(req.Status); err != nil {
			return httperror.NewBadRequestError("failed to update fulfillment status")
		}

		// Save updated fulfillment
		updatedFulfillment, err := fulfillmentRepo.Update(ctx, fulfillment)
		if err != nil {
			return httperror.NewInternalServerError("failed to update fulfillment")
		}

		// Publish fulfillment status update event
		evt := mq.NewFulfillmentLifecycleEvent(
			updatedFulfillment,
		)

		payload, err := json.Marshal(evt)
		if err != nil {
			return httperror.NewInternalServerError("failed to marshal fulfillment event")
		}

		eventType := s.getEventTypeFromStatus(req.Status)

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "fulfillment",
			AggregateID:   updatedFulfillment.ID,
			EventType:     eventType,
			Topic:         kafka.FulfillmentLifecycleTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err := outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create fulfillment status event")
		}

		res = mapper.MapToFulfillmentResponse(updatedFulfillment)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// SetCarrierInfo sets carrier and shipping label information.
func (s *FulfillmentService) SetCarrierInfo(
	ctx context.Context,
	fulfillmentID uuid.UUID,
	req dto.SetCarrierInfoRequest,
) (*dto.FulfillmentResponse, error) {
	fulfillmentRepo := s.dataStore.FulfillmentRepository()

	fulfillment, err := fulfillmentRepo.FindByID(ctx, fulfillmentID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get fulfillment")
	}

	if fulfillment == nil {
		return nil, httperror.NewFulfillmentNotFoundError()
	}

	if err := fulfillment.SetCarrierInfo(req.Carrier, req.ShippingLabelURL); err != nil {
		return nil, httperror.NewBadRequestError("failed to set carrier info")
	}

	updatedFulfillment, err := fulfillmentRepo.Update(ctx, fulfillment)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to update fulfillment")
	}

	return mapper.MapToFulfillmentResponse(updatedFulfillment), nil
}

// SetDimensions sets package dimensions.
func (s *FulfillmentService) SetDimensions(
	ctx context.Context,
	fulfillmentID uuid.UUID,
	req dto.SetDimensionsRequest,
) (*dto.FulfillmentResponse, error) {
	fulfillmentRepo := s.dataStore.FulfillmentRepository()

	fulfillment, err := fulfillmentRepo.FindByID(ctx, fulfillmentID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get fulfillment")
	}

	if fulfillment == nil {
		return nil, httperror.NewFulfillmentNotFoundError()
	}

	if err := fulfillment.SetDimensions(&req.Dimensions); err != nil {
		return nil, httperror.NewBadRequestError("failed to set dimensions")
	}

	updatedFulfillment, err := fulfillmentRepo.Update(ctx, fulfillment)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to update fulfillment")
	}

	return mapper.MapToFulfillmentResponse(updatedFulfillment), nil
}

// SetActualDelivery sets the actual delivery time.
func (s *FulfillmentService) SetActualDelivery(
	ctx context.Context,
	fulfillmentID uuid.UUID,
	req dto.SetActualDeliveryRequest,
) (*dto.FulfillmentResponse, error) {
	fulfillmentRepo := s.dataStore.FulfillmentRepository()

	fulfillment, err := fulfillmentRepo.FindByID(ctx, fulfillmentID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get fulfillment")
	}

	if fulfillment == nil {
		return nil, httperror.NewFulfillmentNotFoundError()
	}

	if err := fulfillment.SetActualDelivery(req.ActualDeliveryAt); err != nil {
		return nil, httperror.NewBadRequestError("failed to set actual delivery")
	}

	updatedFulfillment, err := fulfillmentRepo.Update(ctx, fulfillment)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to update fulfillment")
	}

	return mapper.MapToFulfillmentResponse(updatedFulfillment), nil
}

// GetFulfillmentByOrderID retrieves fulfillment by order ID.
func (s *FulfillmentService) GetFulfillmentByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*dto.FulfillmentResponse, error) {
	fulfillmentRepo := s.dataStore.FulfillmentRepository()

	fulfillment, err := fulfillmentRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get fulfillment")
	}

	if fulfillment == nil {
		return nil, httperror.NewFulfillmentNotFoundError()
	}

	return mapper.MapToFulfillmentResponse(fulfillment), nil
}

// GetFulfillmentByTrackingNumber retrieves fulfillment by tracking number.
func (s *FulfillmentService) GetFulfillmentByTrackingNumber(
	ctx context.Context,
	trackingNumber string,
) (*dto.FulfillmentResponse, error) {
	fulfillmentRepo := s.dataStore.FulfillmentRepository()

	fulfillment, err := fulfillmentRepo.FindByTrackingNumber(ctx, trackingNumber)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get fulfillment")
	}

	if fulfillment == nil {
		return nil, httperror.NewFulfillmentNotFoundError()
	}

	return mapper.MapToFulfillmentResponse(fulfillment), nil
}

// HandleOrderFulfillmentRequested handles fulfillment requests from order service.
func (s *FulfillmentService) HandleOrderFulfillmentRequested(
	ctx context.Context,
	orderID uuid.UUID,
	trackingNumber string,
	shippingCost decimal.Decimal,
) error {
	// Create fulfillment record for the order
	req := dto.CreateFulfillmentRequest{
		OrderID:             orderID,
		TrackingNumber:      trackingNumber,
		ShippingCost:        shippingCost,
		WeightKG:            decimal.NewFromFloat(1.0),   // Default 1kg
		EstimatedDeliveryAt: time.Now().AddDate(0, 0, 7), // 7 days from now
	}

	_, err := s.CreateFulfillment(ctx, &req)
	if err != nil {
		s.logger.Errorf("Failed to create fulfillment for order %s: %v", orderID, err)

		return err
	}

	s.logger.Infof("Successfully created fulfillment record for order %s", orderID)

	return nil
}

// GetShippingRates retrieves shipping rates for a fulfillment request.
func (s *FulfillmentService) GetShippingRates(
	ctx context.Context,
	req *dto.GetShippingRatesRequest,
) ([]dto.ShippingRateResponse, error) {
	// Create shipping request for carrier client
	shipReq := dto.ShippingRequest{
		OrderID:     req.OrderID,
		FromAddress: req.FromAddress,
		ToAddress:   req.ToAddress,
		Package:     req.Package,
	}

	rates, err := s.carrierClient.GetRates(ctx, &shipReq)
	if err != nil {
		s.logger.Errorf("Failed to get shipping rates: %v", err)

		return nil, fmt.Errorf("failed to get shipping rates: %w", err)
	}

	// Convert carrier rates to response format
	response := make([]dto.ShippingRateResponse, len(rates))
	for i, rate := range rates {
		response[i] = dto.ShippingRateResponse(rate)
	}

	return response, nil
}

// CreateShipment creates a shipment with carrier and generates tracking number.
func (s *FulfillmentService) CreateShipment(
	ctx context.Context,
	fulfillmentID uuid.UUID,
	req *dto.CreateShipmentRequest,
) (*dto.FulfillmentResponse, error) {
	// Get existing fulfillment
	fulfillmentRepo := s.dataStore.FulfillmentRepository()

	existingFulfillment, err := fulfillmentRepo.FindByID(ctx, fulfillmentID)
	if err != nil {
		s.logger.Errorf("Failed to get fulfillment %s: %v", fulfillmentID, err)

		return nil, httperror.NewFulfillmentNotFoundError()
	}

	// Create shipping request
	insuranceAmount := decimal.Zero
	if req.InsuranceAmount != nil {
		insuranceAmount = *req.InsuranceAmount
	}

	shipReq := dto.ShippingRequest{
		OrderID:         existingFulfillment.OrderID,
		Carrier:         req.Carrier,
		Service:         req.Service,
		FromAddress:     req.FromAddress,
		ToAddress:       req.ToAddress,
		Package:         req.Package,
		InsuranceAmount: insuranceAmount,
		Signature:       req.Signature,
	}

	// Create shipment with carrier
	label, err := s.carrierClient.CreateShipment(ctx, &shipReq)
	if err != nil {
		s.logger.Errorf("Failed to create shipment with carrier: %v", err)

		return nil, fmt.Errorf("failed to create shipment: %w", err)
	}

	// Update fulfillment with shipping information
	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		fulfillmentRepo := ds.FulfillmentRepository()

		// Update with carrier information
		existingFulfillment.Carrier = &label.Carrier
		existingFulfillment.ShippingLabelURL = &label.LabelURL
		existingFulfillment.TrackingNumber = label.TrackingNumber
		existingFulfillment.Status = constant.FulfillmentStatusShipped
		existingFulfillment.UpdatedAt = time.Now().UTC()

		_, err := fulfillmentRepo.Update(ctx, existingFulfillment)

		return err
	})
	if err != nil {
		s.logger.Errorf("Failed to update fulfillment with shipping info: %v", err)

		return nil, fmt.Errorf("failed to update fulfillment: %w", err)
	}

	// Publish fulfillment shipped event
	event := mq.NewFulfillmentLifecycleEvent(
		existingFulfillment,
	)

	if err := s.fulfillmentLifecycleProducer.Send(ctx, event); err != nil {
		s.logger.Errorf("Failed to publish fulfillment shipped event: %v", err)
		// Don't fail the operation, just log the error
	}

	return mapper.MapToFulfillmentResponse(existingFulfillment), nil
}

// UpdateTrackingStatus updates fulfillment status based on carrier tracking.
func (s *FulfillmentService) UpdateTrackingStatus(
	ctx context.Context,
	trackingNumber string,
) (*dto.FulfillmentResponse, error) {
	// Get fulfillment by tracking number
	fulfillmentRepo := s.dataStore.FulfillmentRepository()

	existingFulfillment, err := fulfillmentRepo.FindByTrackingNumber(ctx, trackingNumber)
	if err != nil {
		s.logger.Errorf("Failed to get fulfillment by tracking number %s: %v", trackingNumber, err)

		return nil, httperror.NewFulfillmentNotFoundError()
	}

	// Get carrier tracking information
	carrier := ""
	if existingFulfillment.Carrier != nil {
		carrier = *existingFulfillment.Carrier
	}

	trackingInfo, err := s.carrierClient.GetTracking(ctx, trackingNumber, carrier)
	if err != nil {
		s.logger.Errorf("Failed to get tracking info for %s: %v", trackingNumber, err)

		return nil, fmt.Errorf("failed to get tracking information: %w", err)
	}

	// Update fulfillment status if it has changed
	if trackingInfo.Status != existingFulfillment.Status {
		err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
			fulfillmentRepo := ds.FulfillmentRepository()

			existingFulfillment.Status = trackingInfo.Status
			existingFulfillment.UpdatedAt = time.Now().UTC()

			// Set delivery time if delivered
			if trackingInfo.Status == constant.FulfillmentStatusDelivered &&
				trackingInfo.DeliveredAt != nil {
				existingFulfillment.ActualDeliveryAt = trackingInfo.DeliveredAt
			}

			_, err := fulfillmentRepo.Update(ctx, existingFulfillment)

			return err
		})
		if err != nil {
			s.logger.Errorf("Failed to update fulfillment status: %v", err)

			return nil, fmt.Errorf("failed to update fulfillment status: %w", err)
		}

		// Publish status update event
		event := mq.NewFulfillmentLifecycleEvent(existingFulfillment)

		if err := s.fulfillmentLifecycleProducer.Send(ctx, event); err != nil {
			s.logger.Errorf("Failed to publish fulfillment status update event: %v", err)
			// Don't fail the operation, just log the error
		}
	}

	return mapper.MapToFulfillmentResponse(existingFulfillment), nil
}

// getEventTypeFromStatus returns the appropriate event type based on fulfillment status.
func (s *FulfillmentService) getEventTypeFromStatus(status constant.FulfillmentStatus) string {
	switch status {
	case constant.FulfillmentStatusShipped:
		return kafka.FulfillmentShippedEventType
	case constant.FulfillmentStatusDelivered:
		return kafka.FulfillmentDeliveredEventType
	case constant.FulfillmentStatusCanceled:
		return kafka.FulfillmentCanceledEventType
	case constant.FulfillmentStatusPending:
		return kafka.FulfillmentCreatedEventType
	case constant.FulfillmentStatusProcessing:
		return kafka.FulfillmentProcessingEventType
	case constant.FulfillmentStatusInTransit:
		return kafka.FulfillmentInTransitEventType
	case constant.FulfillmentStatusReturned:
		return kafka.FulfillmentReturnedEventType
	default:
		return kafka.FulfillmentUpdatedEventType
	}
}
