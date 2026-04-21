package ociinstaller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/errcode"
	"oras.land/oras-go/v2/registry/remote/retry"
)

type MediaTypeProvider interface {
	MediaTypeForPlatform(imageType ImageType) ([]string, error)
	SharedMediaTypes(imageType ImageType) []string
	ConfigMediaTypes() []string
}

type ImageProvider[I OciImageData, C OciImageConfig] interface {
	GetImageData(layers []ocispec.Descriptor) (I, error)
	EmptyConfig() C
}
type OciDownloader[I OciImageData, C OciImageConfig] struct {
	resolver           remotes.Resolver
	Images             []*OciImage[I, C]
	baseImageRef       string
	MediaTypesProvider MediaTypeProvider
	ImageProvider      ImageProvider[I, C]
}

// NewOciDownloader creates and returns a OciDownloader instance
func NewOciDownloader[I OciImageData, C OciImageConfig](baseImageRef string, mediaTypesProvider MediaTypeProvider, imageProvider ImageProvider[I, C]) *OciDownloader[I, C] {
	// oras uses containerd, which uses logrus and is set up to log
	// warning and above.  Set to ErrrLevel to get rid of unwanted error message
	logrus.SetLevel(logrus.ErrorLevel)
	return &OciDownloader[I, C]{
		resolver:           docker.NewResolver(docker.ResolverOptions{}),
		MediaTypesProvider: mediaTypesProvider,
		ImageProvider:      imageProvider,
		baseImageRef:       baseImageRef,
	}
}

/*
Pull downloads the image from the given `ref` to the supplied `destDir`

Returns

	imageDescription, configDescription, config, imageLayers, error
*/
func (o *OciDownloader[I, C]) Pull(ctx context.Context, ref string, mediaTypes []string, destDir string) (*ocispec.Descriptor, *ocispec.Descriptor, []byte, []ocispec.Descriptor, error) {
	split := strings.Split(ref, ":")
	tag := split[len(split)-1]
	log.Println("[TRACE] OciDownloader.Pull:", "preparing to pull ref", ref, "tag", tag, "destDir", destDir)

	// Create the target file store
	memoryStore := memory.New()
	fileStore, err := file.NewWithFallbackStorage(destDir, memoryStore)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer fileStore.Close()

	// Connect to the remote repository
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Get credentials from the platform-default native credentials store
	storeOpts := credentials.StoreOptions{DetectDefaultNativeStore: true}
	var credStore *credentials.DynamicStore
	if strings.HasPrefix(ref, o.baseImageRef) {
		credStore, err = credentials.NewStore("", storeOpts)
	} else {
		credStore, err = credentials.NewStoreFromDocker(storeOpts)
	}
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Prepare the auth client for the registry and credential store
	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.DefaultCache,
		Credential: credentials.Credential(credStore), // Use the credential store
	}

	// Copy from the remote repository to the file store
	log.Println("[TRACE] OciDownloader.Pull:", "pulling...")

	copyOpt := oras.DefaultCopyOptions
	manifestDescriptor, err := oras.Copy(ctx, repo, tag, fileStore, tag, copyOpt)
	if err != nil && isRegistryAuthError(err) {
		// The first attempt used whatever the credential store returned for
		// the host. On a 401/403 from the registry, it's useful to know
		// whether those credentials came from the user's Docker config —
		// an expired or revoked PAT stored there produces the same opaque
		// "denied" error as a genuine permission problem, and the surfaced
		// error does not mention the credential source. Retry anonymously
		// when the store had an entry to see if the issue is stale creds.
		if retryDesc, ok := retryAnonymouslyIfStoredCreds(ctx, repo, credStore, ref, tag, fileStore, copyOpt); ok {
			manifestDescriptor = retryDesc
			err = nil
		} else {
			err = wrapAuthError(err, ref)
		}
	}
	if err != nil {
		log.Println("[TRACE] OciDownloader.Pull:", "failed to pull", ref, err)
		return nil, nil, nil, nil, err
	}
	log.Println("[TRACE] OciDownloader.Pull:", "manifest", manifestDescriptor.Digest, manifestDescriptor.MediaType)

	// FIXME: this seems redundant as oras.Copy() already downloads all artifacts, but that's the only I found
	// to access the manifest config. Also, it shouldn't be an issue as files are not re-downloaded.
	manifestJson, err := content.FetchAll(ctx, fileStore, manifestDescriptor)
	if err != nil {
		log.Println("[TRACE] OciDownloader.Pull:", "failed to fetch manifest", manifestDescriptor)
		return nil, nil, nil, nil, err
	}
	log.Println("[TRACE] OciDownloader.Pull:", "manifest content", string(manifestJson))

	// Parse the fetched manifest
	var manifest ocispec.Manifest
	err = json.Unmarshal(manifestJson, &manifest)
	if err != nil {
		log.Println("[TRACE] OciDownloader.Pull:", "failed to unmarshall manifest", manifestJson)
		return nil, nil, nil, nil, err
	}

	// Fetch the config from the file store
	configData, err := content.FetchAll(ctx, fileStore, manifest.Config)
	if err != nil {
		log.Println("[TRACE] OciDownloader.Pull:", "failed to fetch config", manifest.Config.MediaType, err)
		return nil, nil, nil, nil, err
	}
	log.Println("[TRACE] OciDownloader.Pull:", "config", string(configData))

	return &manifestDescriptor, &manifest.Config, configData, manifest.Layers, err
}

