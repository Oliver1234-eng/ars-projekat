package model

import (
	tracer "ars-projekat/tracer"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"sort"
	"strings"
)

func DecodeConfig(r io.Reader) (*ConfigJSON, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt ConfigJSON
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func DecodeGroupConfig(ctx context.Context, r io.Reader) (*GroupConfigJSON, error) {
	span := tracer.StartSpanFromContext(ctx, "DecodeGroupConfig")
	defer span.Finish()

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt GroupConfigJSON
	if err := dec.Decode(&rt); err != nil {
		tracer.LogError(span, err)
		return nil, err
	}
	return &rt, nil
}

func DecodeGroup(r io.Reader) (*GroupJSON, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt GroupJSON
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func DecodeQueryLabels(labelsMap map[string][]string) string {
	keys := make([]string, 0, len(labelsMap))
	pairs := make([]string, 0, len(labelsMap))

	for k := range labelsMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		val := fmt.Sprintf("%s=%s", k, labelsMap[k][0])
		pairs = append(pairs, val)
	}

	return strings.Join(pairs[:], "&")
}

func DecodeJSONLabels(labels []LabelJSON) string {
	keys := make([]string, 0, len(labels))
	pairs := make([]string, 0, len(labels))

	labelsMap := make(map[string]string)

	for _, l := range labels {
		keys = append(keys, l.Key)
		labelsMap[l.Key] = l.Value
	}

	sort.Strings(keys)

	for _, k := range keys {
		val := fmt.Sprintf("%s=%s", k, labelsMap[k])
		pairs = append(pairs, val)
	}

	return strings.Join(pairs[:], "&")
}

func RenderJSONOld(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/model")
	w.Write(js)
}

func RenderJSON(ctx context.Context, w http.ResponseWriter, v interface{}) {
	span := tracer.StartSpanFromContext(ctx, "RenderJSON")
	defer span.Finish()

	js, err := json.Marshal(v)
	if err != nil {
		tracer.LogError(span, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/model")
	w.Write(js)
}

func CreateId() string {
	return uuid.New().String()
}
