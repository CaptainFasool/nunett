package models

type DeploymentRequest struct {
	AddressUser string `json:"address_user"`
	MaxNtx      string `json:"max_ntx"`
	Blockchain  string `json:"blockchain"`
	ServiceType string `json:"service_type"`
	Params      struct {
		ImageID  string `json:"image_id"`
		ModelURL string `json:"model_url"`
		Packages string `json:"packages"`
	} `json:"params"`
	Constraints struct {
		CPU   string `json:"cpu"`
		RAM   string `json:"ram"`
		Vram  string `json:"vram"`
		Power string `json:"power"`
		Time  string `json:"time"`
	} `json:"constraints"`
}
