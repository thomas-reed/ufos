package server

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorRes struct {
	Err string `json:"error"`
}

func respondWithError(w http.ResponseWriter, errCode int, errMsg string, err error) {
	if err != nil {
		log.Printf("%s: %v", errMsg, err)
	}
	invalidJson := errorRes{
		Err: errMsg,
	}
	jsonResponse(w, errCode, invalidJson)
}

func jsonResponse(w http.ResponseWriter, code int, payload interface{}) {
	// No body for 204, 304
	if code == http.StatusNoContent || code == http.StatusNotModified {
		w.WriteHeader(code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write(json)
}
