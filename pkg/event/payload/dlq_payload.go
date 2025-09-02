// Package payload provides the commont event definitions for microservices.
package payload

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// DLQPayload holds the data for the DLQ event.
type DLQPayload struct {
	OutboxEventID   uuid.UUID          `json:"outbox_event_id"`
	AggregateType   string             `json:"aggregate_type"`
	AggregateID     uuid.UUID          `json:"aggregate_id"`
	OriginalTopic   string             `json:"original_topic"`
	OriginalPayload json.RawMessage    `json:"original_payload"`
	Reason          constant.DLQReason `json:"reason"`
	LastError       string             `json:"last_error"`
	Attempts        int64              `json:"attempts"`
	CreatedAt       time.Time          `json:"created_at"`
	LastProcessedAt *time.Time         `json:"last_processed_at"`
	FailedAt        time.Time          `json:"failed_at"`
}
