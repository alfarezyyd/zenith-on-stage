package validator

import (
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"zenith-on-stage/pkg/exception"

	universalTranslator "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

type ServiceImpl struct {
	validatorInstance *validator.Validate
	engTranslator     universalTranslator.Translator
}

func NewService(
	validatorInstance *validator.Validate,
	engTranslator universalTranslator.Translator) *ServiceImpl {
	return &ServiceImpl{
		validatorInstance: validatorInstance,
		engTranslator:     engTranslator}
}

// ValidateStruct - Validasi struct dengan opsi return error atau panic
func (validatorService *ServiceImpl) ValidateStruct(target interface{}) error {
	return validatorService.validatorInstance.Struct(target)
}

// ValidateVar - Validasi single variable dengan opsi return error atau panic
func (validatorService *ServiceImpl) ValidateVar(target interface{}, validatorTags string) error {
	return validatorService.validatorInstance.Var(target, validatorTags)
}

// ParseValidationError - Parsing error validasi ke dalam format yang lebih mudah dibaca
func (validatorService *ServiceImpl) ParseValidationError(validationError error, dtoStruct interface{}) {
	if validationError != nil {
		parsedMap := make(map[string]interface{})
		reflectType := reflect.TypeOf(dtoStruct)
		if reflectType.Kind() == reflect.Ptr {
			reflectType = reflectType.Elem()
		}
		for _, fieldError := range validationError.(validator.ValidationErrors) {
			fieldJSON := getJSONFieldName(reflectType, fieldError.StructField())
			fieldName := fieldError.StructField()
			translatedMessage := fieldError.Translate(validatorService.engTranslator)

			// Hapus nama field dari pesan error
			cleanMessage := strings.Replace(translatedMessage, fieldName, getPropertyFieldName(reflectType, fieldName), 1)
			cleanMessage = strings.TrimSpace(cleanMessage) // Hilangkan spasi berlebih

			parsedMap[fieldJSON] = cleanMessage
		}
		panic(exception.NewApplicationErrorSpecific(http.StatusBadRequest, exception.StatusValidationError, exception.MsgValidationError, parsedMap))
	}
}

// getPropertyFieldName mengembalikan nama field berdasarkan tag "property" atau memisahkan nama struct field berdasarkan huruf besar.
func getPropertyFieldName(t reflect.Type, fieldName string) string {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.Name == fieldName {
			// Cek apakah ada tag "property"
			propertyTag := field.Tag.Get("property")
			if propertyTag != "" {
				return propertyTag
			}

			// Jika tidak ada, pisahkan berdasarkan huruf besar (CamelCase → "Camel Case")
			return splitCamelCase(field.Name)
		}
	}
	return fieldName
}

// splitCamelCase memisahkan nama field yang dalam format CamelCase menjadi kata-kata terpisah.
func splitCamelCase(s string) string {
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	return re.ReplaceAllString(s, "$1 $2")
}

func getJSONFieldName(structType reflect.Type, fieldName string) string {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	field, found := structType.FieldByName(fieldName)
	if !found {
		return fieldName // fallback ke nama field struct
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		jsonTag = field.Tag.Get("form")
		if jsonTag == "" {
			return fieldName
		}
	}

	// json tag bisa seperti "username,omitempty", ambil sebelum koma
	parts := strings.Split(jsonTag, ",")
	return parts[0]
}
