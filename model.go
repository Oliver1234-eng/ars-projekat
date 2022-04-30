package main

type Config struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Group struct {
	Configs []Config `json:"configs"`
}
