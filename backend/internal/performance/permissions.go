package performance

import "mapp/internal/platform/authz"

// Table A rows since D-367 reserved them for 046; these aliases keep the area reading as its own vocabulary.
const (
	PermissionImport = authz.PerformanceImport
	PermissionLink   = authz.PerformanceLink
)
