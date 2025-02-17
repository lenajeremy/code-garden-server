package models

type Environment struct {
	ContainerId string `json:"containerId"`
	Status      string `json:"status"`
}
