package main

import (
	"github.com/google/uuid"
)

type User struct {
	id uuid.UUID
	ip string
}

// VoteOptions Represents a vote sent to the user
type VoteOptions struct {
	ID uuid.UUID `json:"id"`
	A  string    `json:"a"`
	B  string    `json:"b"`
}
