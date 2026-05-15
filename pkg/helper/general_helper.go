package helper

import (
	"context"
	"encoding/json"
	"errors"
	"zenith-on-stage/pkg/exception"
	"zenith-on-stage/pkg/logger"

	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CheckErrorOperation(indicatedError error, applicationError *exception.ApplicationError) bool {

	if errors.Is(indicatedError, context.Canceled) {
		// client aborted request → ignore / log ringan
		return false
	}
	if indicatedError != nil {
		logger.Debug(indicatedError)
		panic(applicationError)
		return true
	}

	return false
}

func CheckPointerWrapper[T any](targetChecking *T, renderPayload func()) {
	if targetChecking != nil {
		renderPayload()
	}
}

func ExtractIndexedFiles(
	ginContext *gin.Context,
	prefixKey string,
	suffixKey string,
	expectedLen int,
) ([]*multipart.FileHeader, error) {

	multipartForm, err := ginContext.MultipartForm()
	CheckErrorOperation(err, exception.NewApplicationError(http.StatusBadRequest, exception.ErrPayloadInvalid))
	fileHeadersArray := make([]*multipart.FileHeader, expectedLen)

	for key, fileHeaders := range multipartForm.File {
		if strings.HasPrefix(key, prefixKey) &&
			strings.HasSuffix(key, suffixKey) {

			idxStr := key[len(prefixKey):strings.Index(key, "]")]
			idxInt, err := strconv.Atoi(idxStr)
			if err != nil || idxInt < 0 || idxInt >= expectedLen {
				continue
			}

			if len(fileHeaders) > 0 {
				fileHeadersArray[idxInt] = fileHeaders[0]
			}
		}
	}

	return fileHeadersArray, nil
}

func ExtractRequestMeta(ginContext *gin.Context) (string, string) {
	ipAddress, isIpAddressExists := ginContext.Get("ipAddress")
	userAgent, isUserAgentExists := ginContext.Get("userAgent")
	if isUserAgentExists && isIpAddressExists {
		return ipAddress.(string), userAgent.(string)
	}
	return "", ""
}

func TakePointer[T any](value T) *T {
	return &value
}

func DebugArrPointer[T any](arrayOfPointer []*T) {
	for i, pointerData := range arrayOfPointer {
		if pointerData == nil {
			logger.Debug("[%d] nil pointer", i)
			continue
		}

		jsonBytes, err := json.MarshalIndent(pointerData, "", "  ")
		if err != nil {
			logger.Debug("[%d] Marshal error: %v", i, err)
			continue
		}

		logger.Debug("[%d]\n%s", i, string(jsonBytes))
	}
}
func DiffEntity(before interface{}, after interface{}) map[string]map[string]interface{} {
	diff := make(map[string]map[string]interface{})

	getValue := func(i interface{}) reflect.Value {
		v := reflect.ValueOf(i)
		if v.Kind() == reflect.Ptr {
			return v.Elem()
		}
		return v
	}

	beforeVal := getValue(before)
	afterVal := getValue(after)
	beforeType := beforeVal.Type()

	for i := 0; i < beforeVal.NumField(); i++ {
		field := beforeType.Field(i)
		fieldName := field.Name

		beforeField := beforeVal.Field(i).Interface()
		afterField := afterVal.Field(i).Interface()

		// Skip complex types: struct, slice, map, pointer
		kind := beforeVal.Field(i).Kind()
		if kind == reflect.Struct || kind == reflect.Slice || kind == reflect.Map || kind == reflect.Ptr {
			continue
		}

		if reflect.DeepEqual(beforeField, afterField) {
			continue
		}

		diff[fieldName] = map[string]interface{}{
			"before": beforeField,
			"after":  afterField,
		}
	}

	return diff
}

func DiffUint64Slice(beforeIds []uint64, afterIds []uint64) (added []uint64, removed []uint64, unchanged []uint64) {
	beforeMap := map[uint64]bool{}
	afterMap := map[uint64]bool{}
	unchanged = []uint64{}

	for _, beforeId := range beforeIds {
		beforeMap[beforeId] = true
	}

	for _, afterId := range afterIds {
		afterMap[afterId] = true
		if !beforeMap[afterId] {
			added = append(added, afterId)
		} else {
			unchanged = append(unchanged, afterId)
		}
	}

	for _, beforeId := range beforeIds {
		if !afterMap[beforeId] {
			removed = append(removed, beforeId)
		}
	}

	return
}

// Convert struct → map[string]interface{} jika bukan nil
func NormalizeStruct(sourceStruct interface{}) map[string]interface{} {
	if sourceStruct == nil {
		return nil
	}

	val := reflect.ValueOf(sourceStruct)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]interface{})
	valType := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := valType.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		kind := fieldVal.Kind()
		switch kind {
		case reflect.Struct, reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface:
			// Skip nested structs/slices/maps
			continue
		default:
			result[field.Name] = fieldVal.Interface()
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ConvertIntoSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
