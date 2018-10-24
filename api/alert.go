package api

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/cyrinux/monitor2call/database" // common database lib
	"github.com/cyrinux/monitor2call/helpers"
	"github.com/cyrinux/monitor2call/models"        //structs
	"github.com/cyrinux/monitor2call/notifications" //notifications
	"github.com/cyrinux/monitor2call/translate"     //translate
	"github.com/gin-gonic/gin"
)

// PostAlert godoc
// @Summary Create an alert
// @Description create an alert
// @Accept application/json
// @Accept multipart/form-data
// @Accept json
// @Param host query string true "host"
// @Param service query string true "service"
// @Param monitor query string true "monitor"
// @Param state query string true "state"
// @Param tags query array true "tag1,tag2"
// @Produce json
// @Success 201 {object} models.Alert
// @Failure 503 {string} string
// @Failure 422 {string} string
// @Failure 400 {string} string
// @Router /api/v1/alerts [post]
func PostAlert(c *gin.Context) {

	var alert models.Alert

	publicURL := helpers.GetEnv("PUBLIC_URL", "http://localhost:3000")
	publicURL = strings.TrimRight(publicURL, "/")
	// // always return 400 if used BUG!: https://github.com/gin-gonic/gin/pull/1047
	// if err := c.ShouldBindJSON(&alert); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	// try form bind
	errBind := c.Bind(&alert)
	if errBind != nil {
		// or bind Json
		c.BindJSON(&alert)
	}

	if err := alert.Validate(); err != "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err})
		return
	}

	// prepare the message
	msg := alert.Prepare()

	// Init DB
	db := database.InitDb()

	// Fetch all users
	rows, err := db.Query(context.TODO(), "_design/users", "_view/all")
	if err != nil {
		log.Print(err)
	}

USER_LOOP:
	for rows.Next() {
		var user *models.User
		// fetch user
		id := rows.ID()
		row, err := db.Get(context.TODO(), id)
		if err != nil {
			log.Printf("Can't get user %s", id)
		}

		// read user data
		if err := row.ScanDoc(&user); err != nil {
			log.Print(err)
			continue USER_LOOP
		}

		// log loop over tags
		for _, alertTag := range alert.Tags {
			for _, userTag := range user.Tags {
				// if alert and user have a common tag then
				if alertTag == userTag {
					log.Printf(
						"Ding! user %s (lang: %s) has tag %s which match alert %s tag %s, sending alert on pushover %s then phone %s",
						user.Name, user.Language, userTag, alert.ID, alertTag, user.PushoverAPIKey, user.Phone,
					)

					// make translation
					// generate voice mp3
					var translation string
					var errTranslate error
					var voice string
					// var voiceURL string
					var errDoTTS error
					if _, ok := alert.Translations[user.Language]; !ok {
						translation, errTranslate = translate.WithGoogle(alert.Message, user.Language)
						voice, errDoTTS = translate.DoTTS(&models.TTS{
							Text:         translation,
							LanguageCode: user.Language,
							Filename:     alert.Filename,
						})
					} else {
						translation = alert.Translations[user.Language].Message
						voice = alert.Translations[user.Language].Voice
					}

					if errTranslate == nil && errDoTTS == nil {
						voiceURL := publicURL + "/statics/" + voice
						// create Message Translation
						alert.Translations[user.Language] = &models.Translation{
							Language: user.Language,
							Message:  translation,
							Voice:    voice,
							VoiceURL: voiceURL,
						}

						// set title with traduction
						if title := alert.Translations[user.Language].Message; title != "" {
							msg.Title = title
						}
					}

					// send notifications and keep receipt foreach user
					receipt, err := notifications.SendNotification(msg, user)
					alert.Acknowledgements[user.ID] = &models.AlertAcknowledgement{
						UserID:          user.ID,
						User:            user.Name,
						Phone:           user.Phone,
						PushoverReceipt: receipt,
					}

					if err != nil {
						log.Printf("Alert can't be sent: %s", err)
					}

					continue USER_LOOP
				}
			}
		}
	}

	if rows.Err() != nil {
		log.Print(rows.Err())
	}

	// if alert ID number set, edit the alert
	if alert.ID != "" {
		rev, err := db.Put(context.TODO(), alert.ID, &alert)
		if err != nil {
			c.JSON(http.StatusNotModified, gin.H{"error": alert})
			return
		}
		alert.Rev = rev
		c.JSON(http.StatusAccepted, gin.H{"success": alert})
	} else {
		// write alert to database and return _id and _rev
		docID, rev, err := db.CreateDoc(context.TODO(), &alert)
		if err != nil {
			log.Print(err.Error())
		}
		alert.ID = docID
		alert.Rev = rev

		c.JSON(http.StatusCreated, gin.H{"success": alert})
	}
	return
}

// GetAlerts godoc
// @Summary Get all alerts
// @Description Get all monitor alerts
// @Accept  multipart/form-data
// @Accept json
// @Produce json
// @Success 200 {object} models.Alert[]
// @Failure 404 {string} string
// @Router /api/v1/alerts [get]
func GetAlerts(c *gin.Context) {
	db := database.InitDb()

	var alert models.Alert
	var alerts []models.Alert

	rows, err := db.AllDocs(context.TODO())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No alerts found"})
		return
	}

	for rows.Next() {
		id := rows.ID()
		row, err := db.Get(context.TODO(), id)
		if err != nil {
			log.Printf("Can't get doc %s", id)
		}
		if err := row.ScanDoc(&alert); err != nil {
			log.Print(err)
		} else {
			alerts = append(alerts, alert)
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": &alerts})
	return
}

// GetAlert godoc
// @Summary Get alert by id
// @Description get struct array by ID
// @Param id path int true "4"
// @Success 200 {object} models.Alert
// @Success 404 {string} string
// @Router /api/v1/alerts/{id} [get]
func GetAlert(c *gin.Context) {
	db := database.InitDb()

	var alert models.Alert
	c.Bind(&alert)

	id := c.Params.ByName("id")

	row, err := db.Get(context.TODO(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	if err = row.ScanDoc(&alert); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Can't read alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": &alert})
	return
}

// DeleteAlert godoc
// @Summary Delete an alert
// @Description Delete an alert with id and rev
// @Accept  multipart/form-data
// @Param id path string true "07798e794fbad1663a4a0c811b03df55"
// @Param _rev query string true "1-f91a0d21bd7476a9cd8853da0846223d"
// @Accept json
// @Produce json
// @Success 404 {string} string
// @Router /api/v1/alerts/{id} [delete]
func DeleteAlert(c *gin.Context) {
	db := database.InitDb()

	var json models.Alert
	c.BindJSON(&json)

	docID := c.Params.ByName("id")

	newRev, err := db.Delete(context.TODO(), docID, json.Rev)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"success": "Alert " + docID + " with revision " + newRev + " deleted"})
	}
	return
}

func OptionsAlert(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, POST, PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	c.Next()
}
