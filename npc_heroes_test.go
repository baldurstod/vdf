package vdf_test

import (
	"errors"
	"os"
	"path"
	"testing"

	"github.com/baldurstod/vdf"
)

func TestNpcHeroes(t *testing.T) {
	filename := "npc_heroes.txt"

	dat, err := os.ReadFile(path.Join("./var/", filename))

	if err != nil {
		t.Error(err)
		return
	}

	vdf := vdf.VDF{}
	root := vdf.Parse(dat)

	heroes, err := root.Get("DOTAHeroes")
	if err != nil {
		t.Error(err)
		return
	}

	antimage, err := heroes.Get("npc_dota_hero_antimage")
	if err != nil {
		t.Error(err)
		return
	}

	heroID, err := antimage.GetInt("HeroID")
	if err != nil {
		t.Error(err)
		return
	}

	if heroID != 1 {
		t.Error(errors.New("wrong HeroID"))
		return
	}
}
