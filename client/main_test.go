package main

import (
	"testing"
)

func BenchmarkCreate(b *testing.B) {
	client, connection := CreateConnection()
	for i := 0; i < b.N; i++ {
		GetGames(&client)
	}
	connection.Close()
}
