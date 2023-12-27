package rest

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

var InternalServerError = ErrorResponse{
	Message: "internal server error",
	Status:  http.StatusInternalServerError,
}

var BadRequestError = ErrorResponse{
	Message: "bad request",
	Status:  http.StatusBadRequest,
}

var NotFoundError = ErrorResponse{
	Message: "not found",
	Status:  http.StatusNotFound,
}

var UnauthorizedError = ErrorResponse{
	Message: "unauthorized",
	Status:  http.StatusUnauthorized,
}

var ForbiddenError = ErrorResponse{
	Message: "forbidden",
	Status:  http.StatusForbidden,
}

type KeyError struct {
	Key     string `json:"key"`
	Message string `json:"message"`
}

type BadRequestErrorResponse struct {
	Message string `json:"message"`
	// JSON body key errors.
	Fields []KeyError `json:"fields,omitempty"`
	// GET parameter errors.
	Params []KeyError `json:"params,omitempty"`
}

func (e *BadRequestErrorResponse) Error() string {
	return BadRequestError.Error()
}

func RenderError(w http.ResponseWriter, r *http.Request, e error) {
	if er, ok := e.(*ErrorResponse); ok {
		render.Status(r, er.Status)
		render.JSON(w, r, er)
		return
	}
	if bre, ok := e.(*BadRequestErrorResponse); ok {
		render.Status(r, BadRequestError.Status)
		if bre.Message == "" {
			bre.Message = BadRequestError.Error()
		}
		render.JSON(w, r, bre)
		return
	}
	render.Status(r, InternalServerError.Status)
	render.JSON(w, r, InternalServerError)
}
