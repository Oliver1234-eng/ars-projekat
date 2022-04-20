package main

type Config struct {
	Key   string `json:"key"`
	Kalue string `json:"value"`
}

type Group struct {
	Configs []Config `json:"configs"`
}
