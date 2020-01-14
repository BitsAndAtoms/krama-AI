package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
)

func getOrderByOrderID(w http.ResponseWriter, r *http.Request) {

	if !pre(w, r) {
		return
	}

	var jx []byte

	redisC := REDISCLIENT.Get(r.URL.Path)

	if redisC.Err() != redis.Nil {

		jx = []byte(redisC.Val())

	} else {

		pth := strings.Split(r.URL.Path, "/")
		oid := pth[len(pth)-1]

		dbcol := REDISCLIENT.Get(r.Header.Get("x-access-token")).Val() + OrdersExtension

		results := findMongoDocument(ExternalDB, dbcol, bson.M{"orderid": oid})

		if len(results) != 1 {
			respondWith(w, r, nil, OrderNotFoundMessage, nil, http.StatusNotFound)
			return
		}

		j, err0 := bson.MarshalExtJSON(results[0], false, false)

		if err0 != nil {
			respondWith(w, r, err0, HTTPInternalServerErrorMessage, nil, http.StatusInternalServerError)
			return
		}

		jx = j

		REDISCLIENT.Set(r.URL.Path, j, 0)

	}

	var order ORDER

	err1 := json.Unmarshal([]byte(jx), &order)

	if err1 != nil {
		respondWith(w, r, err1, HTTPInternalServerErrorMessage, nil, http.StatusInternalServerError)
		return
	}

	respondWith(w, r, nil, OrderFoundMessage, order, http.StatusOK)

}

func getOrderByCustomerID(w http.ResponseWriter, r *http.Request) {

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

		dbcol := REDISCLIENT.Get(r.Header.Get("x-access-token")).Val() + OrdersExtension

		results := findMongoDocument(ExternalDB, dbcol, bson.M{"customerid": cid})

		if len(results) != 1 {
			respondWith(w, r, nil, OrderNotFoundMessage, nil, http.StatusNotFound)
			return
		}

		j, err0 := bson.MarshalExtJSON(results[0], false, false)

		if err0 != nil {
			respondWith(w, r, err0, HTTPInternalServerErrorMessage, nil, http.StatusInternalServerError)
			return
		}

		jx = j

		REDISCLIENT.Set(r.URL.Path, j, 0)

	}

	var order ORDER

	err1 := json.Unmarshal([]byte(jx), &order)

	if err1 != nil {
		respondWith(w, r, err1, HTTPInternalServerErrorMessage, nil, http.StatusInternalServerError)
		return
	}

	respondWith(w, r, nil, OrderFoundMessage, order, http.StatusOK)

}

func postOrder(w http.ResponseWriter, r *http.Request) {

	if !pre(w, r) {
		return
	}

	var order ORDER

	err := json.NewDecoder(r.Body).Decode(&order)

	if err != nil {
		respondWith(w, r, err, HTTPBadRequestMessage, nil, http.StatusBadRequest)
		return
	}

	order.OrderCreationDate = time.Now().UnixNano()

	dbcol := REDISCLIENT.Get(r.Header.Get("x-access-token")).Val() + OrdersExtension

	insertMongoDocument(ExternalDB, dbcol, order)

	respondWith(w, r, nil, OrderCreatedMessage, order, http.StatusCreated)

}

func putOrder(w http.ResponseWriter, r *http.Request) {

	if !pre(w, r) {
		return
	}

	var order ORDER

	err := json.NewDecoder(r.Body).Decode(&order)

	if err != nil {
		respondWith(w, r, err, HTTPBadRequestMessage, nil, http.StatusBadRequest)
		return
	}

	dbcol := REDISCLIENT.Get(r.Header.Get("x-access-token")).Val() + OrdersExtension

	result := updateMongoDocument(ExternalDB, dbcol, bson.M{"orderid": order.OrderID}, bson.M{"$set": order})

	if result[0] == 1 && result[1] == 1 {

		REDISCLIENT.Del(r.URL.Path)
		respondWith(w, r, nil, OrderUpdatedMessage, order, http.StatusAccepted)

	} else if result[0] == 1 && result[1] == 0 {

		respondWith(w, r, nil, OrderNotUpdatedMessage, nil, http.StatusNotModified)

	} else if result[0] == 0 && result[1] == 0 {

		respondWith(w, r, nil, OrderNotFoundMessage, nil, http.StatusNotModified)

	}

}

func deleteOrder(w http.ResponseWriter, r *http.Request) {

	if !pre(w, r) {
		return
	}

	dbcol := REDISCLIENT.Get(r.Header.Get("x-access-token")).Val() + OrdersExtension

	pth := strings.Split(r.URL.Path, "/")
	oid := pth[len(pth)-1]

	results := findMongoDocument(ExternalDB, dbcol, bson.M{"orderid": oid})

	if len(results) != 1 {
		respondWith(w, r, nil, OrderNotFoundMessage, nil, http.StatusNotFound)
		return
	}

	j, err0 := bson.MarshalExtJSON(results[0], false, false)

	if err0 != nil {
		respondWith(w, r, err0, HTTPInternalServerErrorMessage, nil, http.StatusInternalServerError)
		return
	}

	var order ORDER

	err1 := json.Unmarshal([]byte(j), &order)

	if err1 != nil {
		respondWith(w, r, err1, HTTPInternalServerErrorMessage, nil, http.StatusInternalServerError)
		return
	}

	if deleteMongoDocument(ExternalDB, dbcol, bson.M{"orderid": oid}) == 1 {

		REDISCLIENT.Del(r.URL.Path)
		respondWith(w, r, nil, OrderDeletedMessage, nil, http.StatusOK)

	} else {

		respondWith(w, r, nil, OrderNotFoundMessage, nil, http.StatusNotModified)

	}

}