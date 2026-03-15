package interfaces

// ShortcodeGenerator creates unique shortcodes for shortened URLs.
type ShortcodeGenerator interface {
	// GenerateShortcode returns a new unique shortcode string.
	GenerateShortcode() (string, error)
}
