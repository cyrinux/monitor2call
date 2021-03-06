package models

import (
	"time"

	"github.com/gregdel/pushover"
)

// AlertAcknowledgement define ack for Alert
type AlertAcknowledgement struct {
	UserID                 string    `json:"user_id"`
	User                   string    `json:"user"`
	Phone                  string    `json:"phone"`
	PushoverReceipt        string    `json:"pushover_receipt"`
	PushoverAcknowledged   bool      `json:"pushover_acknowledged"`
	PushoverAcknowledgedAt time.Time `json:"pushover_acknowledged_at"`
	SmsSent                bool      `json:"sms_sent"`
	SmsSentAt              time.Time `json:"sms_sent_at"`
	SmsAcknowledged        bool      `json:"sms_acknowledged"`
	SmsAcknowledgedAt      time.Time `json:"sms_acknowledged_at"`
}

// Translation structure
type Translation struct {
	Language string `json:"language" binding:"-"`
	Message  string `json:"message" binding:"-"`
	Voice    string `json:"voice" binding:"-"`
	VoiceURL string `json:"voice_url" binding:"-"`
}

// Alert define alert content
type Alert struct {
	ID                   string                           `json:"_id,omitempty" form:"id" binding:"-"`
	Rev                  string                           `json:"_rev,omitempty" form:"rev" binding:"-"`
	Host                 string                           `json:"host" form:"host" binding:"-" example:"dbprep01"`
	Service              string                           `json:"service" form:"service" binding:"-" example:"postgres"`
	Monitor              string                           `json:"monitor" form:"monitor" binding:"required" example:"nagios"`
	Message              string                           `json:"message" form:"message" binding:"-" example:"postgres on dbprep01 is down"`
	State                string                           `json:"state" form:"state" binding:"-" example:"down"`
	Filename             string                           `json:"filename" binding:"-" example:"./cache/nagios-postgres-dbprep01-down-fr.mp3"`
	Tags                 []string                         `json:"tags" binding:"-"`
	Acknowledgements     map[string]*AlertAcknowledgement `json:"acknowledgements,omitempty" binding:"-"`
	Translations         map[string]*Translation          `json:"translations,omitempty" binding:"-"`
	URL                  string                           `json:"url" binding:"-" example:"https://nagios.eg"`
	URLTitle             string                           `json:"url_title" binding:"-" example:"Nagios Alert"`
	Priority             int                              `json:"priority" binding:"-" example:"4"`
	CreatedAt            int64                            `json:"created_at" binding:"-"`
	UpdatedAt            int64                            `json:"updated_at" binding:"-"`
	PushoverAcknowledged bool                             `json:"pushover_acknowledged" binding:"-"`
	SmsAcknowledged      bool                             `json:"sms_acknowledged" binding:"-"`
	Canceled             bool                             `json:"canceled" binding:"-"`
}

// Validate the Alert input input
func (alert *Alert) Validate() string {
	if alert.Message == "" {
		if alert.Host == "" {
			return "Field 'host' cannot be empty"
		}
		if alert.Service == "" {
			return "Field 'service' cannot be empty"
		}
		if alert.State == "" {
			return "Field 'state' cannot be empty"
		}
	}

	if alert.Monitor == "" {
		return "Field 'monitor' cannot be empty"
	}

	return ""
}

// Prepare the pushover alert message
func (alert *Alert) Prepare() *pushover.Message {
	alert.Acknowledgements = make(map[string]*AlertAcknowledgement)
	alert.Translations = make(map[string]*Translation)

	// if Message is not set, generate on
	if alert.Message == "" {
		alert.Message = "Monitor: " + alert.Monitor + ". Alert " + alert.Service + " on " + alert.Host + " is " + alert.State + "."
	}

	// voice filename prefix
	if alert.Host != "" && alert.Monitor != "" && alert.Service != "" && alert.State != "" {
		alert.Filename = alert.Monitor + "-" + alert.Host + "-" + alert.Service + "-" + alert.State
	} else {
		alert.Filename = alert.Message
	}

	// Generate and add Tags
	// add state as tag
	// if no tag, broadcast on all users (all users must have "all" tag)
	if len(alert.Tags) == 0 {
		alert.Tags = append(alert.Tags, "all")
	}
	// add service, host, state, monitor in tags if filled
	autoTags := []string{alert.Host, alert.Service, alert.State, alert.Monitor}
	for idx := 0; idx < len(autoTags); idx++ {
		if autoTags[idx] != "" {
			alert.Tags = append(alert.Tags, autoTags[idx])
		}

	}

	// contruct pushover message
	message := &pushover.Message{
		Title:    alert.Message,
		Message:  alert.Message,
		Retry:    5 * time.Minute,
		Expire:   1 * time.Hour,
		Sound:    pushover.SoundCosmic,
		Priority: pushover.PriorityNormal,
		URL:      alert.URL,
		URLTitle: alert.URLTitle,
	}

	// adapt priority about some criterea
	switch {
	case alert.Monitor == "internetvista":
		message.Priority = pushover.PriorityEmergency
	case alert.Priority <= 6:
		message.Priority = alert.Priority
	case alert.State == "down" || alert.State == "critical":
		message.Priority = pushover.PriorityEmergency
	case alert.State == "recovery":
		message.Priority = pushover.PriorityHigh
	default:
		message.Priority = pushover.PriorityNormal
	}

	// Timestamps
	alert.CreatedAt = time.Now().Unix()
	alert.UpdatedAt = time.Now().Unix()

	return message
}
