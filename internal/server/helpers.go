package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

// generateName creates a random name for use in identifiers
func generateName(s string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", s, hex.EncodeToString(b))
}

// sortDirection will return a mongo bson sortable.
func sortDirection(s string) bson.D {
	switch strings.ToLower(s) {
	case "asc":
		return bson.D{{Key: "_id", Value: -1}}
	case "desc":
		return bson.D{{Key: "_id", Value: 1}}
	default:
		return bson.D{{Key: "_id", Value: -1}}
	}
}
