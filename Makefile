# DO NO EDIT
# use like this
# make TZ=Europe/Berlin COUCHDB_USER=user ... up 
TZ?=Europe/Paris
export TZ
COUCHDB_USER?=m2c
export COUCHDB_USER
COUCHDB_PASSWORD?=m2c
export COUCHDB_PASSWORD
COUCHDB_HOST?=couchdb
export COUCHDB_HOST
COUCHDB_NAME?=monitor2call
export COUCHDB_NAME
COUCHDB_URL?=http://$$COUCHDB_USER:$$COUCHDB_PASSWORD@$$COUCHDB_HOST:5984/
export COUCHDB_URL
PUBLIC_URL?=http://localhost:3000/
export PUBLIC_URL
WRITE_PASSWORD?=WriteMonitor2Call
export WRITE_PASSWORD
READ_PASSWORD?=ReadMonitor2Call
export READ_PASSWORD
SMS_ENABLED?=false
export SMS_ENABLED
SMS_FOR_MONITOR?=internetvista
export SMS_FOR_MONITOR
SMS_ENABLED?=
export SMS_ENABLED
SMS_OVH_ACCOUNT?=
export SMS_OVH_ACCOUNT
SMS_OVH_LOGIN?=
export SMS_OVH_LOGIN
SMS_OVH_PASSWORD?=
export SMS_OVH_PASSWORD
SMS_OVH_SENDER?=
export SMS_OVH_SENDER
SMS_FOR_MONITOR?=
export SMS_FOR_MONITOR
HTTPS_ENABLED?=false
export HTTPS_ENABLED

.PHONY: all
all: dev

dev:
	cd front && $(MAKE) dev
	cd worker && $(MAKE) dev
	cd sms_worker && $(MAKE) dev
	
release: docker-front docker-worker docker-sms-worker

docker: compose up logs

run:
	cd front && $(MAKE) run
	cd worker && $(MAKE) run
	cd sms_worker && $(MAKE) run

compose:	
	[ -f ./docker-compose.yml ] || envsubst < ./docker-compose.yml.tpl > ./docker-compose.yml
	sudo docker-compose build

up: compose
	sudo docker-compose up -d

clean:
	rm -f docker-compose.yml

down:
	sudo docker-compose down

logs:
	sudo docker-compose logs -f --tail=30

docker-front:
		cd front && $(MAKE) release
		
docker-worker:
		cd worker && $(MAKE) release

docker-sms-worker:
		cd sms_worker && $(MAKE) release
