package helpers

import (
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

func GetErrorMessages(validate *validator.Validate, errs error) string {
	eng := en.New()
	uni := ut.New(eng, eng)
	trans, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, trans)

	var errorMessages []string
	for _, e := range errs.(validator.ValidationErrors) {
		errorMessages = append(errorMessages, e.Translate(trans))
	}
	return strings.Join(errorMessages, ", ")
}
