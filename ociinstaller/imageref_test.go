package ociinstaller

import (
	"testing"
)

func TestFriendlyImageRef(t *testing.T) {
	cases := map[string]string{
		"hub.steampipe.io/plugins/turbot/aws@latest":   "aws",
		"turbot/aws@latest":                            "aws",
		"aws@latest":                                   "aws",
		"hub.steampipe.io/plugins/turbot/aws@1.0.0":    "aws@1.0.0",
		"hub.steampipe.io/plugins/otherOrg/aws@latest": "otherOrg/aws",
		"otherOrg/aws@latest":                          "otherOrg/aws",
		"hub.steampipe.io/plugins/otherOrg/aws@1.0.0":  "otherOrg/aws@1.0.0",
		"otherOrg/aws@1.0.0":                           "otherOrg/aws@1.0.0",
		"differentRegistry.com/otherOrg/aws@latest":    "differentRegistry.com/otherOrg/aws@latest",
		"differentRegistry.com/otherOrg/aws@1.0.0":     "differentRegistry.com/otherOrg/aws@1.0.0",
	}

	for testCase, want := range cases {
		t.Run(testCase, func(t *testing.T) {
			r := NewSteampipeImageRef(testCase)

			if got := r.GetFriendlyName(); got != want {
				t.Errorf("TestFriendlyImageRef failed for case '%s': expected %s, got %s", testCase, want, got)
			}
		})
	}

}

func TestActualImageRef(t *testing.T) {
	cases := map[string]string{
		"us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0":                                                                   "us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0",
		"us-docker.pkg.dev/steampipe/plugin/turbot/aws@sha256:766389c9dd892132c7e7b9124f446b9599a80863d466cd1d333a167dedf2c2b1": "us-docker.pkg.dev/steampipe/plugin/turbot/aws@sha256:766389c9dd892132c7e7b9124f446b9599a80863d466cd1d333a167dedf2c2b1",
		"aws":                                 "us-docker.pkg.dev/steampipe/plugins/turbot/aws:latest",
		"aws:1":                               "us-docker.pkg.dev/steampipe/plugins/turbot/aws:1",
		"turbot/aws:1":                        "us-docker.pkg.dev/steampipe/plugins/turbot/aws:1",
		"turbot/aws:1.0":                      "us-docker.pkg.dev/steampipe/plugins/turbot/aws:1.0",
		"turbot/aws:1.1.1":                    "us-docker.pkg.dev/steampipe/plugins/turbot/aws:1.1.1",
		"turbot/aws":                          "us-docker.pkg.dev/steampipe/plugins/turbot/aws:latest",
		"mycompany/my-plugin":                 "us-docker.pkg.dev/steampipe/plugins/mycompany/my-plugin:latest",
		"mycompany/my-plugin:some-random_tag": "us-docker.pkg.dev/steampipe/plugins/mycompany/my-plugin:some-random_tag",
		"dockerhub.org/myimage:mytag":         "dockerhub.org/myimage:mytag",
		"us-docker.pkg.dev/steampipe/plugins/turbot/aws:latest": "us-docker.pkg.dev/steampipe/plugins/turbot/aws:latest",
		"hub.steampipe.io/plugins/turbot/aws:latest":            "us-docker.pkg.dev/steampipe/plugins/turbot/aws:latest",
		"hub.steampipe.io/plugins/someoneelse/myimage:mytag":    "us-docker.pkg.dev/steampipe/plugins/someoneelse/myimage:mytag",

		"us-docker.pkg.dev/steampipe/plugin/turbot/aws@1.0.0": "us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0",
		"aws@1":                               "us-docker.pkg.dev/steampipe/plugins/turbot/aws:1",
		"turbot/aws@1":                        "us-docker.pkg.dev/steampipe/plugins/turbot/aws:1",
		"turbot/aws@1.0":                      "us-docker.pkg.dev/steampipe/plugins/turbot/aws:1.0",
		"turbot/aws@1.1.1":                    "us-docker.pkg.dev/steampipe/plugins/turbot/aws:1.1.1",
		"mycompany/my-plugin@some-random_tag": "us-docker.pkg.dev/steampipe/plugins/mycompany/my-plugin:some-random_tag",
		"dockerhub.org/myimage@mytag":         "dockerhub.org/myimage:mytag",
		"us-docker.pkg.dev/steampipe/plugins/turbot/aws@latest": "us-docker.pkg.dev/steampipe/plugins/turbot/aws:latest",
		"hub.steampipe.io/plugins/turbot/aws@latest":            "us-docker.pkg.dev/steampipe/plugins/turbot/aws:latest",
		"hub.steampipe.io/plugins/someoneelse/myimage@mytag":    "us-docker.pkg.dev/steampipe/plugins/someoneelse/myimage:mytag",
	}

	for testCase, want := range cases {

		t.Run(testCase, func(t *testing.T) {
			r := NewSteampipeImageRef(testCase)

			if got := r.ActualImageRef(); got != want {
				t.Errorf("ActualImageRef failed for case '%s': expected %s, got %s", testCase, want, got)
			}
		})
	}

}

