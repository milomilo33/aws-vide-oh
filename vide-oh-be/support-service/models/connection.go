package models

type Connection struct {
	ID           string `json:"id" dynamodbav:"id"`
	ConnectionID string `json:"connectionId" dynamodbav:"connectionId"`
	UserEmail    string `json:"userEmail" dynamodbav:"userEmail"`
}
