package vdf_test

import (
	"encoding/json"
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
	kv := vdf.Parse(dat, nil)

	p, _ := json.MarshalIndent(&kv, "", "\t")
	os.WriteFile(path.Join("./var/", "items_game.json"), p, 0666)

	kv.Print()
}
