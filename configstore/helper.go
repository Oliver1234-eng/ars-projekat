package poststore

import (
	tracer "ars-projekat/tracer"
	"context"
	"fmt"
	"github.com/google/uuid"
)

const (
	configs             = "configs/%s/%s/"
	groups              = "groups/%s/%s/%s/"
	groupsNoLabels      = "groups/%s/%s/"
	groupConfig         = "groups/%s/%s/%s/%s/"
	groupConfigNoLabels = "groups/%s/%s/%s/"
	idempotency         = "idempotency/%s/"
)

func createId() string {
	return uuid.New().String()
}

func generateConfigKey(ctx context.Context, version string) (string, string) {
	span := tracer.StartSpanFromContext(ctx, "generateConfigKey")
	defer span.Finish()
	id := uuid.New().String()
	return constructConfigKey(ctx, id, version), id
}

func generateGroupConfigKey(ctx context.Context, groupId string, version string, labels string) (string, string) {
	span := tracer.StartSpanFromContext(ctx, "generateGroupConfigKey")
	defer span.Finish()

	configId := uuid.New().String()
	return constructGroupConfigKey(ctx, groupId, configId, version, labels), configId
}

func constructConfigKey(ctx context.Context, id string, version string) string {
	span := tracer.StartSpanFromContext(ctx, "constructConfigKey")
	defer span.Finish()
	return fmt.Sprintf(configs, id, version)
}

func constructGroupKey(ctx context.Context, id string, version string, labels string) string {
	span := tracer.StartSpanFromContext(ctx, "constructGroupKey")
	defer span.Finish()
	if labels == "" {
		return fmt.Sprintf(groupsNoLabels, id, version)
	} else {
		return fmt.Sprintf(groups, id, version, labels)
	}
}

func constructGroupConfigKey(ctx context.Context, groupId string, configId string, version string, labels string) string {
	span := tracer.StartSpanFromContext(ctx, "constructGroupConfigKey")
	defer span.Finish()
	if labels == "" {
		return fmt.Sprintf(groupConfigNoLabels, groupId, version, configId)
	} else {
		return fmt.Sprintf(groupConfig, groupId, version, labels, configId)
	}
}

func generateIdempotencyKey(ctx context.Context) (string, string) {
	span := tracer.StartSpanFromContext(ctx, "generateIdempotencyKey")
	defer span.Finish()
	id := uuid.New().String()
	return constructIdempotencyKey(ctx, id), id
}

func constructIdempotencyKey(ctx context.Context, key string) string {
	span := tracer.StartSpanFromContext(ctx, "constructIdempotencyKey")
	defer span.Finish()
	return fmt.Sprintf(idempotency, key)
}
