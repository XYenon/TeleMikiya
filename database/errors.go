package database

import "errors"

var ErrNotAllowedToClearEmbedding = errors.New("embedding dimensions changed, but clearing embedding is not allowed")
