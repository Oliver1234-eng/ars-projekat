package main

import (
	poststore "ars-projekat/configstore"
	"ars-projekat/model"
	"errors"
	"github.com/gorilla/mux"
	"mime"
	"net/http"
)

type Service struct {
	store *poststore.ConfigStore
}

func (ts *Service) IdempotencyCheck(handlerFunc func(http.ResponseWriter, *http.Request) string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		idempotencyKey := req.Header.Get("Idempotency-Key")
		if idempotencyKey == "" {
			http.Error(w, "Missing Idempotency-Key header", http.StatusBadRequest)
			return
		}

		keyExists, storedKey, err := ts.store.IdempotencyKeyExists(idempotencyKey)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if keyExists {
			model.RenderJSON(w, storedKey)
			return
		}

		id := handlerFunc(w, req)
		if id != "" {
			ts.store.SaveIdempotencyKey(idempotencyKey, id)
		}
	}
}

func (ts *Service) createConfigHandler(w http.ResponseWriter, req *http.Request) string {
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	if mediatype != "application/json" {
		err := errors.New("Expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return ""
	}

	rt, err := model.DecodeConfig(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	id, err := ts.store.CreateConfig(rt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}

	model.RenderJSON(w, id)

	return id

}

func (ts *Service) createConfigVersionHandler(w http.ResponseWriter, req *http.Request) string {
	id := mux.Vars(req)["uuid"]

	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	if mediatype != "application/json" {
		err := errors.New("Expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return ""
	}

	rt, err := model.DecodeConfig(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	_, err = ts.store.CreateConfigVersion(id, rt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}

	model.RenderJSON(w, id)

	return id
}

func (ts *Service) createGroupHandler(w http.ResponseWriter, req *http.Request) string {
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	if mediatype != "application/json" {
		err := errors.New("Expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return ""
	}

	rt, err := model.DecodeGroup(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	id, err := ts.store.CreateGroup(rt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}

	model.RenderJSON(w, id)

	return id
}

func (ts *Service) getConfigHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["uuid"]
	ver := mux.Vars(req)["ver"]

	config, err := ts.store.GetConfig(id, ver)
	if err != nil {
		err := errors.New("key not found")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	model.RenderJSON(w, config)
}

func (ts *Service) getGroupHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["uuid"]
	ver := mux.Vars(req)["ver"]
	labels := model.DecodeQueryLabels(req.URL.Query())

	group, err := ts.store.GetGroup(id, ver, labels)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model.RenderJSON(w, group)
}

func (ts *Service) delConfigHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["uuid"]
	ver := mux.Vars(req)["ver"]

	r, err := ts.store.DeleteConfig(id, ver)
	if err != nil {
		err := errors.New("key not found")
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	model.RenderJSON(w, r)
}

func (ts *Service) addConfigToGroupHandler(w http.ResponseWriter, req *http.Request) string {
	id := mux.Vars(req)["uuid"]
	ver := mux.Vars(req)["ver"]

	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	if mediatype != "application/json" {
		err := errors.New("Expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return ""
	}

	groupConfig, err := model.DecodeGroupConfig(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	_, err = ts.store.AddConfigToGroup(id, ver, groupConfig)

	if err != nil {
		err := errors.New("key not found")
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	model.RenderJSON(w, id)

	return id
}

func (ts *Service) createGroupVersionHandler(w http.ResponseWriter, req *http.Request) string {
	id := mux.Vars(req)["uuid"]

	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	if mediatype != "application/json" {
		err := errors.New("Expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return ""
	}

	rt, err := model.DecodeGroup(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	_, err = ts.store.CreateGroupVersion(id, rt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}

	model.RenderJSON(w, id)

	return id
}

func (ts *Service) delGroupHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["uuid"]
	ver := mux.Vars(req)["ver"]

	r, err := ts.store.DeleteGroup(id, ver)
	if err != nil {
		err := errors.New("key not found")
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	model.RenderJSON(w, r)
}
