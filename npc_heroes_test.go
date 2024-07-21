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

	heroes, ok := root.Get("DOTAHeroes")
	if !ok {
		t.Error(errors.New("missing key DOTAHeroes"))
		return
	}

	antimage, ok := heroes.Get("npc_dota_hero_antimage")
	if !ok {
		t.Error(errors.New("missing key npc_dota_hero_antimage"))
		return
	}

	heroID, ok := antimage.GetInt("HeroID")
	if !ok {
		t.Error(errors.New("missing key HeroID"))
		return
	}

	if heroID != 1 {
		t.Error(errors.New("wrong HeroID"))
		return
	}
}
