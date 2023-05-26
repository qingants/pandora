package session

import "testing"

type Task struct {
	ID   string `miniorm:"PRIMARY KEY"`
	Name string
}

func TestSession_CreateTable(t *testing.T) {
	s := NewSession().Model(&Task{})
	_ = s.DropTable()
	_ = s.CreateTable()
	if !s.HasTable() {
		t.Fatal("failed to create table user")
	}
}
