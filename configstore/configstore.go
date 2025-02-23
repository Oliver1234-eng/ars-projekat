package poststore

import (
	model "ars-projekat/model"
	tracer "ars-projekat/tracer"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"os"
)

type ConfigStore struct {
	cli *api.Client
}

func New() (*ConfigStore, error) {
	db := os.Getenv("DB")
	dbport := os.Getenv("DBPORT")

	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%s", db, dbport)
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &ConfigStore{
		cli: client,
	}, nil
}

func (ps *ConfigStore) IdempotencyKeyExists(ctx context.Context, key string) (bool, string, error) {
	span := tracer.StartSpanFromContext(ctx, "IdempotencyKeyExists")
	defer span.Finish()

	kv := ps.cli.KV()

	idempotencyKey := fmt.Sprintf("idempotency/%s/", key)

	uuid, _, err := kv.Get(idempotencyKey, nil)
	if err != nil {
		return false, "", err
	}

	if uuid == nil {
		return false, "", nil
	}

	return true, string(uuid.Value), nil
}

func (ps *ConfigStore) CreateConfig(ctx context.Context, configJSON *model.ConfigJSON) (string, error) {
	span := tracer.StartSpanFromContext(ctx, "CreateConfig")
	defer span.Finish()

	kv := ps.cli.KV()

	sid, rid := generateConfigKey(configJSON.Version)
	config := model.Config{
		Key:   configJSON.Key,
		Value: configJSON.Value,
	}

	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	p := &api.KVPair{Key: sid, Value: data}

	putSpan := tracer.StartSpanFromContext(ctx, "Put")
	_, err = kv.Put(p, nil)
	putSpan.Finish()

	if err != nil {
		return "", err
	}

	return rid, nil
}

func (ps *ConfigStore) CreateConfigVersion(ctx context.Context, id string, configJSON *model.ConfigJSON) (string, error) {
	span := tracer.StartSpanFromContext(ctx, "CreateConfigVersion")
	defer span.Finish()
	kv := ps.cli.KV()

	confExists := ps.CheckIfConfigExists(ctx, id)
	if !confExists {
		return "", errors.New("Config not found")
	}

	confVersionExists := ps.CheckIfConfigVersionExists(ctx, id, configJSON.Version)
	if confVersionExists {
		return "", errors.New("Config version already exists")
	}

	configKey := constructConfigKey(id, configJSON.Version)

	config := model.Config{
		Key:   configJSON.Key,
		Value: configJSON.Value,
	}

	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	p := &api.KVPair{Key: configKey, Value: data}

	putSpan := tracer.StartSpanFromContext(ctx, "Put")
	_, err = kv.Put(p, nil)
	putSpan.Finish()

	if err != nil {
		return "", err
	}

	return configKey, nil
}

func (ps *ConfigStore) CreateGroup(ctx context.Context, groupJSON *model.GroupJSON) (string, error) {
	span := tracer.StartSpanFromContext(ctx, "CreateGroup")
	defer span.Finish()

	kv := ps.cli.KV()

	groupId := createId()

	for _, c := range groupJSON.Configs {
		labels := model.DecodeJSONLabels(ctx, c.Labels)
		groupConfigKey, _ := generateGroupConfigKey(groupId, groupJSON.Version, labels)

		config := model.Config{
			Key:   c.Key,
			Value: c.Value,
		}

		data, err := json.Marshal(config)
		if err != nil {
			return "", err
		}

		p := &api.KVPair{Key: groupConfigKey, Value: data}

		putSpan := tracer.StartSpanFromContext(ctx, "Put")
		_, err = kv.Put(p, nil)
		putSpan.Finish()

		if err != nil {
			return "", err
		}
	}

	return groupId, nil
}

