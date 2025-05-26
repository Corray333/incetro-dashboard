package utils

import "github.com/google/uuid"

// ParseUUIDOrNil parses a string into a uuid.UUID, returning uuid.Nil if the string is not a valid UUID.
func ParseUUIDOrNil(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}
