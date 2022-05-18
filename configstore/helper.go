package poststore

import (
	"fmt"
	"github.com/google/uuid"
)

const (
	configs        = "configs/%s/%s/"
	groups         = "groups/%s/%s/%s/"
	groupsNoLabels = "groups/%s/%s/"
)

func createId() string {
	return uuid.New().String()
}

func generateConfigKey(version string) (string, string) {
	id := uuid.New().String()
	return constructConfigKey(id, version), id
}

func generateGroupKey(version string, labels string) (string, string) {
	id := uuid.New().String()
	return constructGroupKey(id, version, labels), id
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
