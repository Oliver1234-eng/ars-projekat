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

func (ts *Service) IdempotencyCheck(handlerFunc func(context.Context, http.ResponseWriter, *http.Request) string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		span := tracer.StartSpanFromRequest("IdempotencyCheck", ts.tracer, req)
		defer span.Finish()

		ctx := tracer.ContextWithSpan(context.Background(), span)

		idempotencyKey := req.Header.Get("Idempotency-Key")
		if idempotencyKey == "" {
			http.Error(w, "Missing Idempotency-Key header", http.StatusBadRequest)
			return
		}

		keyExists, storedKey, err := ts.store.IdempotencyKeyExists(ctx, idempotencyKey)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if keyExists {
			model.RenderJSON(ctx, w, storedKey)
			return
		}

		id := handlerFunc(ctx, w, req)
		if id != "" {
			ts.store.SaveIdempotencyKey(ctx, idempotencyKey, id)
		}
	}
}

func (ts *Service) createConfigHandler(ctx context.Context, w http.ResponseWriter, req *http.Request) string {
	span := tracer.StartSpanFromContext(ctx, "createConfigHandler")
	defer span.Finish()
	span.LogFields(
		tracer.LogString("handler", fmt.Sprintf("handling create config at %s\n", req.URL.Path)),
	)
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

	rt, err := model.DecodeConfig(ctx, req.Body)
	if err != nil {
		tracer.LogError(span, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	id, err := ts.store.CreateConfig(ctx, rt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		tracer.LogError(span, err)
		return ""
	}

	model.RenderJSON(ctx, w, id)

	return id

}

func (ts *Service) createConfigVersionHandler(ctx context.Context, w http.ResponseWriter, req *http.Request) string {
	span := tracer.StartSpanFromContext(ctx, "createConfigVersionHandler")
	defer span.Finish()

	span.LogFields(
		tracer.LogString("handler", fmt.Sprintf("handling create config version at %s\n", req.URL.Path)),
	)
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

	rt, err := model.DecodeConfig(ctx, req.Body)
	if err != nil {
		tracer.LogError(span, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	_, err = ts.store.CreateConfigVersion(ctx, id, rt)
	if err != nil {
		tracer.LogError(span, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}

	model.RenderJSON(ctx, w, id)

	return id
}

func (ts *Service) createGroupHandler(ctx context.Context, w http.ResponseWriter, req *http.Request) string {
	span := tracer.StartSpanFromContext(ctx, "createGroupHandler")
	defer span.Finish()
	span.LogFields(
		tracer.LogString("handler", fmt.Sprintf("handling create group at %s\n", req.URL.Path)),
	)
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

	rt, err := model.DecodeGroup(ctx, req.Body)
	if err != nil {
		tracer.LogError(span, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	id, err := ts.store.CreateGroup(ctx, rt)
	if err != nil {
		tracer.LogError(span, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}

	model.RenderJSON(ctx, w, id)

	return id
}

func (ts *Service) getConfigHandler(w http.ResponseWriter, req *http.Request) {
	span := tracer.StartSpanFromRequest("getConfigHandler", ts.tracer, req)
	defer span.Finish()

	span.LogFields(tracer.LogString("handler", fmt.Sprintf("handling get config from %s\n", req.URL.Path)))

	id := mux.Vars(req)["uuid"]
	ver := mux.Vars(req)["ver"]

	ctx := tracer.ContextWithSpan(context.Background(), span)

	config, err := ts.store.GetConfig(ctx, id, ver)
	if err != nil {
		err := errors.New("key not found")
		tracer.LogError(span, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	model.RenderJSON(ctx, w, config)
}

func (ts *Service) getGroupHandler(w http.ResponseWriter, req *http.Request) {
	span := tracer.StartSpanFromRequest("getGroupHandler", ts.tracer, req)
	defer span.Finish()

	span.LogFields(tracer.LogString("handler", fmt.Sprintf("handling get group from %s\n", req.URL.Path)))

	id := mux.Vars(req)["uuid"]
	ver := mux.Vars(req)["ver"]
	labels := model.DecodeQueryLabels(req.URL.Query())
	ctx := tracer.ContextWithSpan(context.Background(), span)

	group, err := ts.store.GetGroup(ctx, id, ver, labels)
	if err != nil {
		tracer.LogError(span, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model.RenderJSON(ctx, w, group)
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

func (ts *Service) addConfigToGroupHandler(ctx context.Context, w http.ResponseWriter, req *http.Request) string {
	span := tracer.StartSpanFromContext(ctx, "addConfigToGroupHandler")
	defer span.Finish()

	span.LogFields(
		tracer.LogString("handler", fmt.Sprintf("handling add config to group at %s\n", req.URL.Path)),
	)

	id := mux.Vars(req)["uuid"]
	ver := mux.Vars(req)["ver"]

	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		tracer.LogError(span, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	if mediatype != "application/json" {
		err := errors.New("Expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return ""
	}

	groupConfig, err := model.DecodeGroupConfig(ctx, req.Body)
	if err != nil {
		tracer.LogError(span, err)
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

func (ts *Service) createGroupVersionHandler(ctx context.Context, w http.ResponseWriter, req *http.Request) string {
	span := tracer.StartSpanFromContext(ctx, "createGroupVersionHandler")
	defer span.Finish()

	span.LogFields(
		tracer.LogString("handler", fmt.Sprintf("handling create group version at %s\n", req.URL.Path)),
	)

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

	rt, err := model.DecodeGroup(ctx, req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return ""
	}

	_, err = ts.store.CreateGroupVersion(ctx, id, rt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}

	model.RenderJSON(ctx, w, id)

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
