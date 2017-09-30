package config

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

const (
	FLAG_COMMENT = '#'
	FLAG_FIELD   = '$'
	SPLIT_CHAR   = '\t'
	SPLIT_TOKEN  = '='
)

var (
	UTF8_BOM = [3]byte{0xef, 0xbb, 0xbf}

	ErrNotUtf8     = errors.New("Err NotUtf8")
	ErrFieldsEmpty = errors.New("Err FieldsEmpty")
)

type FileTableParse struct {
	fields []string
	lines  [][]string
}

func (parse *FileTableParse) line(i int) []string {
	if len(parse.lines) <= i {
		return nil
	}
	return parse.lines[i]
}

func (parse *FileTableParse) fieldIdx(field string) int {
	for i, k := range parse.fields {
		if k == field {
			return i
		}
	}
	return -1
}

func (parse *FileTableParse) getItem(i int, field string) (bool, string) {
	idx := parse.fieldIdx(field)
	if idx < 0 {
		return false, ""
	}
	fields := parse.line(i)
	if fields != nil && len(fields) > idx {
		return true, fields[idx]
	}
	return false, ""
}

func (parse *FileTableParse) GetItemFloat32(i int, field string) float32 {
	if ok, item := parse.getItem(i, field); ok {
		if value, err := strconv.ParseFloat(item, 32); err == nil {
			return float32(value)
		}
	}
	return 0
}

func (parse *FileTableParse) GetItemFloat64(i int, field string) float64 {
	if ok, item := parse.getItem(i, field); ok {
		if value, err := strconv.ParseFloat(item, 32); err == nil {
			return value
		}
	}
	return 0
}

func (parse *FileTableParse) GetItemInt64(i int, field string) int64 {
	if ok, item := parse.getItem(i, field); ok {
		if value, err := strconv.ParseInt(item, 10, 64); err == nil {
			return value
		}
	}
	return 0
}

func (parse *FileTableParse) GetItemInt32(i int, field string) int32 {
	if ok, item := parse.getItem(i, field); ok {
		if value, err := strconv.ParseInt(item, 10, 32); err == nil {
			return int32(value)
		}
	}
	return 0
}

func (parse *FileTableParse) GetItemUint64(i int, field string) uint64 {
	if ok, item := parse.getItem(i, field); ok {
		if value, err := strconv.ParseUint(item, 10, 64); err == nil {
			return value
		}
	}
	return 0
}

func (parse *FileTableParse) GetItemUint32(i int, field string) uint32 {
	if ok, item := parse.getItem(i, field); ok {
		if value, err := strconv.ParseUint(item, 10, 32); err == nil {
			return uint32(value)
		}
	}
	return 0
}

func (parse *FileTableParse) GetItem(i int, field string) string {
	_, value := parse.getItem(i, field)
	return value
}

func (parse *FileTableParse) Count() int {
	return len(parse.lines)
}

func (parse *FileTableParse) Parse(rd io.Reader) error {
	reader := bufio.NewReader(rd)
	var (
		count int
		blen  int
	)
	parse.fields = nil
	parse.lines = make([][]string, 0, 50)
	for {
		bytes, _, _ := reader.ReadLine()
		blen = len(bytes)
		if blen == 0 {
			break
		}
		count++
		if count == 1 {
			if blen < 3 || UTF8_BOM[0] != bytes[0] || UTF8_BOM[1] != bytes[1] || UTF8_BOM[2] != bytes[2] {
				break
			}
			bytes = bytes[3:]
		}
		if bytes[0] == byte(FLAG_COMMENT) {
			continue
		}
		if bytes[0] == byte(FLAG_FIELD) {
			parse.fields = strings.Split(string(bytes[1:]), string(SPLIT_CHAR))
			continue
		}
		values := strings.Split(string(bytes), string(SPLIT_CHAR))
		parse.lines = append(parse.lines, values)
	}
	if len(parse.fields) == 0 {
		return ErrFieldsEmpty
	}
	return nil
}

type FileKvParse struct {
	fields []string
	values []string
}

func (parse *FileKvParse) fieldIdx(field string) int {
	for i, k := range parse.fields {
		if k == field {
			return i
		}
	}
	return -1
}

func (parse *FileKvParse) getItem(field string) (bool, string) {
	idx := parse.fieldIdx(field)
	if idx < 0 {
		return false, ""
	}
	return true, parse.values[idx]
}

func (parse *FileKvParse) GetItemFloat32(field string) float32 {
	if ok, item := parse.getItem(field); ok {
		if value, err := strconv.ParseFloat(item, 32); err == nil {
			return float32(value)
		}
	}
	return 0
}

func (parse *FileKvParse) GetItemFloat64(field string) float64 {
	if ok, item := parse.getItem(field); ok {
		if value, err := strconv.ParseFloat(item, 32); err == nil {
			return value
		}
	}
	return 0
}

func (parse *FileKvParse) GetItemInt64(field string) int64 {
	if ok, item := parse.getItem(field); ok {
		if value, err := strconv.ParseInt(item, 10, 64); err == nil {
			return value
		}
	}
	return 0
}

func (parse *FileKvParse) GetItemInt32(field string) int32 {
	if ok, item := parse.getItem(field); ok {
		if value, err := strconv.ParseInt(item, 10, 32); err == nil {
			return int32(value)
		}
	}
	return 0
}

func (parse *FileKvParse) GetItemUint64(field string) uint64 {
	if ok, item := parse.getItem(field); ok {
		if value, err := strconv.ParseUint(item, 10, 64); err == nil {
			return value
		}
	}
	return 0
}

func (parse *FileKvParse) GetItemUint32(field string) uint32 {
	if ok, item := parse.getItem(field); ok {
		if value, err := strconv.ParseUint(item, 10, 32); err == nil {
			return uint32(value)
		}
	}
	return 0
}

func (parse *FileKvParse) GetItem(field string) string {
	_, value := parse.getItem(field)
	return value
}

func (parse *FileKvParse) Parse(rd io.Reader) error {
	reader := bufio.NewReader(rd)
	var (
		count int
		blen  int
	)
	parse.values = make([]string, 0, 10)
	parse.fields = make([]string, 0, 10)
	for {
		bytes, _, _ := reader.ReadLine()
		blen = len(bytes)
		if blen == 0 {
			break
		}
		count++
		if count == 1 {
			if blen < 3 || UTF8_BOM[0] != bytes[0] || UTF8_BOM[1] != bytes[1] || UTF8_BOM[2] != bytes[2] {
				break
			}
			bytes = bytes[3:]
		}
		if bytes[0] == byte(FLAG_COMMENT) {
			continue
		}
		values := strings.Split(string(bytes), string(SPLIT_TOKEN))
		if len(values) != 2 {
			continue
		}
		parse.fields = append(parse.fields, strings.TrimSpace(values[0]))
		parse.values = append(parse.values, strings.TrimSpace(values[1]))
	}
	if len(parse.fields) == 0 {
		return ErrFieldsEmpty
	}
	return nil
}
