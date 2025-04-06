package errs

import "errors"

var (
	ErrCompileHashingRegex = errors.New("failed to compile hashtag regex")
	ErrCompileMentionRegex = errors.New("failed to compile mention regex")
)
