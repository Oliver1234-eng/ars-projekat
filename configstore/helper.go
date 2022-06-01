package poststore

import (
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

func generateConfigKey(version string) (string, string) {
	id := uuid.New().String()
	return constructConfigKey(id, version), id
}

func generateGroupConfigKey(groupId string, version string, labels string) (string, string) {
	configId := uuid.New().String()
	return constructGroupConfigKey(groupId, configId, version, labels), configId
}

func constructConfigKey(id string, version string) string {
	return fmt.Sprintf(configs, id, version)
}

func constructGroupKey(id string, version string, labels string) string {
	if labels == "" {
		return fmt.Sprintf(groupsNoLabels, id, version)
	} else {
		return fmt.Sprintf(groups, id, version, labels)
	}
}

func constructGroupConfigKey(groupId string, configId string, version string, labels string) string {
	if labels == "" {
		return fmt.Sprintf(groupConfigNoLabels, groupId, version, configId)
	} else {
		return fmt.Sprintf(groupConfig, groupId, version, labels, configId)
	}
}

func generateIdempotencyKey() (string, string) {
	id := uuid.New().String()
	return constructIdempotencyKey(id), id
}

func constructIdempotencyKey(key string) string {
	return fmt.Sprintf(idempotency, key)
}
