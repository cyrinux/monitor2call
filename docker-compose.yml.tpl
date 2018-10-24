version: '3'
services:
  couchdb:
    container_name: couchdb
    image: couchdb
    volumes:
      - "couchdb:/opt/couchdb/data:rw"
    networks:
      - backend
    ports:
     - "5984:5984"
    environment:
     - TZ=$TZ
     - COUCHDB_USER=$COUCHDB_USER
     - COUCHDB_PASSWORD=$COUCHDB_PASSWORD
    restart: unless-stopped

  front:
    container_name: monitor2call_front
    build:
      context: .
      dockerfile: ./dockerfiles/front/Dockerfile
    depends_on:
     - couchdb
    ports:
     - "3000:3000"
    networks:
      - front
      - backend
    links:
     - "couchdb:couchdb"
    volumes:
     - "./keys:/go/keys:ro"
     - "./front/cache:/go/cache:rw"
    environment:
     - TZ=$TZ
     - GIN_MODE=release
     - HTTPS_ENABLED=$HTTPS_ENABLED
     - PUSHOVER_APP_API_KEY=$PUSHOVER_APP_API_KEY
     - GOOGLE_APPLICATION_CREDENTIALS=/go/keys/google_api_keys.json
     - COUCHDB_URL=http://$COUCHDB_USER:$COUCHDB_PASSWORD@$COUCHDB_HOST:5984/
     - COUCHDB_NAME=$COUCHDB_NAME
     - PUBLIC_URL=$PUBLIC_URL
     - CACHE_DIR=/go/cache
     - WRITE_PASSWORD=$WRITE_PASSWORD
     - READ_PASSWORD=$READ_PASSWORD
    restart: unless-stopped

  worker:
    container_name: monitor2call_worker
    build:
      context: .
      dockerfile: ./dockerfiles/worker/Dockerfile
    depends_on:
     - couchdb
     - front
    networks:
      - backend
    links:
     - "couchdb:couchdb"
    environment:
     - TZ=$TZ
     - PUSHOVER_APP_API_KEY=$PUSHOVER_APP_API_KEY
     - COUCHDB_URL=http://$COUCHDB_USER:$COUCHDB_PASSWORD@$COUCHDB_HOST:5984/
     - COUCHDB_NAME=$COUCHDB_NAME
     - PUBLIC_URL=$PUBLIC_URL
    restart: unless-stopped

  sms_worker:
    container_name: monitor2call_sms_worker
    build:
      context: .
      dockerfile: ./dockerfiles/sms_worker/Dockerfile
    depends_on:
     - couchdb
     - front
    networks:
      - backend
    links:
     - "couchdb:couchdb"
    environment:
     - TZ=$TZ
     - COUCHDB_URL=http://$COUCHDB_USER:$COUCHDB_PASSWORD@$COUCHDB_HOST:5984/
     - COUCHDB_NAME=$COUCHDB_NAME
     - PUBLIC_URL=$PUBLIC_URL
     - SMS_ENABLED=$SMS_ENABLED
     - SMS_OVH_ACCOUNT=$SMS_OVH_ACCOUNT
     - SMS_OVH_LOGIN=$SMS_OVH_LOGIN
     - SMS_OVH_PASSWORD=$SMS_OVH_PASSWORD
     - SMS_OVH_SENDER=$SMS_OVH_SENDER
     - SMS_FOR_MONITOR=$SMS_FOR_MONITOR
    restart: unless-stopped

volumes:
  couchdb:
    driver: local

networks:
  front:
  backend: