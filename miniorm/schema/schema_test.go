package schema

import (
	"testing"

	"github.com/qingants/pandora/miniorm/dialect"
)

type Task struct {
	ID   string `miniorm:"PRIMARY KEY"`
	Name string
}

var TestDial, _ = dialect.GetDialect("mysql")

func TestParse(t *testing.T) {
	schema := Parse(&Task{}, TestDial)
	if schema.Name != "Task" || len(schema.Fields) != 2 {
		t.Fatal("failed to parse Task struct")
	}

	if schema.GetField("ID").Tag != "PRIMARY KEY" {
		t.Fatal("failed to parse primary key!")
	}
}
