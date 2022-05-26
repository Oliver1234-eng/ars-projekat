package poststore

import (
	model "ars-projekat/model"
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

func (ps *ConfigStore) IdempotencyKeyExists(key string) (bool, string, error) {
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

func (ps *ConfigStore) CreateConfig(configJSON *model.ConfigJSON) (string, error) {
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
	_, err = kv.Put(p, nil)
	if err != nil {
		return "", err
	}

	return rid, nil
}

func (ps *ConfigStore) CreateConfigVersion(id string, configJSON *model.ConfigJSON) (string, error) {
	kv := ps.cli.KV()

	confExists := ps.CheckIfConfigExists(id)
	if !confExists {
		return "", errors.New("Config not found")
	}

	confVersionExists := ps.CheckIfConfigVersionExists(id, configJSON.Version)
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
	_, err = kv.Put(p, nil)
	if err != nil {
		return "", err
	}

	return configKey, nil
}

func (ps *ConfigStore) CreateGroup(groupJSON *model.GroupJSON) (string, error) {
	kv := ps.cli.KV()

	groupId := createId()

	for _, c := range groupJSON.Configs {
		labels := model.DecodeJSONLabels(c.Labels)
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
		_, err = kv.Put(p, nil)
		if err != nil {
			return "", err
		}
	}

	return groupId, nil
}

func (ps *ConfigStore) GetConfig(id string, version string) (*model.Config, error) {
	kv := ps.cli.KV()

	configKey := constructConfigKey(id, version)

	pair, _, err := kv.Get(configKey, nil)
	if err != nil {
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

func (ps *ConfigStore) GetGroup(id string, version string, labels string) ([]*model.Config, error) {
	kv := ps.cli.KV()

	groupKey := constructGroupKey(id, version, labels)

	data, _, err := kv.List(groupKey, nil)
	if err != nil {
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

func (ps *ConfigStore) DeleteConfig(id string, version string) (map[string]string, error) {
	kv := ps.cli.KV()

	configKey := constructConfigKey(id, version)

	_, err := kv.Delete(configKey, nil)
	if err != nil {
		return nil, err
	}

	return map[string]string{"Deleted": id}, nil
}

func (ps *ConfigStore) AddConfigToGroup(id string, version string, groupConfigJSON *model.GroupConfigJSON) (string, error) {
	kv := ps.cli.KV()

	verExists := ps.CheckIfGroupVersionExists(id, version)
	if !verExists {
		return "", errors.New("Group not found")
	}

	labels := model.DecodeJSONLabels(groupConfigJSON.Labels)
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
	_, err = kv.Put(p, nil)
	if err != nil {
		return "", err
	}

	return groupConfigKey, nil

}

func (ps *ConfigStore) CheckIfConfigExists(id string) bool {
	kv := ps.cli.KV()

	groupKey := fmt.Sprintf("configs/%s/", id)

	data, _, err := kv.List(groupKey, nil)
	if err != nil {
		return false
	}

	if data == nil {
		return false
	}

	return true
}

func (ps *ConfigStore) CheckIfGroupVersionExists(id string, version string) bool {
	kv := ps.cli.KV()

	groupKey := fmt.Sprintf("groups/%s/%s/", id, version)

	data, _, err := kv.List(groupKey, nil)
	if err != nil {
		return false
	}

	if data == nil {
		return false
	}

	return true
}

func (ps *ConfigStore) CreateGroupVersion(groupId string, groupJSON *model.GroupJSON) (string, error) {
	kv := ps.cli.KV()

	groupExists := ps.CheckIfGroupExists(groupId)
	if !groupExists {
		return "", errors.New("Group not found")
	}

	groupVersionExists := ps.CheckIfGroupVersionExists(groupId, groupJSON.Version)
	if groupVersionExists {
		return "", errors.New("Group version already exists")
	}

	for _, c := range groupJSON.Configs {
		labels := model.DecodeJSONLabels(c.Labels)
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
		_, err = kv.Put(p, nil)
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

func (ps *ConfigStore) DeleteGroup(id string, version string) (map[string]string, error) {
	kv := ps.cli.KV()

	groupKey := constructGroupKey(id, version, "")

	_, err := kv.DeleteTree(groupKey, nil)
	if err != nil {
		return nil, err
	}

	return map[string]string{"Deleted": id}, nil
}

func (ps *ConfigStore) CheckIfConfigVersionExists(id string, version string) bool {
	kv := ps.cli.KV()

	groupKey := fmt.Sprintf("configs/%s/%s/", id, version)

	data, _, err := kv.List(groupKey, nil)
	if err != nil {
		return false
	}

	if data == nil {
		return false
	}

	return true
}
