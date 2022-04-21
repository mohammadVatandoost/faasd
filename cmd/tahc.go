package cmd

import "github.com/openfaas/faasd/internal/lru"

const (
	UseTAHC               = false
	TAHCCacheSize         = 10
)

var TAHCCache *lru.Cache

