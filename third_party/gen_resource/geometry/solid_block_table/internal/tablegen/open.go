package tablegen

import (
	"fmt"

	model "neomega-blocks/geometry/java_block_model"
	java "neomega-blocks/presets/java/v1"
	"neomega-blocks/third_party/gen_resource/geometry/solid_block_table/internal/resources"
)

type OpenOptions struct {
	AssetsDir string
	Policy    model.Policy
}

type Builder struct {
	Bundle    *model.Bundle
	models    *model.Resolver
	materials *model.MaterialLibrary
}

func Open(options OpenOptions) (*Builder, error) {
	assets, err := model.Load(options.AssetsDir)
	if err != nil {
		return nil, err
	}
	if assets.Info.Version != java.JavaTargetVersion {
		return nil, fmt.Errorf("resources are Java %s; Java states target %s", assets.Info.Version, java.JavaTargetVersion)
	}
	if options.Policy == nil {
		options.Policy = resources.PreviewPolicy{}
	}
	engine := &Builder{
		Bundle:    assets,
		models:    model.NewResolver(assets.Models),
		materials: model.NewMaterials(assets.Atlas, assets.AtlasImage, options.Policy),
	}
	return engine, nil
}
