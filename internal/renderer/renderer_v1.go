package renderer

import (
	"github.com/chubin/wttr.in/internal/domain"
	"github.com/chubin/wttr.in/internal/localization"
)

// Renderer Implementations (Stubs)
type V1Renderer struct{}

func (r *V1Renderer) Render(query domain.Query, localizer localization.Localizer) (domain.RenderOutput, error) {
	// Stub: To be implemented
	return domain.RenderOutput{}, nil
}
