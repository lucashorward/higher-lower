/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "example.com/higher-lower-game/higherlower"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func CreateConnection() (client pb.HigherLowerGameClient, conn *grpc.ClientConn) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client = pb.NewHigherLowerGameClient(conn)
	return client, conn
}

func GetGames(connection *pb.HigherLowerGameClient) {
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := (*connection).GetGameState(ctx, &pb.GameState{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	r.GetGame()
	// fmt.Println(game)
}

func main() {
	client, connection := CreateConnection()
	GetGames(&client)
	connection.Close()
}
