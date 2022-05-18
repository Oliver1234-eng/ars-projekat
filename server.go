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
	configs map[string]*model.ConfigJSON
	groups  map[string]*model.GroupJSON
	store   *poststore.ConfigStore
}

func (ts *Service) createConfigHandler(w http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mediatype != "application/json" {
		err := errors.New("Expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	rt, err := model.DecodeConfig(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := ts.store.CreateConfig(rt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model.RenderJSON(w, id)
}

func (ts *Service) createGroupHandler(w http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mediatype != "application/json" {
		err := errors.New("Expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	rt, err := model.DecodeGroup(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := ts.store.CreateGroup(rt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model.RenderJSON(w, id)
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

//func (ts *Service) delConfigHandler(w http.ResponseWriter, req *http.Request) {
//	id := mux.Vars(req)["uuid"]
//	if v, ok := ts.configs[id]; ok {
//		delete(ts.configs, id)
//		model.renderJSON(w, v)
//	} else {
//		err := errors.New("key not found")
//		http.Error(w, err.Error(), http.StatusNotFound)
//	}
//}

//func (ts *Service) delGroupHandler(w http.ResponseWriter, req *http.Request) {
//	id := mux.Vars(req)["uuid"]
//	if v, ok := ts.groups[id]; ok {
//		delete(ts.groups, id)
//		model.renderJSON(w, v)
//	} else {
//		err := errors.New("key not found")
//		http.Error(w, err.Error(), http.StatusNotFound)
//	}
//}
//
//func (ts *Service) addConfigToGroupHandler(w http.ResponseWriter, req *http.Request) {
//	id := mux.Vars(req)["uuid"]
//
//	contentType := req.Header.Get("Content-Type")
//	mediatype, _, err := mime.ParseMediaType(contentType)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	if mediatype != "application/json" {
//		err := errors.New("Expect application/json Content-Type")
//		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
//		return
//	}
//
//	rt, err := model.decodeGroupConfig(req.Body)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	if v, ok := ts.groups[id]; ok {
//		v.Configs = append(v.Configs, *rt)
//		model.renderJSON(w, v)
//	} else {
//		err := errors.New("key not found")
//		http.Error(w, err.Error(), http.StatusNotFound)
//	}
//}
