package dto

import (
	"github.com/Gkemhcs/hookwave/server/internal/apierror"
	"github.com/google/uuid"
)

type CreateMessageInput struct {
	ApplicationID uuid.UUID	`json:"-"`
	EventType      string	`json:"event_type"`
	Payload        *map[string]string	`json:"payload"`
	Metadata       map[string]string	`json:"metadata"`
}

func(req *CreateMessageInput) Validate()error{
	if req.Payload==nil || len(*req.Payload)==0{
		return apierror.NewValidationError("message payload cannot be empty")
	}
	return nil 
}