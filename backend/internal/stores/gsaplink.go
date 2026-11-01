package stores

import "mapp/internal/measure/gsaplink"

// GsapLinkStore reads what the GSAP link report needs beyond the year (feature 044).
type GsapLinkStore interface {
	gsaplink.Store
}
