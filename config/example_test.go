package config

import "fmt"

func ExampleNewConfig() {
	config := &Config{
		FlagAddress:     ":8080",
		FlagBaseAddress: "http://localhost:8080",
		FlagFilePath:    "/tmp/short-url-db.json",
		DataBaseDSN:     "user:password@/dbname",
	}
	fmt.Printf("%+v\n", config)
	// Output:
	// &{FlagAddress::8080 FlagBaseAddress:http://localhost:8080 FlagFilePath:/tmp/short-url-db.json DataBaseDSN:user:password@/dbname}
}
