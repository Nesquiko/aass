package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/Nesquiko/aass/common/server/api"
)

const (
	ContentType            = "Content-Type"
	ApplicationJSON        = "application/json"
	ApplicationProblemJSON = "application/problem+json"
	MaxBytes               = 1_048_576

	EncodingError   = "unexptected encoding error"
	UnexpectedError = "unexptected error"
)

func Encode[T any](w http.ResponseWriter, status int, response T) {
	EncodeWithContentType(w, status, response, ApplicationJSON)
}

func EncodeError(w http.ResponseWriter, err *ApiError) {
	EncodeWithContentType(w, err.Status, err.ErrorDetail, ApplicationProblemJSON)
}

func EncodeWithContentType[T any](
	w http.ResponseWriter,
	code int,
	response T,
	contentType string,
) {
	w.Header().Set(ContentType, contentType)
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error(
			UnexpectedError,
			slog.String("where", "encode"),
			slog.String("error", err.Error()),
		)
		http.Error(w, EncodingError, http.StatusInternalServerError)
	}
}

func Decode[T any](w http.ResponseWriter, r *http.Request) (T, *ApiError) {
	dst, err := decode[T](w, r)
	var decErr *DecodeErr
	if errors.As(err, &decErr) {
		return dst, decodeErrToApiErrorWithCode(decErr, decErr.code)
	}
	if err != nil {
		return dst, decodeErrToApiError(err)
	}
	return dst, nil
}

type DecodeErr struct {
	err  error
	code string
}

func (e *DecodeErr) Error() string {
	return e.err.Error()
}

func decode[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var dst T
	r.Body = http.MaxBytesReader(w, r.Body, int64(MaxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError
		var invalidUnmarshalErr *json.InvalidUnmarshalError

		invalidFieldPrefix := "json: unknown field "
		largeBodyErrorStr := "http: request body too large"

		switch {
		case errors.As(err, &syntaxErr):
			return dst, fmt.Errorf(
				"body contains badly-formed JSON (at character %d)",
				syntaxErr.Offset,
			)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return dst, errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeErr):
			if unmarshalTypeErr.Field != "" {
				return dst, fmt.Errorf(
					"body contains incorrect JSON type for field %q",
					unmarshalTypeErr.Field,
				)
			}
			return dst, fmt.Errorf(
				"body contains incorrect JSON type (at character %d)",
				unmarshalTypeErr.Offset,
			)

		case errors.Is(err, io.EOF):
			return dst, errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), invalidFieldPrefix):
			fieldName := strings.TrimPrefix(err.Error(), invalidFieldPrefix)
			return dst, &DecodeErr{
				fmt.Errorf("body contains unknown key %s", fieldName),
				ValidationErrorCode,
			}

		case err.Error() == largeBodyErrorStr:
			return dst, fmt.Errorf("body must not be larger than %d bytes", MaxBytes)

		case errors.As(err, &invalidUnmarshalErr):
			return dst, fmt.Errorf("invalid unmarshal target")

		default:
			return dst, err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return dst, errors.New("body must only contain a single JSON value")
	}

	return dst, nil
}

const (
	DecodingErrorCode  = "undecodable.request"
	DecodingErrorTitle = "Request couldn't be decoded"
)

func decodeErrToApiError(err error) *ApiError {
	return decodeErrToApiErrorWithCode(err, DecodingErrorCode)
}

func decodeErrToApiErrorWithCode(err error, code string) *ApiError {
	return &ApiError{
		ErrorDetail: api.ErrorDetail{
			Code:   code,
			Detail: err.Error(),
			Status: http.StatusBadRequest,
			Title:  DecodingErrorTitle,
		},
	}
}

const (
	SchemaValidationErrorCode = "invalid.request.schema"
	ValidationErrorCode       = "invalid.request"
	ValidationErrorTitle      = "Request doesn't comply with schema"
	ValidationErrorDetail     = "Validation failed, %s"
)

func validationError(path string, reason string, statusCode int, schema *string) *ApiError {
	additionalProperties := make(map[string]any)
	if path != "" {
		additionalProperties["path"] = path
	}
	if reason != "" {
		additionalProperties["reason"] = reason
	}
	if schema != nil {
		additionalProperties["schema"] = schema
	}
	if len(additionalProperties) == 0 {
		additionalProperties = nil
	}

	return &ApiError{
		ErrorDetail: api.ErrorDetail{
			Code:                 SchemaValidationErrorCode,
			Detail:               fmt.Sprintf(ValidationErrorDetail, reason),
			Status:               statusCode,
			Title:                ValidationErrorTitle,
			AdditionalProperties: additionalProperties,
		},
	}
}

func InternalServerError() *ApiError {
	return &ApiError{
		ErrorDetail: api.ErrorDetail{
			Code:   "internal.server.error",
			Title:  "Internal Server Error",
			Detail: "Unexpected error on server",
			Status: http.StatusInternalServerError,
		},
	}
}

const (
	NotFoundCode         = "%s.not.found"
	NotFoundTitleFormat  = "%s was not found"
	NotFoundDetailFormat = "%s with id '%s' was not found"
)

func NotFoundId(resoure string, id uuid.UUID) *ApiError {
	return NotFound(resoure, id.String())
}

func NotFound(resoure, id string) *ApiError {
	return &ApiError{
		ErrorDetail: api.ErrorDetail{
			Code:   fmt.Sprintf(NotFoundCode, strings.ToLower(resoure)),
			Title:  fmt.Sprintf(NotFoundTitleFormat, resoure),
			Detail: fmt.Sprintf(NotFoundDetailFormat, resoure, id),
			Status: http.StatusNotFound,
		},
	}
}

const (
	InvalidParamErrorCode   = "invalid.path.param"
	InvalidParamErrorTitle  = "Invalid path param"
	InvalidParamErrorDetail = "Invalid path param %q: %q"
)

func fromInvalidParamErr(e *api.InvalidParamFormatError, invalidParam string) *ApiError {
	return &ApiError{
		ErrorDetail: api.ErrorDetail{
			Code:   InvalidParamErrorCode,
			Title:  InvalidParamErrorTitle,
			Detail: fmt.Sprintf(InvalidParamErrorDetail, e.ParamName, invalidParam),
			Status: http.StatusBadRequest,
		},
	}
}

const (
	RequiredParamErrorCode   = "required.path.param"
	RequiredParamErrorTitle  = "Required path param"
	RequiredParamErrorDetail = "Required path param %q is missing"
)

func fromRequiredParamErr(e *api.RequiredParamError) *ApiError {
	return &ApiError{
		ErrorDetail: api.ErrorDetail{
			Code:   RequiredParamErrorCode,
			Title:  RequiredParamErrorTitle,
			Detail: fmt.Sprintf(RequiredParamErrorDetail, e.ParamName),
			Status: http.StatusBadRequest,
		},
	}
}
