package notifications

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/cyrinux/monitor2call/models" //structs
	"github.com/cyrinux/monitor2call/helpers"      //structs

	"github.com/gregdel/pushover" // Stable version of Kivik
)

// InitPushover initialize pushover app
func InitPushover() *pushover.Pushover {
	// Create a new pushover app with a token
	appapikey := helpers.GetEnv("PUSHOVER_APP_API_KEY", "")
	if appapikey == "" {
		log.Fatal("Pushover api key env var PUSHOVER_APP_API_KEY is undefined")
		os.Exit(1)
	}
	app := pushover.New(appapikey)

	return app
}

// CheckPushoverUserKey check if the user api key is valid
func CheckPushoverUserKey(apikey string) error {
	app := InitPushover()
	// Create a new recipient
	recipient := pushover.NewRecipient(apikey)
	_, err := app.GetRecipientDetails(recipient)
	if err != nil {
		log.Error(err)
	}

	return err
}

// ToPushover send alert throught pushover
func ToPushover(apikey string, msg *pushover.Message) (string, error) {
	app := InitPushover()

	// Create a new recipient
	recipient := pushover.NewRecipient(apikey)

	response, err := app.SendMessage(msg, recipient)
	if err != nil {
		log.Errorf("Can't send alert with error: %s", err.Error())
	}

	return response.Receipt, err
}

// SendNotification is a meta method which send alert throught
// all channel (pushover, SMS, Call ?)
// TO be enhance!
func SendNotification(msg *pushover.Message, user *models.User) (receipt string, err error) {
	receipt, err = ToPushover(user.PushoverAPIKey, msg)

	// send pushover message
	log.Printf("Message %s sent to user %s with pushover apiKey: %s", msg.Message, user.Name, user.PushoverAPIKey)

	return receipt, err
}

// =====================================
// Pushover message params
// message := &pushover.Message{
//     Message:     "My awesome message",
//     Title:       "My title",
//     Priority:    pushover.PriorityEmergency,
//     URL:         "http://google.com",
//     URLTitle:    "Google",
//     Timestamp:   time.Now().Unix(),
//     Retry:       60 * time.Second,
//     Expire:      time.Hour,
//     DeviceName:  "SuperDevice",
//     CallbackURL: "http://yourapp.com/callback",
//     Sound:       pushover.SoundCosmic,
// }
