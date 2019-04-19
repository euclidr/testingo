package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAnimal(t *testing.T) {
	tests := []struct {
		name   string
		place  string
		hasErr bool
	}{
		{name: "Kangaroo", place: "Australia", hasErr: false},
		{name: "羊驼", place: "南美洲", hasErr: false},
		{name: "", place: "Outer Space", hasErr: true},
	}

	for row, test := range tests {
		animal, err := CreateAnimal(test.name, test.place)
		if test.hasErr {
			assert.Error(t, err, "row: %d", row)
			continue
		}
		assert.NoError(t, err, "row: %d", row)
		assert.Equal(t, test.name, animal.Name)
		assert.Equal(t, test.place, animal.Place)
	}

	// clean Database
	_, err := DBConn().Exec("TRUNCATE TABLE animal")
	assert.NoError(t, err)
}

func TestAnimalChangePlace(t *testing.T) {
	tests := []struct {
		name     string
		place    string
		newPlace string
		hasErr   bool
	}{
		{name: "Kangaroo", place: "Anywhere", newPlace: "Australia", hasErr: false},
		{name: "三文鱼", place: "日本", newPlace: "挪威", hasErr: false},
		{name: "角马", place: "非洲", newPlace: "", hasErr: true},
	}

	for row, test := range tests {
		animal, err := CreateAnimal(test.name, test.place)
		assert.NoError(t, err, "row: %d", row)
		err = animal.ChangePlace(test.newPlace)
		if test.hasErr {
			assert.Error(t, err, "row: %d", row)
			continue
		}
		assert.NoError(t, err, "row: %d", row)
		assert.Equal(t, test.newPlace, animal.Place)
	}

	// clean Database
	_, err := DBConn().Exec("TRUNCATE TABLE animal")
	assert.NoError(t, err)
}
