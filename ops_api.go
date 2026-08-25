package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func newOpsRouter(service *OpsService) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ops/records", opsRecordsHandler(service))
	mux.HandleFunc("/ops/records/", opsRecordHandler(service))
	mux.HandleFunc("/ops/snapshot", opsSnapshotHandler(service))
	mux.HandleFunc("/ops/rules", opsRulesHandler())
	return mux
}

func opsRecordsHandler(service *OpsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opsNoStore(w)
		switch r.Method {
		case http.MethodGet:
			query := OpsQuery{
				Subject:  r.URL.Query().Get("subject"),
				Status:   OpsStatus(r.URL.Query().Get("status")),
				Priority: OpsPriority(r.URL.Query().Get("priority")),
				Owner:    r.URL.Query().Get("owner"),
			}
			if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
				query.Page = page
			}
			if size, err := strconv.Atoi(r.URL.Query().Get("pageSize")); err == nil {
				query.PageSize = size
			}
			page, err := service.Search(r.Context(), query)
			if err != nil {
				opsJSON(w, opsHTTPStatus(err), map[string]string{"error": err.Error()})
				return
			}
			opsJSON(w, http.StatusOK, page)
		case http.MethodPost:
			var record OpsRecord
			if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
				opsJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
				return
			}
			created, err := service.Create(r.Context(), record)
			if err != nil {
				opsJSON(w, opsHTTPStatus(err), map[string]string{"error": err.Error()})
				return
			}
			opsJSON(w, http.StatusCreated, created)
		default:
			opsJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	}
}

func opsRecordHandler(service *OpsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opsNoStore(w)
		if r.Method != http.MethodPost {
			opsJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		id := opsPathID(r.URL.Path, "/ops/records/")
		if !strings.HasSuffix(id, "/status") || strings.Contains(strings.TrimSuffix(id, "/status"), "/") {
			opsJSON(w, http.StatusNotFound, map[string]string{"error": "operations record not found"})
			return
		}
		id = strings.TrimSuffix(id, "/status")
		var body struct {
			Status   OpsStatus `json:"status"`
			Expected int       `json:"expectedRevision"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !opsStatusValid(body.Status) {
			opsJSON(w, http.StatusBadRequest, map[string]string{"error": "valid status is required"})
			return
		}
		updated, err := service.Transition(r.Context(), id, body.Expected, body.Status, opsActorFromRequest(r))
		if err != nil {
			opsJSON(w, opsHTTPStatus(err), map[string]string{"error": err.Error()})
			return
		}
		opsJSON(w, http.StatusOK, updated)
	}
}

func opsSnapshotHandler(service *OpsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			opsJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		opsNoStore(w)
		opsJSON(w, http.StatusOK, service.Snapshot())
	}
}

func opsRulesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			opsJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		opsNoStore(w)
		opsJSON(w, http.StatusOK, map[string]any{"rules": opsRules(), "total": len(opsRules())})
	}
}

func opsHTTPStatus(err error) int {
	switch {
	case opsIsNotFound(err):
		return http.StatusNotFound
	case opsIsConflict(err), opsIsTransition(err):
		return http.StatusConflict
	case opsIsInvalid(err):
		return http.StatusBadRequest
	case opsIsPolicy(err):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
