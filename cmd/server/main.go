package main

import (
	"code.in.spdigital.sg/sp-digital/dts-cops-pg-mongo/mongo"
	"code.in.spdigital.sg/sp-digital/dts-cops-pg-mongo/postgres"
)

func main() {
	postgres.MeasurePostgres()
	mongo.MeasureMongo()
}
