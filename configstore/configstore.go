package poststore

import (
	model "ars-projekat/model"
	"encoding/json"
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

func (ps *ConfigStore) CreateGroup(groupJSON *model.GroupJSON) (string, error) {
	kv := ps.cli.KV()

	uuid := createId()

	for _, c := range groupJSON.Configs {
		labels := model.DecodeJSONLabels(c.Labels)
		groupConfigKey := constructGroupKey(uuid, groupJSON.Version, labels)

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

	return uuid, nil
}

func (ps *ConfigStore) GetConfig(id string, version string) (*model.Config, error) {
	kv := ps.cli.KV()

	configKey := constructConfigKey(id, version)

	pair, _, err := kv.Get(configKey, nil)
	if err != nil {
		return nil, err
	}

	post := &model.Config{}
	err = json.Unmarshal(pair.Value, post)
	if err != nil {
		return nil, err
	}

	return post, nil
}
