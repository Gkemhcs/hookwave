package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Gkemhcs/hookwave/server/internal/domain"
	"github.com/Gkemhcs/hookwave/server/internal/repository"
	"github.com/Gkemhcs/hookwave/server/internal/transactor"
	"github.com/google/uuid"
)

const (
	DefaultEventType   = "*"
	DefaultMaxAttempts = 5

)

type MessageRepository interface {
	CreateDeliveries(ctx context.Context, arg []repository.CreateDeliveriesParams) (int64, error)
	CreateMessage(ctx context.Context, arg repository.CreateMessageParams) (repository.Message, error)
	GetAllEndpointsByApplicationID(ctx context.Context, applicationID uuid.UUID) ([]repository.Endpoint, error)
}

type MessageService struct {
	*transactor.Transactor
}
func NewMessageService(transactor *transactor.Transactor)*MessageService{
	return &MessageService{
		Transactor: transactor,
	}
}

type EndpointFilter struct {
	EventType string
}

func (f *EndpointFilter) isActive(endpoint repository.Endpoint)bool{
	return endpoint.IsActive
}
func(f *EndpointFilter) checkEventType(endpoint repository.Endpoint)bool{
	for _,event:=range endpoint.EventTypes{
		if event==DefaultEventType{
			return true
		}
	}
	return true
}
func (f *EndpointFilter) Match(endpoint repository.Endpoint) bool {

	return f.isActive(endpoint) && f.checkEventType(endpoint)
}

type CreateMessageParams struct {
	ApplicationID  uuid.UUID
	EventType      string
	Payload        map[string]string
	Metadata       map[string]string
}

func (s *MessageService) CreateMessage(ctx context.Context, args *CreateMessageParams) (*domain.Message, error) {

	var msg repository.Message
	err := s.Execute(ctx, func(queries *repository.Queries) error {
		eventType := DefaultEventType
		if args.EventType != "" {
			eventType = args.EventType
		}
		createMessageParams := repository.CreateMessageParams{
			ID:             uuid.New(),
			ApplicationID:  args.ApplicationID,
			EventType:      eventType,
			IdempotencyKey: uuid.NewString(),
			CreatedAt:      time.Now().UTC(),
		}
		if args.Payload != nil {
			payload, err := json.Marshal(args.Payload)
			if err != nil {
				return err
			}
			createMessageParams.Payload = payload
		}
		
		if args.Metadata!=nil{
			metadata,err:=json.Marshal(args.Metadata)
			if err!=nil{
				return repository.MapError(err)
			}
			rawMessage:=json.RawMessage(metadata)
			createMessageParams.Metadata=&rawMessage
		}
		var err error
		msg, err = queries.CreateMessage(ctx, createMessageParams)
		if err != nil {
			return err
		}
		endpoints, err := queries.GetAllEndpointsByApplicationID(ctx, args.ApplicationID)
		if err != nil {
			return err
		}
		eligibleEndpoints := s.filterEndpoints(endpoints, EndpointFilter{EventType: DefaultEventType})
		if len(eligibleEndpoints) != 0 {
			deliveries := s.buildDeliveries(eligibleEndpoints, msg.ID)
			_, err := queries.CreateDeliveries(ctx, deliveries)
			if err != nil {
				return err
			}

		}
		return nil
	})
	if err != nil {
		return nil, repository.MapError(err)
	}
	return s.buildMessage(msg)

}
func (s *MessageService) buildMessage(msgFromRepo repository.Message) (*domain.Message, error) {
	message := &domain.Message{
		ID:            msgFromRepo.ID,
		ApplicationID: msgFromRepo.ApplicationID,
		EventType:     msgFromRepo.EventType,
		CreatedAt:     msgFromRepo.CreatedAt,
	}
	var payload map[string]string
	err := json.Unmarshal(msgFromRepo.Payload, &payload)
	if err != nil {
		return nil, fmt.Errorf("parse payload: %w", err)
	}
	message.Payload = payload
	message.IdempotencyKey = msgFromRepo.IdempotencyKey
	return message, nil
}

func (s *MessageService) buildDeliveries(endpoints []repository.Endpoint, messageID uuid.UUID) []repository.CreateDeliveriesParams {
	var deliveries []repository.CreateDeliveriesParams
	for _, endpoint := range endpoints {
		deliveries = append(deliveries, repository.CreateDeliveriesParams{
			ID:          uuid.New(),
			MessageID:   messageID,
			EndpointID:  endpoint.ID,
			MaxAttempts: DefaultMaxAttempts,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		})
	}
	return deliveries
}

func (s *MessageService) filterEndpoints(endpoints []repository.Endpoint, filter EndpointFilter) []repository.Endpoint {
	var filteredEndpoints []repository.Endpoint
	for _, endpoint := range endpoints {
		if filter.Match(endpoint) {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}
	return filteredEndpoints
}
