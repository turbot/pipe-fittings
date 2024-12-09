package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/ociinstaller"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/versionfile"
)

const (
	VersionCheckerSchema        = "https"
	VersionCheckerHostSteampipe = "hub.steampipe.io"
	// TODO Pskr Update the endpoint when hub.tailpipe.io is ready
	VersionCheckerHostTailpipe = "hub-tailpipe-io.vercel.app"
	VersionCheckerEndpoint     = "api/plugin/version"
)

// PluginVersionCheckReport
type PluginVersionCheckReport struct {
	Plugin        *versionfile.InstalledVersion
	CheckResponse versionCheckCorePayload
	CheckRequest  versionCheckCorePayload
}

func (vr *PluginVersionCheckReport) ShortName() string {
	return fmt.Sprintf("%s/%s", vr.CheckResponse.Org, vr.CheckResponse.Name)
}

func (vr *PluginVersionCheckReport) ShortNameWithConstraint() string {
	// dont show constraints for latest
	if vr.CheckResponse.Constraint != "latest" {
		return fmt.Sprintf("%s/%s@%s", vr.CheckResponse.Org, vr.CheckResponse.Name, vr.CheckResponse.Constraint)
	}
	return fmt.Sprintf("%s/%s", vr.CheckResponse.Org, vr.CheckResponse.Name)
}

// VersionChecker :: wrapper struct over the plugin version check utilities
type VersionChecker struct {
	pluginsToCheck []*versionfile.InstalledVersion
	signature      string
}

// GetUpdateReport looks up and reports the updated version of selective turbot plugins which are listed in versions.json
func GetUpdateReport(ctx context.Context, installationID string, check []*versionfile.InstalledVersion) map[string]PluginVersionCheckReport {
	versionChecker := new(VersionChecker)
	versionChecker.signature = installationID

	for _, c := range check {
		if strings.HasPrefix(c.Name, app_specific.DefaultImageRepoDisplayURL) {
			versionChecker.pluginsToCheck = append(versionChecker.pluginsToCheck, c)
		}
	}

	return versionChecker.reportPluginUpdates(ctx)
}

// GetAllUpdateReport looks up and reports the updated version of all turbot plugins which are listed in versions.json
func GetAllUpdateReport(ctx context.Context, installationID string, pluginVersions map[string]*versionfile.InstalledVersion) map[string]PluginVersionCheckReport {
	versionChecker := new(VersionChecker)
	versionChecker.signature = installationID
	versionChecker.pluginsToCheck = []*versionfile.InstalledVersion{}

	for _, p := range pluginVersions {
		if strings.HasPrefix(p.Name, app_specific.DefaultImageRepoDisplayURL) {
			versionChecker.pluginsToCheck = append(versionChecker.pluginsToCheck, p)
		}
	}

	return versionChecker.reportPluginUpdates(ctx)
}

func (v *VersionChecker) reportPluginUpdates(ctx context.Context) map[string]PluginVersionCheckReport {
	// retrieve the plugin version data from steampipe config
	versionFileData, err := versionfile.LoadPluginVersionFile(ctx)
	if err != nil {
		log.Printf("[TRACE] reportPluginUpdates could not load version file: %s", err.Error())
		return nil
	}

	if len(v.pluginsToCheck) == 0 {
		// there's no plugin installed. no point continuing
		return nil
	}
	reports := v.getLatestVersionsForPlugins(ctx, v.pluginsToCheck)

	// remove elements from `reports` which have empty strings in CheckResponse
	// this happens if we have sent a plugin to the API which doesn't exist
	// in the registry
	for key, value := range reports {
		if value.CheckResponse.Name == "" {
			// delete this key
			delete(reports, key)
		}
	}

	// update the version file
	for _, plugin := range v.pluginsToCheck {
		versionFileData.Plugins[plugin.Name].LastCheckedDate = utils.FormatTime(time.Now())
	}

	if err = versionFileData.Save(); err != nil {
		log.Printf("[WARN] reportPluginUpdates could not save version file: %s", err.Error())
		return nil
	}

	return reports
}

