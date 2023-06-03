package config

import (
	"fmt"
	"github.com/joeshaw/envdecode"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	os.Setenv("DATABASE_NAME", "db")
	os.Setenv("DATABASE_PORT", "9999")

	cfg := AppConfig()

	if cfg.Db.DbName != "db" {
		t.Errorf("expected %q, got %q", "db", os.Getenv("DATABASE_NAME"))
	}
	if cfg.Db.DbPort != 9999 {
		t.Errorf("expected %q, got %q", "9999", os.Getenv("DATABASE_PORT"))
	}
}

func ExampleAppConfig() {
	type exampleStruct struct {
		String string `env:"STRING"`
	}
	os.Setenv("STRING", "an example string!")

	var e exampleStruct
	err := envdecode.StrictDecode(&e)
	if err != nil {
		panic(err)
	}

	// if STRING is set, e.String will contain its value
	fmt.Println(e.String)

	// Output:
	// an example string!

}
