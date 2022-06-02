package model

type LabelJSON struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ConfigJSON struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Version string `json:"version"`
}

type GroupConfigJSON struct {
	Key    string      `json:"key"`
	Value  string      `json:"value"`
	Labels []LabelJSON `json:"labels"`
}

type GroupJSON struct {
	Configs []GroupConfigJSON `json:"configs"`
	Version string            `json:"version"`
}