func (ps *ConfigStore) GetConfig(ctx context.Context, id string, version string) (*model.Config, error) {
	span := tracer.StartSpanFromContext(ctx, "GetConfig")
	defer span.Finish()

	kv := ps.cli.KV()

	configKey := constructConfigKey(id, version)

	getSpan := tracer.StartSpanFromContext(ctx, "Get")
	pair, _, err := kv.Get(configKey, nil)
	getSpan.Finish()

	if err != nil {
		tracer.LogError(span, err)
		return nil, err
	}

	if pair == nil {
		return nil, errors.New("Config not found")
	}

	post := &model.Config{}
	err = json.Unmarshal(pair.Value, post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (ps *ConfigStore) GetGroup(ctx context.Context, id string, version string, labels string) ([]*model.Config, error) {
	span := tracer.StartSpanFromContext(ctx, "GetGroup")
	defer span.Finish()

	kv := ps.cli.KV()

	groupKey := constructGroupKey(id, version, labels)

	listSpan := tracer.StartSpanFromContext(ctx, "List")
	data, _, err := kv.List(groupKey, nil)
	listSpan.Finish()

	if err != nil {
		tracer.LogError(span, err)
		return nil, err
	}

	if data == nil {
		return nil, errors.New("Group not found")
	}

	groupConfigs := []*model.Config{}
	for _, pair := range data {
		config := &model.Config{}
		err = json.Unmarshal(pair.Value, config)
		if err != nil {
			return nil, err
		}
		groupConfigs = append(groupConfigs, config)
	}

	return groupConfigs, nil
}

func (ps *ConfigStore) DeleteConfig(ctx context.Context, id string, version string) (map[string]string, error) {
	span := tracer.StartSpanFromContext(ctx, "DeleteConfig")
	defer span.Finish()

	kv := ps.cli.KV()

	configKey := constructConfigKey(id, version)

	deleteSpan := tracer.StartSpanFromContext(ctx, "Delete")
	_, err := kv.Delete(configKey, nil)
	deleteSpan.Finish()

	if err != nil {
		tracer.LogError(span, err)
		return nil, err
	}

	return map[string]string{"Deleted": id}, nil
}

func (ps *ConfigStore) AddConfigToGroup(ctx context.Context, id string, version string, groupConfigJSON *model.GroupConfigJSON) (string, error) {
	span := tracer.StartSpanFromContext(ctx, "AddConfigToGroup")
	defer span.Finish()

	kv := ps.cli.KV()

	verExists := ps.CheckIfGroupVersionExists(ctx, id, version)
	if !verExists {
		return "", errors.New("Group not found")
	}

	labels := model.DecodeJSONLabels(ctx, groupConfigJSON.Labels)
	groupConfigKey := constructGroupKey(id, version, labels)

	config := model.Config{
		Key:   groupConfigJSON.Key,
		Value: groupConfigJSON.Value,
	}

	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	p := &api.KVPair{Key: groupConfigKey, Value: data}

	putSpan := tracer.StartSpanFromContext(ctx, "Put")
	_, err = kv.Put(p, nil)
	putSpan.Finish()

	if err != nil {
		return "", err
	}

	return groupConfigKey, nil

}

func (ps *ConfigStore) CheckIfConfigExists(ctx context.Context, id string) bool {
	span := tracer.StartSpanFromContext(ctx, "CheckIfConfigExists")
	defer span.Finish()

	kv := ps.cli.KV()

	groupKey := fmt.Sprintf("configs/%s/", id)

	listSpan := tracer.StartSpanFromContext(ctx, "List")
	data, _, err := kv.List(groupKey, nil)
	listSpan.Finish()

	if err != nil {
		return false
	}

	if data == nil {
		return false
	}

	return true
}

func (ps *ConfigStore) CheckIfGroupVersionExists(ctx context.Context, id string, version string) bool {
	span := tracer.StartSpanFromContext(ctx, "CheckIfGroupVersionExists")
	defer span.Finish()

	kv := ps.cli.KV()

	groupKey := fmt.Sprintf("groups/%s/%s/", id, version)

	listSpan := tracer.StartSpanFromContext(ctx, "List")
	data, _, err := kv.List(groupKey, nil)
	listSpan.Finish()

	if err != nil {
		tracer.LogError(span, err)
		return false
	}

	if data == nil {
		return false
	}

	return true
}

func (ps *ConfigStore) CreateGroupVersion(ctx context.Context, groupId string, groupJSON *model.GroupJSON) (string, error) {
	span := tracer.StartSpanFromContext(ctx, "CreateGroupVersion")
	defer span.Finish()

	kv := ps.cli.KV()

	groupExists := ps.CheckIfGroupExists(groupId)
	if !groupExists {
		return "", errors.New("Group not found")
	}

	groupVersionExists := ps.CheckIfGroupVersionExists(ctx, groupId, groupJSON.Version)
	if groupVersionExists {
		return "", errors.New("Group version already exists")
	}

	for _, c := range groupJSON.Configs {
		labels := model.DecodeJSONLabels(ctx, c.Labels)
		groupConfigKey, _ := generateGroupConfigKey(groupId, groupJSON.Version, labels)

		config := model.Config{
			Key:   c.Key,
			Value: c.Value,
		}

		data, err := json.Marshal(config)
		if err != nil {
			return "", err
		}

		p := &api.KVPair{Key: groupConfigKey, Value: data}

		putSpan := tracer.StartSpanFromContext(ctx, "Put")
		_, err = kv.Put(p, nil)
		putSpan.Finish()

		if err != nil {
			return "", err
		}
	}

	return groupId, nil
}

func (ps *ConfigStore) CheckIfGroupExists(id string) bool {
	kv := ps.cli.KV()

	groupKey := fmt.Sprintf("groups/%s/", id)

	data, _, err := kv.List(groupKey, nil)
	if err != nil {
		return false
	}

	if data == nil {
		return false
	}

	return true
}

func (ps *ConfigStore) DeleteGroup(ctx context.Context, id string, version string) (map[string]string, error) {
	span := tracer.StartSpanFromContext(ctx, "DeleteGroup")
	defer span.Finish()

	kv := ps.cli.KV()

	groupKey := constructGroupKey(id, version, "")

	deleteTreeSpan := tracer.StartSpanFromContext(ctx, "DeleteTree")
	_, err := kv.DeleteTree(groupKey, nil)
	deleteTreeSpan.Finish()

	if err != nil {
		tracer.LogError(span, err)
		return nil, err
	}

	return map[string]string{"Deleted": id}, nil
}

func (ps *ConfigStore) CheckIfConfigVersionExists(ctx context.Context, id string, version string) bool {
	span := tracer.StartSpanFromContext(ctx, "CheckConfigVersion")
	defer span.Finish()

	kv := ps.cli.KV()

	groupKey := fmt.Sprintf("configs/%s/%s/", id, version)

	listSpan := tracer.StartSpanFromContext(ctx, "List")
	data, _, err := kv.List(groupKey, nil)
	listSpan.Finish()

	if err != nil {
		tracer.LogError(span, err)
		return false
	}

	if data == nil {
		tracer.LogError(span, err)
		return false
	}

	return true
}

func (ps *ConfigStore) SaveIdempotencyKey(ctx context.Context, key string, itemId string) {
	span := tracer.StartSpanFromContext(ctx, "SaveIdempotencyKey")
	defer span.Finish()

	kv := ps.cli.KV()

	idempotencyKey := constructIdempotencyKey(key)

	p := &api.KVPair{Key: idempotencyKey, Value: []byte(itemId)}
	kv.Put(p, nil)
}
