// Package pageutils provides utility functions for cursor-based pagination.
package pageutils

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/bytedance/sonic"
)

// CursorData represents the data structure for cursor-based pagination.
type CursorData struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Value     string `json:"value,omitempty"`
}

// EncodeCursor encodes cursor data to a base64 string.
func EncodeCursor(data *CursorData) (string, error) {
	if data == nil {
		return "", nil
	}

	jsonData, err := sonic.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor data: %w", err)
	}

	encoded := base64.URLEncoding.EncodeToString(jsonData)

	return encoded, nil
}

// DecodeCursor decodes a base64 encoded cursor string to CursorData.
func DecodeCursor(cursor string) (*CursorData, error) {
	if cursor == "" {
		return nil, errors.New("cursor is empty")
	}

	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cursor: %w", err)
	}

	var data CursorData
	if err = sonic.Unmarshal(decoded, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cursor data: %w", err)
	}

	return &data, nil
}

// GenerateNextCursor creates a cursor for the next page based on the last item.
func GenerateNextCursor(id string, timestamp int64, value string) (string, error) {
	cursorData := &CursorData{
		ID:        id,
		Timestamp: timestamp,
		Value:     value,
	}

	return EncodeCursor(cursorData)
}

// GeneratePrevCursor creates a cursor for the previous page based on the first item.
func GeneratePrevCursor(id string, timestamp int64, value string) (string, error) {
	cursorData := &CursorData{
		ID:        id,
		Timestamp: timestamp,
		Value:     value,
	}

	return EncodeCursor(cursorData)
}
