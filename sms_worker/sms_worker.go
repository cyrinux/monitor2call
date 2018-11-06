// go worker which send SMS if pushover alert not acked
// in 15min
// https://gobyexample.com/worker-pools
package main

import (
	"context"
	"strconv"
	"time"

	"github.com/cyrinux/monitor2call/database" // database lib
	"github.com/cyrinux/monitor2call/helpers"  // common helpers lib
	"github.com/cyrinux/monitor2call/models"   //structs
	"github.com/kahlys/sms"
	_ "github.com/kahlys/sms/driver/ovh"
	log "github.com/sirupsen/logrus"
)

var nbRows int64 // MUST I REALLY DO THIS???

// iniSmsAPI initialize ovh sms app
func initSmsAPI() (sender *sms.Sender, err error) {
	param := map[string]string{
		"account":  helpers.GetEnv("SMS_OVH_ACCOUNT", ""),
		"login":    helpers.GetEnv("SMS_OVH_LOGIN", ""),
		"password": helpers.GetEnv("SMS_OVH_PASSWORD", ""),
		"sender":   helpers.GetEnv("SMS_OVH_SENDER", ""),
	}

	sender, err = sms.Init("ovh", param)
	if err != nil {
		log.Error(err)
	}
	return sender, err
}

func sendSMS(message string, phone string) (err error) {
	sender, _ := initSmsAPI()
	err = sender.Send(message, phone)
	if err != nil {
		log.Error(err)
	}
	return err
}

// Here's the worker, of which we'll run several
// concurrent instances. These workers will receive
// work on the `alerts` channel and send the corresponding
// results on `results`.
// ! result channel not used for the moment !
func worker(id int, alerts <-chan *models.Alert, results chan<- *models.Alert) {
	// Monitor for which send SMS if not quickly ack
	// TODO: Must be an array
	smsEnabled, _ := strconv.ParseBool(helpers.GetEnv("SMS_ENABLED", "false"))
	smsForMonitor := helpers.GetEnv("SMS_FOR_MONITOR", "internetvista")

	for {
		// read alerts from alerts channel
		// alerts come from couchDB view query in main
		alert := <-alerts

		// allCanceled boolean default true
		// if all recipients have ack the alert
		// boolean will be true
		allAcknowledged := true

		// object is changed
		changed := false

		log.Infof("worker %v started job %s, will send SMS for %s alert", id, alert.ID, smsForMonitor)

		var fifteenMinutes int64
		fifteenMinutes = 15 * 60

		// loop over all the alert recipients
		for idx, ack := range alert.Acknowledgements {
			now := time.Now().Unix()
			createdAt := alert.CreatedAt
			if now-createdAt > fifteenMinutes {
				nbRows++
				log.Infof("Alert %s is enought old: %v seconds > %v seconds", alert.ID, now-createdAt, fifteenMinutes)
				// if pushover not ACK and
				// if the SMS per recipient not send
				// send it.
				if ack.PushoverAcknowledged == false &&
					ack.SmsSent == false && alert.Monitor == smsForMonitor {
					if smsEnabled == true {
						// really send the SMS if enabled in env
						if alert.Acknowledgements[idx].PushoverAcknowledged == false &&
							alert.Acknowledgements[idx].SmsAcknowledged == false {
							log.Infof("Sending SMS to %s for alert %s on phone %s",
								ack.User,
								alert.ID,
								ack.Phone,
							)
							err := sendSMS(alert.Message, alert.Acknowledgements[idx].Phone)
							if err != nil {
								log.Infof("Error while send the SMS: %+v", err)
								allAcknowledged = false
							} else {
								log.Infof("Sms sent with return code: %+v", err)
								// flag sms as send and keep time
								if alert.Acknowledgements[idx].SmsSent == false {
									alert.Acknowledgements[idx].SmsSent = true
									alert.Acknowledgements[idx].SmsSentAt = time.Now()
									changed = true
								}
							}
						}
					} else {
						log.Warn("!!! WARNING SMS SENDING IS DISABLED, !!!")
						log.Warn("!!! ENABLE IT IN PRODUCTION WITH 'SMS_ENABLED=true' !!!")
						if alert.Acknowledgements[idx].SmsAcknowledged == false {
							alert.Acknowledgements[idx].SmsAcknowledged = true
							changed = true
						}
					}
				} else {
					log.Infof("Alert %s too young to send a SMS", alert.ID)
				}
			}

			// If all SMS are sent, SMS ack the alert
			if alert.SmsAcknowledged != allAcknowledged {
				alert.SmsAcknowledged = allAcknowledged
				changed = true
			}

		}

		if changed == true {
			results <- alert
		} else {
			// create a nil alert to be sure to send the same
			// number of rows in ouput as the input
			// if not, worker can stay block
			var alert *models.Alert
			results <- alert
		}

		time.Sleep(time.Second)
	}
}

