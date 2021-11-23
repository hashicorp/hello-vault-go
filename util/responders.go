package util

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	HeaderContentType   = "Content-Type"
	MIMEApplicationJSON = "application/json"
)

// JSONResponder prepares and sends a JSON response
func JSONResponder(code int, i interface{}, w http.ResponseWriter, r *http.Request) {
	if i == nil {
		i = []byte{}
	}

	j, err := json.Marshal(i)
	if err != nil {
		ErrorResponder(err, w, r)
	}
	w.Header().Set(HeaderContentType, MIMEApplicationJSON)
	w.WriteHeader(code)

	_, err = w.Write(j)
	if err != nil {
		ErrorResponder(err, w, r)
	}

	log.Println("success", r.Method, r.URL.Path, code)
}

// ErrorResponder prepares and sends an error response defaulting to a generic 500
func ErrorResponder(err error, w http.ResponseWriter, r *http.Request) {
	handledError := new(Error)

	switch err.(type) {
	case *Error:
		handledError = err.(*Error)
	default:
		handledError = &InternalServerError
		handledError.Internal = err
	}

	w.WriteHeader(handledError.Code)
	w.Header().Add(HeaderContentType, MIMEApplicationJSON)

	resp, err := json.Marshal(handledError.Response)
	if err != nil {
		_, err = w.Write([]byte("Our technical team has been notified."))
		if err != nil {
			log.Println("failed writing error responder")
		}
	}

	_, err = w.Write(resp)
	if err != nil {
		log.Println("failed writing error responder")
	}

	log.Println("error", r.Method, r.URL.Path, handledError.Code, handledError.Error())

	r.Context().Done()
}
