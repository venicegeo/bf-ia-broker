package csvcolumnmap

import (
	"errors"
	"log"
)

//csvNamedColumn matches a canonical name to a column index in a csv file
type csvNamedColumn struct {
	index int
	key   string
}

//CsvColumnMap matches the name and index of columns
type CsvColumnMap struct {
	//indexMap maps the index of a column to its string key
	entries []csvNamedColumn
}

//CreateValueMap creates an empty map suitable for matching
//values to column names
func (m *CsvColumnMap) CreateValueMap() map[string]string {
	return make(map[string]string, len(m.entries))
}

//UpdateMap populates the valueMap with the values read from the csv.
func (m *CsvColumnMap) UpdateMap(rawValues []string, valueMap map[string]string) {
	for _, namedCol := range m.entries {
		valueMap[namedCol.key] = rawValues[namedCol.index]
	}
}

//New creates a new column map populated with indecies extracted from the provided
//colmnNamesRow
func New(namedColumns []string, columnNamesRow []string) (CsvColumnMap, error) {
	inverseMap := make(map[string]int, len(columnNamesRow))
	for idx, name := range columnNamesRow {
		inverseMap[name] = idx
	}

	var resultMap CsvColumnMap

	entries := make([]csvNamedColumn, len(namedColumns))
	for idx, name := range namedColumns {
		columnIndex, keyExists := inverseMap[name]
		if keyExists {
			entries[idx] = csvNamedColumn{columnIndex, name}
		} else {
			log.Println("Could not find column ", name)
			return resultMap, errors.New("no such column")
		}
	}

	return CsvColumnMap{entries: entries}, nil
}
