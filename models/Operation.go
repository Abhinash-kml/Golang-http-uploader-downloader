package models

type OperationRequest struct {
	Operation string `json:"operation"`
	Value     string `json:"value"`
}
