package session

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/qingants/pandora/miniorm/log"
	"github.com/qingants/pandora/miniorm/schema"
)

func (s *Session) Model(value any) *Session {
	if s.schema == nil || reflect.TypeOf(value) != reflect.TypeOf(s.schema.Model) {
		s.schema = schema.Parse(value, s.dialect)
	}
	return s
}

func (s *Session) Schema() *schema.Schema {
	if s.schema == nil {
		log.Error("Model is not set")
	}
	return s.schema
}

func (s *Session) CreateTable() error {
	table := s.Schema()
	var columns []string

	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Tag))
	}
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE `%s` (%s)", table.Name, desc)).Exec()
	return err
}

func (s *Session) DropTable() error {
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.Schema().Name)).Exec()
	return err
}

func (s *Session) HasTable() bool {
	sql, values := s.dialect.TableExistSQL(s.Schema().Name)
	row := s.Raw(sql, values...).QueryRow()
	var tmp string
	_ = row.Scan(&tmp)
	return tmp == s.Schema().Name
}
