package translate

import (
	trans "github.com/aerokite/go-google-translate/pkg"
)

// WithGoogle translate a string with google translate
func WithGoogle(text string, language string) (translation string, err error) {
	// To be improved
	if language != "" && language != "en" {
		req := &trans.TranslateRequest{
			SourceLang: "en",
			TargetLang: language,
			Text:       text,
		}

		// translate
		translation, err := trans.Translate(req)
		if err == nil {
			return translation, err
		}
	}
	return text, err
}