func TestDisplayImageRef(t *testing.T) {
	cases := map[string]string{
		"us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0":                                                                   "hub.steampipe.io/plugin/turbot/aws@1.0.0",
		"us-docker.pkg.dev/steampipe/plugin/turbot/aws@sha256:766389c9dd892132c7e7b9124f446b9599a80863d466cd1d333a167dedf2c2b1": "hub.steampipe.io/plugin/turbot/aws@sha256-766389c9dd892132c7e7b9124f446b9599a80863d466cd1d333a167dedf2c2b1",
		"aws":                                 "hub.steampipe.io/plugins/turbot/aws@latest",
		"aws:1":                               "hub.steampipe.io/plugins/turbot/aws@1",
		"turbot/aws:1":                        "hub.steampipe.io/plugins/turbot/aws@1",
		"turbot/aws:1.0":                      "hub.steampipe.io/plugins/turbot/aws@1.0",
		"turbot/aws:1.1.1":                    "hub.steampipe.io/plugins/turbot/aws@1.1.1",
		"turbot/aws":                          "hub.steampipe.io/plugins/turbot/aws@latest",
		"mycompany/my-plugin":                 "hub.steampipe.io/plugins/mycompany/my-plugin@latest",
		"mycompany/my-plugin:some-random_tag": "hub.steampipe.io/plugins/mycompany/my-plugin@some-random_tag",
		"dockerhub.org/myimage:mytag":         "dockerhub.org/myimage@mytag",
		"us-docker.pkg.dev/steampipe/plugins/turbot/aws:latest": "hub.steampipe.io/plugins/turbot/aws@latest",
		"hub.steampipe.io/plugins/turbot/aws:latest":            "hub.steampipe.io/plugins/turbot/aws@latest",
		"hub.steampipe.io/plugins/someoneelse/myimage:mytag":    "hub.steampipe.io/plugins/someoneelse/myimage@mytag",

		"us-docker.pkg.dev/steampipe/plugin/turbot/aws@1.0.0": "hub.steampipe.io/plugin/turbot/aws@1.0.0",
		"aws@1":                               "hub.steampipe.io/plugins/turbot/aws@1",
		"turbot/aws@1":                        "hub.steampipe.io/plugins/turbot/aws@1",
		"turbot/aws@1.0":                      "hub.steampipe.io/plugins/turbot/aws@1.0",
		"turbot/aws@1.1.1":                    "hub.steampipe.io/plugins/turbot/aws@1.1.1",
		"mycompany/my-plugin@some-random_tag": "hub.steampipe.io/plugins/mycompany/my-plugin@some-random_tag",
		"dockerhub.org/myimage@mytag":         "dockerhub.org/myimage@mytag",
		"us-docker.pkg.dev/steampipe/plugins/turbot/aws@latest": "hub.steampipe.io/plugins/turbot/aws@latest",
		"hub.steampipe.io/plugins/turbot/aws@latest":            "hub.steampipe.io/plugins/turbot/aws@latest",
		"hub.steampipe.io/plugins/someoneelse/myimage@mytag":    "hub.steampipe.io/plugins/someoneelse/myimage@mytag",

		"aws@v1":            "hub.steampipe.io/plugins/turbot/aws@1",
		"turbot/aws@v1":     "hub.steampipe.io/plugins/turbot/aws@1",
		"turbot/aws@v1.0":   "hub.steampipe.io/plugins/turbot/aws@1.0",
		"turbot/aws@v1.1.1": "hub.steampipe.io/plugins/turbot/aws@1.1.1",
	}

	for testCase, want := range cases {

		t.Run(testCase, func(t *testing.T) {
			r := NewSteampipeImageRef(testCase)

			if got := r.DisplayImageRef(); got != want {
				t.Errorf("DisplayImageRef failed for case '%s': expected %s, got %s", testCase, want, got)
			}
		})
	}

}

func TestIsFromSteampipeHub(t *testing.T) {
	cases := map[string]bool{
		"hub.steampipe.io/plugins/turbot/aws@latest":   true,
		"turbot/aws@latest":                            true,
		"aws@latest":                                   true,
		"hub.steampipe.io/plugins/turbot/aws@1.0.0":    true,
		"hub.steampipe.io/plugins/otherOrg/aws@latest": true,
		"otherOrg/aws@latest":                          true,
		"hub.steampipe.io/plugins/otherOrg/aws@1.0.0":  true,
		"otherOrg/aws@1.0.0":                           true,
		"example.com/otherOrg/aws@latest":              false,
		"example.com/otherOrg/aws@1.0.0":               false,
	}

	for testCase, want := range cases {
		t.Run(testCase, func(t *testing.T) {
			r := NewSteampipeImageRef(testCase)
			got := r.IsFromSteampipeHub()
			if got != want {
				t.Errorf("TestIsFromSteampipeHub failed for case '%s': expected %t, got %t", testCase, want, got)
			}
		})
	}

}