func (v *VersionChecker) getLatestVersionsForPlugins(ctx context.Context, plugins []*versionfile.InstalledVersion) map[string]PluginVersionCheckReport {

	var requestPayload []versionCheckCorePayload
	reports := map[string]PluginVersionCheckReport{}

	for _, ref := range plugins {
		thisPayload := v.getPayloadFromInstalledData(ref)
		requestPayload = append(requestPayload, thisPayload)

		reports[thisPayload.getMapKey()] = PluginVersionCheckReport{
			Plugin:        ref,
			CheckRequest:  thisPayload,
			CheckResponse: versionCheckCorePayload{},
		}
	}

	serverResponse, err := v.requestServerForLatest(ctx, requestPayload)
	if err != nil {
		log.Printf("[TRACE] PluginVersionChecker getLatestVersionsForPlugins returned error: %s", err.Error())
		// return a blank map
		return map[string]PluginVersionCheckReport{}
	}

	for _, pluginResponseData := range serverResponse {
		r := reports[pluginResponseData.getMapKey()]
		r.CheckResponse = pluginResponseData
		reports[pluginResponseData.getMapKey()] = r
	}

	return reports
}

func (v *VersionChecker) getPayloadFromInstalledData(plugin *versionfile.InstalledVersion) versionCheckCorePayload {
	ref := ociinstaller.NewImageRef(plugin.Name)
	org, name, constraint := ref.GetOrgNameAndStream()
	payload := versionCheckCorePayload{
		Org:        org,
		Name:       name,
		Constraint: constraint,
		Version:    plugin.Version,
	}

	return payload
}

func (v *VersionChecker) getVersionCheckURL() url.URL {
	var u url.URL
	u.Scheme = VersionCheckerSchema
	if app_specific.AppName == "tailpipe" {
		u.Host = VersionCheckerHostTailpipe
	} else {
		u.Host = VersionCheckerHostSteampipe
	}
	u.Path = VersionCheckerEndpoint
	return u
}

func (v *VersionChecker) requestServerForLatest(ctx context.Context, payload []versionCheckCorePayload) ([]versionCheckCorePayload, error) {
	// Set a default timeout of 3 sec for the check request (in milliseconds)
	sendRequestTo := v.getVersionCheckURL()
	requestBody := utils.BuildRequestPayload(v.signature, map[string]interface{}{
		"plugins": payload,
	})

	resp, err := utils.SendRequest(ctx, v.signature, "POST", sendRequestTo, requestBody)
	if err != nil {
		log.Printf("[TRACE] Could not send request")
		return nil, err
	}

	if resp.StatusCode != 200 {
		log.Printf("[TRACE] Unknown response during version check: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("requestServerForLatest failed - SendRequest returned %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[TRACE] Error reading body stream: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var responseData []versionCheckCorePayload

	err = json.Unmarshal(bodyBytes, &responseData)
	if err != nil {
		log.Println("[TRACE] Error in unmarshalling plugin update response", err)
		return nil, err
	}

	return responseData, nil
}

func GetLatestPluginVersionByConstraint(ctx context.Context, installationID string, org string, name string, constraint string) (*ResolvedPluginVersion, error) {
	vc := VersionChecker{signature: installationID}
	payload := []versionCheckCorePayload{
		{
			Org:        org,
			Name:       name,
			Constraint: constraint,
			Version:    "0.0.0", // This is used by installer, version is required by the API, this makes sense as nothing installed.
		},
	}
	orgAndName := fmt.Sprintf("%s/%s", org, name)

	vcr, err := vc.requestServerForLatest(ctx, payload)
	if err != nil {
		return nil, err
	}
	if len(vcr) == 0 {
		return nil, fmt.Errorf("no version found for %s with constraint %s", orgAndName, constraint)
	}

	v := vcr[0]
	rpv := NewResolvedPluginVersion(orgAndName, v.Version, constraint)

	return &rpv, nil
}
