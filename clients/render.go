package clients

import (
	"io"

	"github.com/ONSdigital/dis-design-system-go/model"
)

// RenderClient is an interface with methods for require for rendering a template
type RenderClient interface {
	BuildPage(w io.Writer, pageModel interface{}, templateName string)
	NewBasePageModel() model.Page
}
