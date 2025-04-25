package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

type EventHashData struct {
	ProductID         uint
    EventType        string
    EventData        string
    CreatedAt        time.Time
    CreatedBy        uint
    PreviousEventHash string
}

func ComputeEventHash (data EventHashData) (string, error) {
	// Convert the data to json
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "Cannnot marhsal event data", err
	}

	// Pass it to sha
	hash  := sha256.Sum256(jsonData)

	return hex.EncodeToString(hash[:]), nil
}

