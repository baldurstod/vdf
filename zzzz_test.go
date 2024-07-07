package vdf_test

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/baldurstod/vdf"
)

const varFolder = "./var/"

func TestSFMAnimationGroups(t *testing.T) {
	filename := "sfm_default_animation_groups.vcfg"

	dat, err := os.ReadFile(path.Join(varFolder, filename))

	if err != nil {
		t.Error(err)
		return
	}

	vdf := vdf.VDF{}
	root := vdf.Parse(dat)
	root.Print()
	log.Println(root)
}