func main() {
	db := database.InitDb()

	alerts := make(chan *models.Alert, 101) // couchDB query size limit set to 10 so 10+1
	results := make(chan *models.Alert, 101)

	var alert models.Alert

	//  This starts up 5 workers, initially blocked
	// because there are no alerts yet.
	// for w := 1; w <= 5; w++ {
	// 	go worker(w, alerts, results)
	// }
	go worker(1, alerts, results)

	// Main Loop
	for {
		// In order to use our pool of workers we need to send
		// them work and collect their results. We make 2
		// channels for this.
		// Fetch all unack alerts

		// query limit set to 100 max
		// http://docs.couchdb.org/en/stable/api/ddoc/views.html#using-limits-and-skipping-rows
		// https://godoc.org/github.com/go-kivik/kivik#Options
		options := make(map[string]interface{})
		// max result of the query
		options["limit"] = 100

		// Query will return alert from now - startCount
		now := time.Now()
		startCount := 24 * 60
		startKey := now.Add(time.Duration(-startCount) * time.Minute)
		options["startkey"] = "[\"" + startKey.Format(time.RFC3339) + "\"]"
		// to now - endCount
		endCount := 1
		endKey := now.Add(time.Duration(-endCount) * time.Minute)
		options["endkey"] = "[\"" + endKey.Format(time.RFC3339) + "\"]"
		log.Infof("Start Key: %s End Key: %s", startKey.Format(time.RFC3339), endKey.Format(time.RFC3339))

		// query the view
		rows, err := db.Query(context.TODO(), "_design/alerts", "_view/need_sms", options)
		if err != nil {
			log.Error(err)
			// skip if cant be read
			continue
		}
		// loop over unack alert
		// get the doc alert,
		// then wait for 1 minute and LOOP
		for rows.Next() {
			id := rows.ID()
			row, err := db.Get(context.TODO(), id)
			if err != nil {
				log.Errorf("Can't get the alert from database with id: %s", id)
			}
			if err := row.ScanDoc(&alert); err != nil {
				log.Error(err)
			}
			// send alert to alerts channel
			alerts <- &alert
		}

		// read all acked alert
		// number of alerts in results channel should be the total
		// of rows in view unacks_alerts
		var iterator int64
		// loop
		for iterator = 1; iterator <= nbRows; iterator++ {
			// read from results channel acknowleged alert

			smsAlert := <-results

			log.Infof("interator %v", iterator)
			// if alert is nil, skip the next
			if smsAlert == nil {
				continue
			}

			var newSmsAlert models.Alert

			row, err := db.Get(context.TODO(), smsAlert.ID)
			if err != nil {

				log.Infof("Can't get the document %s: %s", smsAlert.ID, err.Error())
			}

			err = row.ScanDoc(&newSmsAlert)
			if err != nil {

				log.Infof("Can't read the document %s", err.Error())
			}

			newSmsAlert.Acknowledgements = smsAlert.Acknowledgements
			newSmsAlert.SmsAcknowledged = smsAlert.SmsAcknowledged

			rev, err := db.Put(context.TODO(), newSmsAlert.ID, &newSmsAlert)
			if err != nil {

				log.Errorf("Error while writing to db: %s", err.Error())

			} else {

				log.Infof("Alert status updated with rev: %s", rev)
			}

		}
		// reset rows numbers
		nbRows = 0
		// Pause between two loop
		log.Info("Loop in 15 seconds...")
		time.Sleep(15 * time.Second)

	}
}

// Send mutli queries
// ❯ curl -g  'http://127.0.0.1:5984/monitor2call/_design/unacks_alerts/_view/need_sms?startkey=["2018-10-17T12:10:57.653Z"]&endkey=["2018-10-17T12:19:57.653Z"]'
// {"total_rows":2,"offset":0,"rows":[
// {"id":"6d45677a38434d60499ee96cd0000867","key":["2018-10-17T12:19:49.233Z"],"value":"6d45677a38434d60499ee96cd0000867"},
// {"id":"bfee459ca7ccf8aac3ac4e808c000a24","key":["2018-10-17T12:19:49.509Z"],"value":"bfee459ca7ccf8aac3ac4e808c000a24"}
// ]}

// ❯ curl -g  'http://127.0.0.1:5984/monitor2call/_design/unacks_alerts/_view/need_sms?startkey=["2018-10-17T12:10:57.653Z"]&endkey=["2018-10-17T12:17:57.653Z"]'
// {"total_rows":2,"offset":0,"rows":[
// ]}

// ❯ curl -g  'http://127.0.0.1:5984/monitor2call/_design/unacks_alerts/_view/need_sms?startkey=["2018-10-17T12:10:57.653Z"]&endkey=["2018-10-17T12:17:57.653Z"]'
// {"total_rows":2,"offset":0,"rows":[

// ]}
