// go worker which must ACK in couchDB
// acked pushover alerts
// https://gobyexample.com/worker-pools
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cyrinux/monitor2call/database"
	"github.com/cyrinux/monitor2call/models"
	"github.com/cyrinux/monitor2call/notifications"
	"github.com/cyrinux/monitor2call/version"
	log "github.com/sirupsen/logrus"
)

// Here's the worker, of which we'll run several
// concurrent instances. These workers will receive
// work on the `alerts` channel and send the corresponding
// results on `results`.
// ! result channel not used for the moment !
func worker(id int, alerts <-chan *models.Alert, results chan<- *models.Alert) {
	app := notifications.InitPushover()

	for {
		// read alerts from alerts channel
		// alerts come from couchDB view query in main
		alert := <-alerts

		// allAcknowledged boolean default true
		// if all recipients have ack the alert
		// boolean will be true
		allAcknowledged := true

		// object is changed
		changed := false

		log.Infof("Worker %v started job %s", id, alert.ID)

		// loop over all the alert recipients
		for idx, ack := range alert.Acknowledgements {
			if ack.PushoverReceipt != "" && ack.PushoverAcknowledged == false {
				log.Infof("Checking alert %s sent to name %s with receipt: %s",
					alert.ID,
					ack.User,
					ack.PushoverReceipt,
				)
				// check with pushover api if recipient have ack the alert
				receiptDetails, err := app.GetReceiptDetails(ack.PushoverReceipt)
				if err != nil {
					log.Error(err)
				}

				// if acked
				if receiptDetails.Acknowledged {
					// read when it was ack
					ackTime := *receiptDetails.AcknowledgedAt
					log.Infof("Alert %s acknowledged status : %v at %s",
						alert.ID,
						receiptDetails.Acknowledged,
						receiptDetails.AcknowledgedAt,
					)
					// set it to the alert
					if alert.Acknowledgements[idx].PushoverAcknowledged == false {
						alert.Acknowledgements[idx].PushoverAcknowledged = true
						alert.Acknowledgements[idx].PushoverAcknowledgedAt = ackTime
						changed = true
					}
				} else {
					// if recipient have not ack
					// the global alert if not ack
					allAcknowledged = false
				}

			}
		}

		// set global acknowledged alert state
		if alert.PushoverAcknowledged != allAcknowledged {
			alert.PushoverAcknowledged = allAcknowledged
			changed = true
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
	var isVerbose bool
	var isVersion bool
	flag.BoolVar(&isVerbose, "v", false, "print verbose output")
	flag.BoolVar(&isVersion, "version", false, "print version and exit")
	flag.Parse()

	if isVersion {
		fmt.Println("Monitor2Call Pushover worker app version:", version.String())
		os.Exit(0)
	}

	// set log formatter
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	log.Debugf("Starting Monitor2Call Pushover worker app v%v", version.String())

	// init database
	db := database.InitDb()

	// create channels
	// couchDB query size limit set to 100 so 100+1
	alerts := make(chan *models.Alert, 101)
	results := make(chan *models.Alert, 101)

	var alert models.Alert

	//  This starts up 5 workers, initially blocked
	// because there are no alerts yet.
	for w := 1; w <= 5; w++ {
		go worker(w, alerts, results)
	}

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
		options["limit"] = 100
		rows, err := db.Query(context.TODO(), "_design/alerts", "_view/unacknowledged", options)
		if err != nil {
			log.Fatal(err)
		} else {
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
				alerts <- &alert
			}
		}

		// read all acked alert
		// number of alerts in results channel should be the total
		// of rows in view unacks_alerts
		var iterator int64 // rows.TotalRows() is an int64
		for iterator = 1; iterator <= rows.TotalRows(); iterator++ {
			// read from results channel acknowleged alert
			ackedAlert := <-results

			// if alert is nil, skip the next
			if ackedAlert == nil {
				continue
			}

			// log.Infof("%+v", smsAlert)
			var newAckedAlert models.Alert

			row, err := db.Get(context.TODO(), ackedAlert.ID)
			if err != nil {
				log.Errorf("Can't get the document %s: %s", ackedAlert.ID, err.Error())
			}

			err = row.ScanDoc(&newAckedAlert)
			if err != nil {
				log.Errorf("Can't read the document %s", err.Error())
			}
			newAckedAlert.Acknowledgements = ackedAlert.Acknowledgements
			newAckedAlert.PushoverAcknowledged = ackedAlert.PushoverAcknowledged

			rev, err := db.Put(context.TODO(), newAckedAlert.ID, &newAckedAlert)
			if err != nil {
				log.Errorf("Error while writing to db: %s", err.Error())
			} else {
				log.Infof("Alert status updated with rev: %s", rev)
			}
		}

		// Pause between two loop
		log.Info("Loop in 15 seconds...")
		time.Sleep(15 * time.Second)

	}
}
