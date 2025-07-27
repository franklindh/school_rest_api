package models

type Student struct {
	ID        int `json:"id,omitempty"`
	FirstName int `json:"first_name,omitempty"`
	LastName  int `json:"last_name,omitempty"`
	Email     int `json:"email,omitempty"`
	Class     int `json:"class,omitempty"`
}
