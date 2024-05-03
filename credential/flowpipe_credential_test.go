package credential

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/modconfig"
)

func TestAwsCredential(t *testing.T) {

	assert := assert.New(t)

	awsCred := AwsCredential{}

	os.Setenv("AWS_ACCESS_KEY_ID", "foo")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "bar")

	newCreds, err := awsCred.Resolve(context.TODO())
	assert.Nil(err)
	assert.NotNil(newCreds)

	newAwsCreds := newCreds.(*AwsCredential)

	assert.Equal("foo", *newAwsCreds.AccessKey)
	assert.Equal("bar", *newAwsCreds.SecretKey)
	assert.Nil(newAwsCreds.SessionToken)
}

func TestAlicloudCredential(t *testing.T) {

	assert := assert.New(t)

	alicloudCred := AlicloudCredential{}

	os.Setenv("ALIBABACLOUD_ACCESS_KEY_ID", "foo")
	os.Setenv("ALIBABACLOUD_ACCESS_KEY_SECRET", "bar")

	newCreds, err := alicloudCred.Resolve(context.TODO())
	assert.Nil(err)
	assert.NotNil(newCreds)

	newAlicloudCreds := newCreds.(*AlicloudCredential)

	assert.Equal("foo", *newAlicloudCreds.AccessKey)
	assert.Equal("bar", *newAlicloudCreds.SecretKey)
}

func TestSlackDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	slackCred := SlackCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("SLACK_TOKEN")
	newCreds, err := slackCred.Resolve(context.TODO())
	assert.Nil(err)

	newSlackCreds := newCreds.(*SlackCredential)
	assert.Equal("", *newSlackCreds.Token)

	os.Setenv("SLACK_TOKEN", "foobar")

	newCreds, err = slackCred.Resolve(context.TODO())
	assert.Nil(err)

	newSlackCreds = newCreds.(*SlackCredential)
	assert.Equal("foobar", *newSlackCreds.Token)
}

func TestAbuseIPDBDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	abuseIPDBCred := AbuseIPDBCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("ABUSEIPDB_API_KEY")
	newCreds, err := abuseIPDBCred.Resolve(context.TODO())
	assert.Nil(err)

	newAbuseIPDBCreds := newCreds.(*AbuseIPDBCredential)
	assert.Equal("", *newAbuseIPDBCreds.APIKey)

	os.Setenv("ABUSEIPDB_API_KEY", "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d")

	newCreds, err = abuseIPDBCred.Resolve(context.TODO())
	assert.Nil(err)

	newAbuseIPDBCreds = newCreds.(*AbuseIPDBCredential)
	assert.Equal("bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d", *newAbuseIPDBCreds.APIKey)
}

func TestSendGridDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	sendGridCred := SendGridCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("SENDGRID_API_KEY")
	newCreds, err := sendGridCred.Resolve(context.TODO())
	assert.Nil(err)

	newSendGridCreds := newCreds.(*SendGridCredential)
	assert.Equal("", *newSendGridCreds.APIKey)

	os.Setenv("SENDGRID_API_KEY", "SGsomething")

	newCreds, err = sendGridCred.Resolve(context.TODO())
	assert.Nil(err)

	newSendGridCreds = newCreds.(*SendGridCredential)
	assert.Equal("SGsomething", *newSendGridCreds.APIKey)
}

func TestVirusTotalDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	virusTotalCred := VirusTotalCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("VTCLI_APIKEY")
	newCreds, err := virusTotalCred.Resolve(context.TODO())
	assert.Nil(err)

	newVirusTotalCreds := newCreds.(*VirusTotalCredential)
	assert.Equal("", *newVirusTotalCreds.APIKey)

	os.Setenv("VTCLI_APIKEY", "w5kukcma7yfj8m8p5rkjx5chg3nno9z7h7wr4o8uq1n0pmr5dfejox4oz4xr7g3c")

	newCreds, err = virusTotalCred.Resolve(context.TODO())
	assert.Nil(err)

	newVirusTotalCreds = newCreds.(*VirusTotalCredential)
	assert.Equal("w5kukcma7yfj8m8p5rkjx5chg3nno9z7h7wr4o8uq1n0pmr5dfejox4oz4xr7g3c", *newVirusTotalCreds.APIKey)
}

func TestZendeskDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	zendeskCred := ZendeskCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("ZENDESK_SUBDOMAIN")
	os.Unsetenv("ZENDESK_EMAIL")
	os.Unsetenv("ZENDESK_API_TOKEN")

	newCreds, err := zendeskCred.Resolve(context.TODO())
	assert.Nil(err)

	newZendeskCreds := newCreds.(*ZendeskCredential)
	assert.Equal("", *newZendeskCreds.Subdomain)
	assert.Equal("", *newZendeskCreds.Email)
	assert.Equal("", *newZendeskCreds.Token)

	os.Setenv("ZENDESK_SUBDOMAIN", "dmi")
	os.Setenv("ZENDESK_EMAIL", "pam@dmi.com")
	os.Setenv("ZENDESK_API_TOKEN", "17ImlCYdfZ3WJIrGk96gCpJn1fi1pLwVdrb23kj4")

	newCreds, err = zendeskCred.Resolve(context.TODO())
	assert.Nil(err)

	newZendeskCreds = newCreds.(*ZendeskCredential)
	assert.Equal("dmi", *newZendeskCreds.Subdomain)
	assert.Equal("pam@dmi.com", *newZendeskCreds.Email)
	assert.Equal("17ImlCYdfZ3WJIrGk96gCpJn1fi1pLwVdrb23kj4", *newZendeskCreds.Token)
}

func TestOktaDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	oktaCred := OktaCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("OKTA_CLIENT_TOKEN")
	os.Unsetenv("OKTA_ORGURL")

	newCreds, err := oktaCred.Resolve(context.TODO())
	assert.Nil(err)

	newOktaCreds := newCreds.(*OktaCredential)
	assert.Equal("", *newOktaCreds.Token)
	assert.Equal("", *newOktaCreds.Domain)

	os.Setenv("OKTA_CLIENT_TOKEN", "00B630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb")
	os.Setenv("OKTA_ORGURL", "https://dev-50078045.okta.com")

	newCreds, err = oktaCred.Resolve(context.TODO())
	assert.Nil(err)

	newOktaCreds = newCreds.(*OktaCredential)
	assert.Equal("00B630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb", *newOktaCreds.Token)
	assert.Equal("https://dev-50078045.okta.com", *newOktaCreds.Domain)
}

func TestTrelloDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	trelloCred := TrelloCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("TRELLO_API_KEY")
	os.Unsetenv("TRELLO_TOKEN")

	newCreds, err := trelloCred.Resolve(context.TODO())
	assert.Nil(err)

	newTrelloCreds := newCreds.(*TrelloCredential)
	assert.Equal("", *newTrelloCreds.APIKey)
	assert.Equal("", *newTrelloCreds.Token)

	os.Setenv("TRELLO_API_KEY", "dmgdhdfhfhfhi")
	os.Setenv("TRELLO_TOKEN", "17ImlCYdfZ3WJIrGk96gCpJn1fi1pLwVdrb23kj4")

	newCreds, err = trelloCred.Resolve(context.TODO())
	assert.Nil(err)

	newTrelloCreds = newCreds.(*TrelloCredential)
	assert.Equal("dmgdhdfhfhfhi", *newTrelloCreds.APIKey)
	assert.Equal("17ImlCYdfZ3WJIrGk96gCpJn1fi1pLwVdrb23kj4", *newTrelloCreds.Token)
}

func TestUptimeRobotDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	uptimeRobotCred := UptimeRobotCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("UPTIMEROBOT_API_KEY")
	newCreds, err := uptimeRobotCred.Resolve(context.TODO())
	assert.Nil(err)

	newUptimeRobotCreds := newCreds.(*UptimeRobotCredential)
	assert.Equal("", *newUptimeRobotCreds.APIKey)

	os.Setenv("UPTIMEROBOT_API_KEY", "u1123455-ecaf32fwer633fdf4f33dd3c445")

	newCreds, err = uptimeRobotCred.Resolve(context.TODO())
	assert.Nil(err)

	newUptimeRobotCreds = newCreds.(*UptimeRobotCredential)
	assert.Equal("u1123455-ecaf32fwer633fdf4f33dd3c445", *newUptimeRobotCreds.APIKey)
}

func TestUrlscanDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	urlscanCred := UrlscanCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("URLSCAN_API_KEY")
	newCreds, err := urlscanCred.Resolve(context.TODO())
	assert.Nil(err)

	newUrlscanCreds := newCreds.(*UrlscanCredential)
	assert.Equal("", *newUrlscanCreds.APIKey)

	os.Setenv("URLSCAN_API_KEY", "4d7e9123-e127-56c1-8d6a-59cad2f12abc")

	newCreds, err = urlscanCred.Resolve(context.TODO())
	assert.Nil(err)

	newUrlscanCreds = newCreds.(*UrlscanCredential)
	assert.Equal("4d7e9123-e127-56c1-8d6a-59cad2f12abc", *newUrlscanCreds.APIKey)
}

func TestClickUpDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	clickUpCred := ClickUpCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("CLICKUP_TOKEN")
	newCreds, err := clickUpCred.Resolve(context.TODO())
	assert.Nil(err)

	newClickUpCreds := newCreds.(*ClickUpCredential)
	assert.Equal("", *newClickUpCreds.Token)

	os.Setenv("CLICKUP_TOKEN", "pk_616_L5H36X3CXXXXXXXWEAZZF0NM5")

	newCreds, err = clickUpCred.Resolve(context.TODO())
	assert.Nil(err)

	newClickUpCreds = newCreds.(*ClickUpCredential)
	assert.Equal("pk_616_L5H36X3CXXXXXXXWEAZZF0NM5", *newClickUpCreds.Token)
}

func TestPagerDutyDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	pagerDutyCred := PagerDutyCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("PAGERDUTY_TOKEN")
	newCreds, err := pagerDutyCred.Resolve(context.TODO())
	assert.Nil(err)

	newPagerDutyCreds := newCreds.(*PagerDutyCredential)
	assert.Equal("", *newPagerDutyCreds.Token)

	os.Setenv("PAGERDUTY_TOKEN", "u+AtBdqvNtestTokeNcg")

	newCreds, err = pagerDutyCred.Resolve(context.TODO())
	assert.Nil(err)

	newPagerDutyCreds = newCreds.(*PagerDutyCredential)
	assert.Equal("u+AtBdqvNtestTokeNcg", *newPagerDutyCreds.Token)
}

func TestDiscordDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	discordCred := DiscordCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("DISCORD_TOKEN")
	newCreds, err := discordCred.Resolve(context.TODO())
	assert.Nil(err)

	newDiscordCreds := newCreds.(*DiscordCredential)
	assert.Equal("", *newDiscordCreds.Token)

	os.Setenv("DISCORD_TOKEN", "00B630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb")

	newCreds, err = discordCred.Resolve(context.TODO())
	assert.Nil(err)

	newDiscordCreds = newCreds.(*DiscordCredential)
	assert.Equal("00B630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb", *newDiscordCreds.Token)
}

func TestIP2LocationIODefaultCredential(t *testing.T) {
	assert := assert.New(t)

	ip2LocationCred := IP2LocationIOCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("IP2LOCATIONIO_API_KEY")
	newCreds, err := ip2LocationCred.Resolve(context.TODO())
	assert.Nil(err)

	newIP2LocationCreds := newCreds.(*IP2LocationIOCredential)
	assert.Equal("", *newIP2LocationCreds.APIKey)

	os.Setenv("IP2LOCATIONIO_API_KEY", "12345678901A23BC4D5E6FG78HI9J101")

	newCreds, err = ip2LocationCred.Resolve(context.TODO())
	assert.Nil(err)

	newIP2LocationCreds = newCreds.(*IP2LocationIOCredential)
	assert.Equal("12345678901A23BC4D5E6FG78HI9J101", *newIP2LocationCreds.APIKey)
}

func TestIPstackDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	ipstackCred := IPstackCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("IPSTACK_ACCESS_KEY")
	os.Unsetenv("IPSTACK_TOKEN")
	newCreds, err := ipstackCred.Resolve(context.TODO())
	assert.Nil(err)

	newIPstackCreds := newCreds.(*IPstackCredential)
	assert.Equal("", *newIPstackCreds.AccessKey)

	os.Setenv("IPSTACK_ACCESS_KEY", "1234801bfsffsdf123455e6cfaf2")
	os.Setenv("IPSTACK_TOKEN", "1234801bfsffsdf123455e6cfaf2")

	newCreds, err = ipstackCred.Resolve(context.TODO())
	assert.Nil(err)

	newIPstackCreds = newCreds.(*IPstackCredential)
	assert.Equal("1234801bfsffsdf123455e6cfaf2", *newIPstackCreds.AccessKey)
}

func TestMicrosoftTeamsDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	msTeamsCred := MicrosoftTeamsCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("TEAMS_ACCESS_TOKEN")
	newCreds, err := msTeamsCred.Resolve(context.TODO())
	assert.Nil(err)

	newMSTeamsCreds := newCreds.(*MicrosoftTeamsCredential)
	assert.Equal("", *newMSTeamsCreds.AccessToken)

	os.Setenv("TEAMS_ACCESS_TOKEN", "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d")

	newCreds, err = msTeamsCred.Resolve(context.TODO())
	assert.Nil(err)

	newMSTeamsCreds = newCreds.(*MicrosoftTeamsCredential)
	assert.Equal("bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d", *newMSTeamsCreds.AccessToken)
}

func TestGitHubDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	githubCred := GithubCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("GITHUB_TOKEN")
	newCreds, err := githubCred.Resolve(context.TODO())
	assert.Nil(err)

	newGithubAccessTokenCreds := newCreds.(*GithubCredential)
	assert.Equal("", *newGithubAccessTokenCreds.Token)

	os.Setenv("GITHUB_TOKEN", "ghpat-ljgllghhegweroyuouo67u5476070owetylh")

	newCreds, err = githubCred.Resolve(context.TODO())
	assert.Nil(err)

	newGithubAccessTokenCreds = newCreds.(*GithubCredential)
	assert.Equal("ghpat-ljgllghhegweroyuouo67u5476070owetylh", *newGithubAccessTokenCreds.Token)
}

func TestGitLabDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	gitlabCred := GitLabCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("GITLAB_TOKEN")
	newCreds, err := gitlabCred.Resolve(context.TODO())
	assert.Nil(err)

	newGitLabAccessTokenCreds := newCreds.(*GitLabCredential)
	assert.Equal("", *newGitLabAccessTokenCreds.Token)

	os.Setenv("GITLAB_TOKEN", "glpat-ljgllghhegweroyuouo67u5476070owetylh")

	newCreds, err = gitlabCred.Resolve(context.TODO())
	assert.Nil(err)

	newGitLabAccessTokenCreds = newCreds.(*GitLabCredential)
	assert.Equal("glpat-ljgllghhegweroyuouo67u5476070owetylh", *newGitLabAccessTokenCreds.Token)
}

func TestPipesDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	pipesCred := PipesCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("PIPES_TOKEN")
	newCreds, err := pipesCred.Resolve(context.TODO())
	assert.Nil(err)

	newPipesCreds := newCreds.(*PipesCredential)
	assert.Equal("", *newPipesCreds.Token)

	os.Setenv("PIPES_TOKEN", "tpt_cld630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb")

	newCreds, err = pipesCred.Resolve(context.TODO())
	assert.Nil(err)

	newPipesCreds = newCreds.(*PipesCredential)
	assert.Equal("tpt_cld630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb", *newPipesCreds.Token)
}

func TestVaultDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	vaultCred := VaultCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("VAULT_TOKEN")
	os.Unsetenv("VAULT_ADDR")

	newCreds, err := vaultCred.Resolve(context.TODO())
	assert.Nil(err)

	newVaultCreds := newCreds.(*VaultCredential)
	assert.Equal("", *newVaultCreds.Token)
	assert.Equal("", *newVaultCreds.Address)

	os.Setenv("VAULT_TOKEN", "hsv-fhhwskfkwh")
	os.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")

	newCreds, err = vaultCred.Resolve(context.TODO())
	assert.Nil(err)

	newVaultCreds = newCreds.(*VaultCredential)
	assert.Equal("hsv-fhhwskfkwh", *newVaultCreds.Token)
	assert.Equal("http://127.0.0.1:8200", *newVaultCreds.Address)
}

func TestJiraDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	jiraCred := JiraCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("JIRA_API_TOKEN")
	os.Unsetenv("JIRA_TOKEN")
	os.Unsetenv("JIRA_URL")
	os.Unsetenv("JIRA_USER")

	newCreds, err := jiraCred.Resolve(context.TODO())
	assert.Nil(err)

	newJiraCreds := newCreds.(*JiraCredential)
	assert.Equal("", *newJiraCreds.APIToken)
	assert.Equal("", *newJiraCreds.BaseURL)
	assert.Equal("", *newJiraCreds.Username)

	os.Setenv("JIRA_API_TOKEN", "ksfhashkfhakskashfghaskfagfgir327934gkegf")
	os.Setenv("JIRA_TOKEN", "ksfhashkfhakskashfghaskfagfgir327934gkegf")
	os.Setenv("JIRA_URL", "https://flowpipe-testorg.atlassian.net/")
	os.Setenv("JIRA_USER", "test@turbot.com")

	newCreds, err = jiraCred.Resolve(context.TODO())
	assert.Nil(err)

	newJiraCreds = newCreds.(*JiraCredential)
	assert.Equal("ksfhashkfhakskashfghaskfagfgir327934gkegf", *newJiraCreds.APIToken)
	assert.Equal("https://flowpipe-testorg.atlassian.net/", *newJiraCreds.BaseURL)
	assert.Equal("test@turbot.com", *newJiraCreds.Username)
}

func TestOpsgenieDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	opsgenieCred := OpsgenieCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("OPSGENIE_ALERT_API_KEY")
	os.Unsetenv("OPSGENIE_INCIDENT_API_KEY")

	newCreds, err := opsgenieCred.Resolve(context.TODO())
	assert.Nil(err)

	newOpsgenieCreds := newCreds.(*OpsgenieCredential)
	assert.Equal("", *newOpsgenieCreds.AlertAPIKey)
	assert.Equal("", *newOpsgenieCreds.IncidentAPIKey)

	os.Setenv("OPSGENIE_ALERT_API_KEY", "ksfhashkfhakskashfghaskfagfgir327934gkegf")
	os.Setenv("OPSGENIE_INCIDENT_API_KEY", "jkgdgjdgjldjgdjlgjdlgjlgjldjgldjlgjdl")

	newCreds, err = opsgenieCred.Resolve(context.TODO())
	assert.Nil(err)

	newOpsgenieCreds = newCreds.(*OpsgenieCredential)
	assert.Equal("ksfhashkfhakskashfghaskfagfgir327934gkegf", *newOpsgenieCreds.AlertAPIKey)
	assert.Equal("jkgdgjdgjldjgdjlgjdlgjlgjldjgldjlgjdl", *newOpsgenieCreds.IncidentAPIKey)
}

func TestOpenAIDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	openAICred := OpenAICredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("OPENAI_API_KEY")

	newCreds, err := openAICred.Resolve(context.TODO())
	assert.Nil(err)

	newOpenAICreds := newCreds.(*OpenAICredential)
	assert.Equal("", *newOpenAICreds.APIKey)

	os.Setenv("OPENAI_API_KEY", "sk-jwgthNa...")

	newCreds, err = openAICred.Resolve(context.TODO())
	assert.Nil(err)

	newOpenAICreds = newCreds.(*OpenAICredential)
	assert.Equal("sk-jwgthNa...", *newOpenAICreds.APIKey)
}

func TestAzureDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	azureCred := AzureCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("AZURE_CLIENT_ID")
	os.Unsetenv("AZURE_CLIENT_SECRET")
	os.Unsetenv("AZURE_TENANT_ID")
	os.Unsetenv("AZURE_ENVIRONMENT")

	newCreds, err := azureCred.Resolve(context.TODO())
	assert.Nil(err)

	newAzureCreds := newCreds.(*AzureCredential)
	assert.Equal("", *newAzureCreds.ClientID)
	assert.Equal("", *newAzureCreds.ClientSecret)
	assert.Equal("", *newAzureCreds.TenantID)
	assert.Equal("", *newAzureCreds.Environment)

	os.Setenv("AZURE_CLIENT_ID", "clienttoken")
	os.Setenv("AZURE_CLIENT_SECRET", "clientsecret")
	os.Setenv("AZURE_TENANT_ID", "tenantid")
	os.Setenv("AZURE_ENVIRONMENT", "environmentvar")

	newCreds, err = azureCred.Resolve(context.TODO())
	assert.Nil(err)

	newAzureCreds = newCreds.(*AzureCredential)
	assert.Equal("clienttoken", *newAzureCreds.ClientID)
	assert.Equal("clientsecret", *newAzureCreds.ClientSecret)
	assert.Equal("tenantid", *newAzureCreds.TenantID)
	assert.Equal("environmentvar", *newAzureCreds.Environment)
}

func TestBitbucketDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	bitbucketCred := BitbucketCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("BITBUCKET_API_BASE_URL")
	os.Unsetenv("BITBUCKET_USERNAME")
	os.Unsetenv("BITBUCKET_PASSWORD")

	newCreds, err := bitbucketCred.Resolve(context.TODO())
	assert.Nil(err)

	newBitbucketCreds := newCreds.(*BitbucketCredential)
	assert.Equal("", *newBitbucketCreds.BaseURL)
	assert.Equal("", *newBitbucketCreds.Username)
	assert.Equal("", *newBitbucketCreds.Password)

	os.Setenv("BITBUCKET_API_BASE_URL", "https://api.bitbucket.org/2.0")
	os.Setenv("BITBUCKET_USERNAME", "test@turbot.com")
	os.Setenv("BITBUCKET_PASSWORD", "ksfhashkfhakskashfghaskfagfgir327934gkegf")

	newCreds, err = bitbucketCred.Resolve(context.TODO())
	assert.Nil(err)

	newBitbucketCreds = newCreds.(*BitbucketCredential)
	assert.Equal("https://api.bitbucket.org/2.0", *newBitbucketCreds.BaseURL)
	assert.Equal("test@turbot.com", *newBitbucketCreds.Username)
	assert.Equal("ksfhashkfhakskashfghaskfagfgir327934gkegf", *newBitbucketCreds.Password)
}

func TestDatadogDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	datadogCred := DatadogCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("DD_CLIENT_API_KEY")
	os.Unsetenv("DD_CLIENT_APP_KEY")

	newCreds, err := datadogCred.Resolve(context.TODO())
	assert.Nil(err)

	newDatadogCreds := newCreds.(*DatadogCredential)
	assert.Equal("", *newDatadogCreds.APIKey)
	assert.Equal("", *newDatadogCreds.AppKey)

	os.Setenv("DD_CLIENT_API_KEY", "b1cf23432fwef23fg24grg31gr")
	os.Setenv("DD_CLIENT_APP_KEY", "1a2345bc23fwefrg13g233f")

	newCreds, err = datadogCred.Resolve(context.TODO())
	assert.Nil(err)

	newDatadogCreds = newCreds.(*DatadogCredential)
	assert.Equal("b1cf23432fwef23fg24grg31gr", *newDatadogCreds.APIKey)
	assert.Equal("1a2345bc23fwefrg13g233f", *newDatadogCreds.AppKey)
}

func TestFreshdeskDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	freshdeskCred := FreshdeskCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("FRESHDESK_API_KEY")
	os.Unsetenv("FRESHDESK_SUBDOMAIN")

	newCreds, err := freshdeskCred.Resolve(context.TODO())
	assert.Nil(err)

	newFreshdeskCreds := newCreds.(*FreshdeskCredential)
	assert.Equal("", *newFreshdeskCreds.APIKey)
	assert.Equal("", *newFreshdeskCreds.Subdomain)

	os.Setenv("FRESHDESK_API_KEY", "b1cf23432fwef23fg24grg31gr")
	os.Setenv("FRESHDESK_SUBDOMAIN", "sub-domain")

	newCreds, err = freshdeskCred.Resolve(context.TODO())
	assert.Nil(err)

	newFreshdeskCreds = newCreds.(*FreshdeskCredential)
	assert.Equal("b1cf23432fwef23fg24grg31gr", *newFreshdeskCreds.APIKey)
	assert.Equal("sub-domain", *newFreshdeskCreds.Subdomain)
}

func TestGuardrailsDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	guardrailsCred := GuardrailsCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("TURBOT_ACCESS_KEY")
	os.Unsetenv("TURBOT_SECRET_KEY")
	os.Unsetenv("TURBOT_WORKSPACE")

	newCreds, err := guardrailsCred.Resolve(context.TODO())
	assert.Nil(err)

	newGuardrailsCreds := newCreds.(*GuardrailsCredential)
	assert.Equal("", *newGuardrailsCreds.AccessKey)
	assert.Equal("", *newGuardrailsCreds.SecretKey)

	os.Setenv("TURBOT_ACCESS_KEY", "c8e2c2ed-1ca8-429b-b369-123")
	os.Setenv("TURBOT_SECRET_KEY", "a3d8385d-47f7-40c5-a90c-123")
	os.Setenv("TURBOT_WORKSPACE", "https://my_workspace.saas.turbot.com")

	newCreds, err = guardrailsCred.Resolve(context.TODO())
	assert.Nil(err)

	newGuardrailsCreds = newCreds.(*GuardrailsCredential)
	assert.Equal("c8e2c2ed-1ca8-429b-b369-123", *newGuardrailsCreds.AccessKey)
	assert.Equal("a3d8385d-47f7-40c5-a90c-123", *newGuardrailsCreds.SecretKey)
}

func TestServiceNowDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	guardrailsCred := ServiceNowCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("SERVICENOW_INSTANCE_URL")
	os.Unsetenv("SERVICENOW_USERNAME")
	os.Unsetenv("SERVICENOW_PASSWORD")

	newCreds, err := guardrailsCred.Resolve(context.TODO())
	assert.Nil(err)

	newServiceNowCreds := newCreds.(*ServiceNowCredential)
	assert.Equal("", *newServiceNowCreds.InstanceURL)
	assert.Equal("", *newServiceNowCreds.Username)
	assert.Equal("", *newServiceNowCreds.Password)

	os.Setenv("SERVICENOW_INSTANCE_URL", "https://a1b2c3d4.service-now.com")
	os.Setenv("SERVICENOW_USERNAME", "john.hill")
	os.Setenv("SERVICENOW_PASSWORD", "j0t3-$j@H3")

	newCreds, err = guardrailsCred.Resolve(context.TODO())
	assert.Nil(err)

	newServiceNowCreds = newCreds.(*ServiceNowCredential)
	assert.Equal("https://a1b2c3d4.service-now.com", *newServiceNowCreds.InstanceURL)
	assert.Equal("john.hill", *newServiceNowCreds.Username)
	assert.Equal("j0t3-$j@H3", *newServiceNowCreds.Password)
}

func TestJumpCloudDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	jumpCloudCred := JumpCloudCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
	}

	os.Unsetenv("JUMPCLOUD_API_KEY")

	newCreds, err := jumpCloudCred.Resolve(context.TODO())
	assert.Nil(err)

	newJumpCloudCreds := newCreds.(*JumpCloudCredential)
	assert.Equal("", *newJumpCloudCreds.APIKey)

	os.Setenv("JUMPCLOUD_API_KEY", "sk-jwgthNa...")

	newCreds, err = jumpCloudCred.Resolve(context.TODO())
	assert.Nil(err)

	newJumpCloudCreds = newCreds.(*JumpCloudCredential)
	assert.Equal("sk-jwgthNa...", *newJumpCloudCreds.APIKey)
}

func TestMastodonDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	// Mastodon has no standard environment variable mentioned anywhere in the docs
	accessToken := "FK2_gBrl7b9sPOSADhx61-fakezv9EDuMrXuc1AlcNU" //nolint:gosec // this is not a valid credential
	server := "https://myserver.social"

	mastodonCred := MastodonCredential{
		CredentialImpl: CredentialImpl{
			HclResourceImpl: modconfig.HclResourceImpl{
				ShortName: "default",
			},
		},
		AccessToken: &accessToken,
		Server:      &server,
	}

	newCreds, err := mastodonCred.Resolve(context.TODO())
	assert.Nil(err)

	newMastodonCreds := newCreds.(*MastodonCredential)
	assert.Equal("FK2_gBrl7b9sPOSADhx61-fakezv9EDuMrXuc1AlcNU", *newMastodonCreds.AccessToken)
	assert.Equal("https://myserver.social", *newMastodonCreds.Server)
}

func XTestAwsCredentialRole(t *testing.T) {

	assert := assert.New(t)

	awsCred := AwsCredential{}

	newCreds, err := awsCred.Resolve(context.TODO())
	assert.Nil(err)
	assert.NotNil(newCreds)

	newAwsCreds := newCreds.(*AwsCredential)

	assert.NotNil(newAwsCreds.SessionToken)
	assert.NotEqual("", newAwsCreds.SessionToken)
}
