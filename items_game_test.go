package vdf_test

import (
	"os"
	"path"
	"testing"

	"github.com/baldurstod/vdf"
)

func TestItems(t *testing.T) {
	filename := "items_game.txt"

	dat, err := os.ReadFile(path.Join("./var/", filename))

	if err != nil {
		t.Error(err)
		return
	}

	vdf := vdf.VDF{}
	_ = vdf.Parse(dat)
}
