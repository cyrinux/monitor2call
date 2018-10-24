package database

import (
	"context"

	"github.com/cyrinux/monitor2call/helpers" // common helpers lib
	// "github.com/go-kivik/kivik" // kivik
	"github.com/flimzy/kivik"       // Stable version of Kivik
	_ "github.com/go-kivik/couchdb" // try
	log "github.com/sirupsen/logrus"
)

// InitDb connect to couchDB
func InitDb() *kivik.DB {
	dbURL := helpers.GetEnv("COUCHDB_URL", "http://localhost:5984/")
	dbName := helpers.GetEnv("COUCHDB_NAME", "monitor2call")

	client, err := kivik.New(context.TODO(), "couch", dbURL)
	if err != nil {
		log.Printf("Can't connect to couchdb database: %s", err.Error())
	}

	// create the database
	client.CreateDB(context.TODO(), dbName)

	db, err := client.DB(context.TODO(), dbName)
	if err != nil {
		log.Printf("Can't login to database %s: %s", dbName, err.Error())
	}

	return db
}

// CreateOrUpdateView manage couchdb view
func CreateOrUpdateView(db *kivik.DB) error {
	// create the users view pushover_users
	rev, err := db.Put(context.TODO(), "_design/users", map[string]interface{}{
		"_id": "_design/users",
		"views": map[string]interface{}{
			"with_pushover": map[string]interface{}{
				"map": "function (doc) { if (doc.tags && doc.name && doc.name != '' && doc.email && doc.email != '' && doc.pushover_api_key && doc.pushover_api_key != '') { emit(doc.name, doc._id); }}",
			},
			"all": map[string]interface{}{
				"map": "function (doc) { if (doc.tags && doc.name && doc.name != '' && doc.email && doc.email != '') { emit(doc.name, doc._id); }}",
			},
		},
	})
	if err == nil && rev != "" {
		log.Print("View _design/pushover_users created/updated")
	}

	// create the unknowledged alerts views
	rev, err = db.Put(context.TODO(), "_design/alerts", map[string]interface{}{
		"_id": "_design/alerts",
		"views": map[string]interface{}{
			"unacknowledged": map[string]interface{}{
				"map": "function (doc){if(doc.message&&!doc.pushover_acknowledged&&!doc.sms_acknowledged&&!doc.language){emit(doc._id,null);}}",
			},
			"need_sms": map[string]interface{}{
				"map": "function(doc){if(doc.acknowledgements&&doc.message&&!doc.sms_sent&&!doc.pushover_acknowledged&&!doc.sms_acknowledged&&!doc.language){var dateNow=new Date(1000*doc.created_at);emit([dateNow],doc._id)}}",
			},
			// "sms_not_ack": map[string]interface{}{
			// 	"map": "function(doc) {if (!doc.pushover_acknowledged && !doc.sms_acknowledged && doc.sms_sent && !doc.language) {emit(doc._id, null);}}",
			// },
		},
	})
	if err == nil && rev != "" {
		log.Println("View _design/unacks_alerts created/updated")
	}

	return err
}
