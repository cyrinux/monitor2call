package models

import (
	"regexp"
	"time"

	"github.com/nyaruka/phonenumbers"
)

// User define an user...
type User struct {
	ID             string   `json:"_id,omitempty"`
	Rev            string   `json:"_rev,omitempty"`
	Name           string   `json:"name" binding:"required" example:"John Doe"`
	Phone          string   `json:"phone" binding:"required" example:"+3310203040"`
	Email          string   `json:"email" binding:"required" example:"john@example.org"`
	Language       string   `json:"language" binding:"required" example:"en"`
	PushoverAPIKey string   `json:"pushover_api_key" binding:"required" example:"DZAFZefzeFZAOFIFfZEf"`
	Tags           []string `json:"tags" example:"tags" type:"array"`
	CreatedAt      int64    `json:"created_at" binding:"-"`
	UpdatedAt      int64    `json:"updated_at" binding:"-"`
}

// Validate user input
func (user *User) Validate() string {

	if user.Name == "" {
		return "Field 'name' cannot be empty"
	}

	if user.Phone == "" {
		return "Field 'phone' cannot be empty"
	}
	number, _ := phonenumbers.Parse(user.Phone, "FR")
	if phonenumbers.IsValidNumberForRegion(number, "FR") == false {
		return "Field 'phone' contains a wrong phone number"
	}

	if user.PushoverAPIKey == "" {
		return "Field 'pushover_api_key' cannot be empty"
	}

	if user.Email == "" {
		return "Field 'email' cannot be empty"
	}
	var reMail = regexp.MustCompile("^.*@.*$")
	if reMail.MatchString(user.Email) == false {
		return "Field 'email' not contains a valid email format"
	}

	if user.Language == "" {
		return "Field 'language' cannot be empty"
	}

	return ""
}

// Prepare user resource
func (user *User) Prepare() User {
	// Phone number
	number, _ := phonenumbers.Parse(user.Phone, "FR")
	formattedNumber := phonenumbers.Format(number, phonenumbers.E164)
	user.Phone = formattedNumber

	// Timestamps
	user.UpdatedAt = time.Now().Unix()
	user.CreatedAt = time.Now().Unix()

	return *user
}
