package ociinstaller

import (
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"log"
)

type OciImage[I OciImageData, C OciImageConfig] struct {
	OCIDescriptor *ocispec.Descriptor
	ImageRef      *ImageRef
	Config        C
	Data          I
}

func FindLayersForMediaType(layers []ocispec.Descriptor, mediaType string) []ocispec.Descriptor {
	log.Println("[TRACE] looking for", mediaType)
	var matchedLayers []ocispec.Descriptor

	for _, layer := range layers {
		if layer.MediaType == mediaType {
			matchedLayers = append(matchedLayers, layer)
		}
	}
	return matchedLayers
}
