package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
)

func getCustomer(w http.ResponseWriter, r *http.Request) {

	if !pre(w, r) {
		return
	}

	var jx []byte

	redisC := REDISCLIENT.Get(r.URL.Path)

	if redisC.Err() != redis.Nil {

		jx = []byte(redisC.Val())

	} else {

		pth := strings.Split(r.URL.Path, "/")
		cid := pth[len(pth)-1]

		dbcol := REDISCLIENT.Get(r.Header.Get("x-access-token")).Val() + CustomersCollectionExtension

		results := findMongoDocument(ExternalDB, dbcol, bson.M{"customerid": cid})

		if len(results) != 1 {
			respondWith(w, r, nil, CustomersNotFoundMessage, nil, http.StatusNotFound, false)
			return
		}

		j, err0 := bson.MarshalExtJSON(results[0], false, false)

		if err0 != nil {
			respondWith(w, r, err0, HTTPInternalServerErrorMessage, nil, http.StatusInternalServerError, false)
			return
		}

		jx = j

		REDISCLIENT.Set(r.URL.Path, j, 0)

	}

	var customer CUSTOMER

	err1 := json.Unmarshal([]byte(jx), &customer)

	if err1 != nil {
		respondWith(w, r, err1, HTTPInternalServerErrorMessage, nil, http.StatusInternalServerError, false)
		return
	}

	respondWith(w, r, nil, CustomersFoundMessage, customer, http.StatusOK, false)

}

func postCustomer(w http.ResponseWriter, r *http.Request) {

	if !pre(w, r) {
		return
	}

	var customer CUSTOMER

	err := json.NewDecoder(r.Body).Decode(&customer)

	if err != nil {
		respondWith(w, r, err, HTTPBadRequestMessage, nil, http.StatusBadRequest, false)
		return
	}

	if !validateCustomer(w, r, customer) {
		return
	}

	groomCustomerData(&customer)

	customer.Updated = time.Now().UnixNano()

	dbcol := REDISCLIENT.Get(r.Header.Get("x-access-token")).Val() + CustomersCollectionExtension

	insertMongoDocument(ExternalDB, dbcol, customer)

	respondWith(w, r, nil, CustomersAddedMessage, customer, http.StatusCreated, true)

}

func putCustomer(w http.ResponseWriter, r *http.Request) {

	if !pre(w, r) {
		return
	}

	var customer CUSTOMER

	err := json.NewDecoder(r.Body).Decode(&customer)

	if err != nil {
		respondWith(w, r, err, HTTPBadRequestMessage, nil, http.StatusBadRequest, false)
		return
	}

	if !validateCustomer(w, r, customer) {
		return
	}

	groomCustomerData(&customer)

	customer.Updated = time.Now().UnixNano()

	dbcol := REDISCLIENT.Get(r.Header.Get("x-access-token")).Val() + CustomersCollectionExtension

	result := updateMongoDocument(ExternalDB, dbcol, bson.M{"customerid": customer.CustomerID}, bson.M{"$set": customer})

	if result[0] == 1 && result[1] == 1 {

		resetCustomerCacheKeys(&customer)
		respondWith(w, r, nil, CustomersUpdatedMessage, customer, http.StatusAccepted, true)

	} else if result[0] == 1 && result[1] == 0 {

		respondWith(w, r, nil, CustomersNotUpdatedMessage, nil, http.StatusNotModified, false)

	} else if result[0] == 0 && result[1] == 0 {

		respondWith(w, r, nil, CustomersNotFoundMessage, nil, http.StatusNotModified, false)

	}

}

func deleteCustomer(w http.ResponseWriter, r *http.Request) {

	if !pre(w, r) {
		return
	}

	dbcol := REDISCLIENT.Get(r.Header.Get("x-access-token")).Val() + CustomersCollectionExtension

	pth := strings.Split(r.URL.Path, "/")
	cid := pth[len(pth)-1]

	results := findMongoDocument(ExternalDB, dbcol, bson.M{"customerid": cid})

	if len(results) != 1 {
		respondWith(w, r, nil, CustomersNotFoundMessage, nil, http.StatusNotFound, false)
		return
	}

	j, err0 := bson.MarshalExtJSON(results[0], false, false)

	if err0 != nil {
		respondWith(w, r, err0, HTTPInternalServerErrorMessage, nil, http.StatusInternalServerError, false)
		return
	}

	var customer CUSTOMER

	err1 := json.Unmarshal([]byte(j), &customer)

	if err1 != nil {
		respondWith(w, r, err1, HTTPInternalServerErrorMessage, nil, http.StatusInternalServerError, false)
		return
	}

	if deleteMongoDocument(ExternalDB, dbcol, bson.M{"customerid": cid}) == 1 {

		resetCustomerCacheKeys(&customer)
		respondWith(w, r, nil, CustomersDeletedMessage, nil, http.StatusOK, true)

	} else {

		respondWith(w, r, nil, CustomersNotFoundMessage, nil, http.StatusNotModified, false)

	}

}
