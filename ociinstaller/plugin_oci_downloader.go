package ociinstaller

import (
	"fmt"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/turbot/pipe-fittings/constants"
	"log"
	"strings"
)

type PluginOciDownloader struct {
	OciDownloader[*PluginImage, *PluginImageConfig]
}

func (p *PluginOciDownloader) EmptyConfig() *PluginImageConfig {
	return &PluginImageConfig{}
}

func NewPluginOciDownloader(baseImageRef string, mediaTypesProvider MediaTypeProvider) *PluginOciDownloader {
	res := &PluginOciDownloader{}

	// create the base downloader, passing res as the image provider
	ociDownloader := NewOciDownloader[*PluginImage, *PluginImageConfig](baseImageRef, mediaTypesProvider, res)

	res.OciDownloader = *ociDownloader

	return res
}

func (p *PluginOciDownloader) GetImageData(layers []ocispec.Descriptor) (*PluginImage, error) {
	res := &PluginImage{}
	var foundLayers []ocispec.Descriptor
	// get the binary plugin file info
	// iterate in order of mediatypes - as given by MediaTypeForPlatform (see function docs)
	mediaTypes, err := p.MediaTypesProvider.MediaTypeForPlatform("plugin")
	if err != nil {
		return nil, err
	}

	for _, mediaType := range mediaTypes {
		// find out the layer with the correct media type
		foundLayers = FindLayersForMediaType(layers, mediaType)
		if len(foundLayers) == 1 {
			// when found, assign and exit
			res.BinaryFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
			res.BinaryDigest = string(foundLayers[0].Digest)
			res.BinaryArchitecture = constants.ArchAMD64
			if strings.Contains(mediaType, constants.ArchARM64) {
				res.BinaryArchitecture = constants.ArchARM64
			}
			break
		}
		// loop over to the next one
		log.Println("[TRACE] could not find data for", mediaType)
		log.Println("[TRACE] falling back to the next one, if any")
	}
	if len(res.BinaryFile) == 0 {
		return nil, fmt.Errorf("invalid image - should contain 1 binary file per platform, found %d", len(foundLayers))
	}

	// try to get the docs dir

	foundLayers = FindLayersForMediaType(layers, MediaTypePluginDocsLayer())
	if len(foundLayers) > 0 {
		res.DocsDir = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	// get the .spc config / connections file dir
	foundLayers = FindLayersForMediaType(layers, MediaTypePluginSpcLayer())
	if len(foundLayers) > 0 {
		res.ConfigFileDir = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	// get the license file info
	foundLayers = FindLayersForMediaType(layers, MediaTypePluginLicenseLayer())
	if len(foundLayers) > 0 {
		res.LicenseFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	return res, nil
}
