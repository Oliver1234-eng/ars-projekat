package main

import (
	poststore "ars-projekat/configstore"
	"ars-projekat/model"
	tracer "ars-projekat/tracer"
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"io"
	"mime"
	"net/http"
)

type Service struct {
	store  *poststore.ConfigStore
	tracer opentracing.Tracer
	closer io.Closer
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
			model.RenderJSONOld(w, storedKey)
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

	model.RenderJSONOld(w, id)

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

	model.RenderJSONOld(w, id)

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

	model.RenderJSONOld(w, id)

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

	model.RenderJSONOld(w, config)
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

	model.RenderJSONOld(w, group)
}

func (ts *Service) delConfigHandler(w http.ResponseWriter, req *http.Request) {
	span := tracer.StartSpanFromRequest("delConfigHandler", ts.tracer, req)
	defer span.Finish()

	span.LogFields(
		tracer.LogString("handler", fmt.Sprintf("handling delete config at %s\n", req.URL.Path)),
	)

	id := mux.Vars(req)["uuid"]
	ver := mux.Vars(req)["ver"]

	ctx := tracer.ContextWithSpan(context.Background(), span)
	r, err := ts.store.DeleteConfig(ctx, id, ver)
	if err != nil {
		err := errors.New("key not found")
		tracer.LogError(span, err)
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	model.RenderJSON(ctx, w, r)
}

func (ts *Service) addConfigToGroupHandler(w http.ResponseWriter, req *http.Request) string {
	span := tracer.StartSpanFromRequest("addConfigToGroupHandler", ts.tracer, req)
	defer span.Finish()

	span.LogFields(
		tracer.LogString("handler", fmt.Sprintf("handling add config to group at %s\n", req.URL.Path)),
	)

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

	ctx := tracer.ContextWithSpan(context.Background(), span)

	groupConfig, err := model.DecodeGroupConfig(ctx, req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	_, err = ts.store.AddConfigToGroup(ctx, id, ver, groupConfig)

	if err != nil {
		err := errors.New("key not found")
		tracer.LogError(span, err)
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	model.RenderJSON(ctx, w, id)

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

	model.RenderJSONOld(w, id)

	return id
}

func (ts *Service) delGroupHandler(w http.ResponseWriter, req *http.Request) {
	span := tracer.StartSpanFromRequest("delGroupHandler", ts.tracer, req)
	defer span.Finish()

	span.LogFields(
		tracer.LogString("handler", fmt.Sprintf("handling delete group at %s\n", req.URL.Path)),
	)

	id := mux.Vars(req)["uuid"]
	ver := mux.Vars(req)["ver"]

	ctx := tracer.ContextWithSpan(context.Background(), span)

	r, err := ts.store.DeleteGroup(ctx, id, ver)
	if err != nil {
		err := errors.New("key not found")
		tracer.LogError(span, err)
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	model.RenderJSON(ctx, w, r)
}
