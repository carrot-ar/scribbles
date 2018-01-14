package main

import (
	"fmt"
	"github.com/carrot-ar/carrot"
	r "github.com/clandry94/go-redis-ringbuffer"
	"github.com/go-redis/redis"
)

const LeftoverObjects = 4096

var ringBuffer *r.RingBuffer

/*
	Ideally controllers could have member fields that self initialize/are "constant", but how?
*/
type DrawController struct{}

func (c *DrawController) Draw(request *carrot.Request, broadcast *carrot.Broadcast) {
	/*
		Creating custom responses is a bit difficult
	*/
	payload, err := carrot.NewPayload(string(request.SessionToken), request.Offset, request.Params)
	if err != nil {
		fmt.Println(err)
		panic("can't create the payload")
	}

	response, err := carrot.NewResponse(string(request.SessionToken), "draw", payload)
	if err != nil {
		panic("can't initialize response")
	}

	sendableResp, err := response.Build()
	if err != nil {
		panic("could not build response")
	}

	ringBuffer.Push(fmt.Sprint(sendableResp))

	broadcast.Broadcast(sendableResp)
}

/*
	EXPERIMENTAL

	NOTE: These objects in the ring buffer have 2 degrees of freedom. They can grow in # of objects which is limited
   		  by the `LeftoverObjects` value as well as the size of the message itself. This *COULD* have an effect
   		  when a client requests a lot of objects in a sync, thus backing up their client send buffer
*/
type SyncController struct{}

func (c *SyncController) Sync(request *carrot.Request, broadcast *carrot.Broadcast) {

	worldObjects, err := ringBuffer.GetRange(0, ringBuffer.Capacity)

	if err != nil {
		panic("could not read world objects")
	}

	for _, object := range worldObjects {
		broadcast.Broadcast([]byte(object), string(request.SessionToken))
	}
}

func main() {
	/*
		Setup redis, our in memory data store
	*/
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	/*
		Create a new ring buffer in redis
	*/
	ringBuffer = r.NewRingBuffer("world_objects", LeftoverObjects, client)

	/*
		Setup the endpoints
	*/
	carrot.Add("draw", DrawController{}, "Draw", true)
	carrot.Add("sync", SyncController{}, "Sync", true)

	carrot.Run()
}