func (o *OciDownloader[I, C]) Download(ctx context.Context, ref *ImageRef, imageType ImageType, destDir string) (*OciImage[I, C], error) {
	image := o.newOciImage()
	image.ImageRef = ref
	// get the media types, including common and config
	platformMediaTypes, err := o.MediaTypesProvider.MediaTypeForPlatform(imageType)
	if err != nil {
		return nil, err
	}
	commonMediaTypes := o.MediaTypesProvider.SharedMediaTypes(imageType)
	configMediaTypes := o.MediaTypesProvider.ConfigMediaTypes()
	allMediaTypes := append(append(platformMediaTypes, commonMediaTypes...), configMediaTypes...)

	log.Println("[TRACE] OciDownloader.Download:", "downloading", ref.ActualImageRef())

	// Download the files
	imageDesc, _, configBytes, layers, err := o.Pull(ctx, ref.ActualImageRef(), allMediaTypes, destDir)
	if err != nil {
		return nil, err
	}

	image.OCIDescriptor = imageDesc

	// unmarshal the config
	emptyConfig := o.ImageProvider.EmptyConfig()
	if err := json.Unmarshal(configBytes, emptyConfig); err != nil {
		return nil, err
	}
	image.Config = emptyConfig

	if image.Config.GetSchemaVersion() == "" {
		image.Config.SetSchemaVersion(DefaultConfigSchema)
	}

	image.Data, err = o.ImageProvider.GetImageData(layers)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (o *OciDownloader[I, C]) newOciImage() *OciImage[I, C] {
	i := &OciImage[I, C]{
		resolver: &o.resolver,
	}
	o.Images = append(o.Images, i)
	return i
}

// isRegistryAuthError reports whether err unwraps to a 401 or 403 response
// from the OCI registry. Used to decide whether a retry without credentials
// could recover the pull.
func isRegistryAuthError(err error) bool {
	var errResp *errcode.ErrorResponse
	if errors.As(err, &errResp) {
		return errResp.StatusCode == http.StatusUnauthorized ||
			errResp.StatusCode == http.StatusForbidden
	}
	return false
}

// registryHostFromRef returns the registry host portion of an OCI image
// reference (e.g. "ghcr.io" from "ghcr.io/turbot/steampipe/plugins/turbot/aws:1.30.2").
// Falls back to returning the ref unchanged if it contains no path separator.
func registryHostFromRef(ref string) string {
	if i := strings.Index(ref, "/"); i >= 0 {
		return ref[:i]
	}
	return ref
}

// retryAnonymouslyIfStoredCreds retries the copy without any credentials when
// the credential store had an entry for the target host. The typical case:
// the user has an expired or revoked PAT in ~/.docker/config.json (or the
// native keychain), and the registry rejects it with DENIED even though the
// image is publicly pullable.
//
// Returns the fresh manifest descriptor and true if the anonymous retry
// succeeded. Returns the zero descriptor and false if there were no stored
// creds (so retry would change nothing) or the anonymous retry itself failed.
func retryAnonymouslyIfStoredCreds(
	ctx context.Context,
	repo *remote.Repository,
	credStore *credentials.DynamicStore,
	ref string,
	tag string,
	fileStore *file.Store,
	copyOpt oras.CopyOptions,
) (ocispec.Descriptor, bool) {
	host := registryHostFromRef(ref)
	cred, credErr := credStore.Get(ctx, host)
	if credErr != nil || cred == auth.EmptyCredential {
		// No stored creds for this host — the original attempt was already
		// anonymous, so retrying won't change anything.
		return ocispec.Descriptor{}, false
	}

	// Swap in an auth client with no credential source and retry.
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
	}
	desc, err := oras.Copy(ctx, repo, tag, fileStore, tag, copyOpt)
	if err != nil {
		log.Println("[TRACE] OciDownloader.Pull:", "anonymous retry also failed for", ref, err)
		return ocispec.Descriptor{}, false
	}

	log.Printf("[WARN] Stored credentials for %s were rejected by the registry; pull succeeded anonymously. "+
		"Clear the stale entry with: docker logout %s\n", host, host)
	return desc, true
}

// wrapAuthError adds a hint about stored credentials to an auth-failure error
// surfaced from the registry. Used when the original pull failed on auth AND
// anonymous retry was either skipped or also failed.
func wrapAuthError(err error, ref string) error {
	host := registryHostFromRef(ref)
	return fmt.Errorf("%w (if you have stored credentials for %s, they may be expired or revoked — try: docker logout %s)", err, host, host)
}
