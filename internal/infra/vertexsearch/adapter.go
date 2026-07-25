// Package vertexsearch is deprecated. Use enterprise-search/internal/infra/agentsearch instead.
package vertexsearch

import (
	"enterprise-search/internal/infra/agentsearch"
)

// Deprecated: Use agentsearch.Config instead.
type Config = agentsearch.Config

// Deprecated: Use agentsearch.Adapter instead.
type Adapter = agentsearch.Adapter

// Deprecated: Use agentsearch.NewAdapter instead.
var NewAdapter = agentsearch.NewAdapter
