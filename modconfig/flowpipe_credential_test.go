package modconfig

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestSlackDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	slackCred := SlackCredential{
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
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
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
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
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
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
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
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
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("ZENDESK_SUBDOMAIN")
	os.Unsetenv("ZENDESK_USER")
	os.Unsetenv("ZENDESK_TOKEN")

	newCreds, err := zendeskCred.Resolve(context.TODO())
	assert.Nil(err)

	newZendeskCreds := newCreds.(*ZendeskCredential)
	assert.Equal("", *newZendeskCreds.Subdomain)
	assert.Equal("", *newZendeskCreds.Email)
	assert.Equal("", *newZendeskCreds.Token)

	os.Setenv("ZENDESK_SUBDOMAIN", "dmi")
	os.Setenv("ZENDESK_USER", "pam@dmi.com")
	os.Setenv("ZENDESK_TOKEN", "17ImlCYdfZ3WJIrGk96gCpJn1fi1pLwVdrb23kj4")

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
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("OKTA_TOKEN")
	os.Unsetenv("OKTA_ORGURL")

	newCreds, err := oktaCred.Resolve(context.TODO())
	assert.Nil(err)

	newOktaCreds := newCreds.(*OktaCredential)
	assert.Equal("", *newOktaCreds.APIToken)
	assert.Equal("", *newOktaCreds.Domain)

	os.Setenv("OKTA_TOKEN", "00B630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb")
	os.Setenv("OKTA_ORGURL", "https://dev-50078045.okta.com")

	newCreds, err = oktaCred.Resolve(context.TODO())
	assert.Nil(err)

	newOktaCreds = newCreds.(*OktaCredential)
	assert.Equal("00B630jSCGU4jV4o5Yh4KQMAdqizwE2OgVcS7N9UHb", *newOktaCreds.APIToken)
	assert.Equal("https://dev-50078045.okta.com", *newOktaCreds.Domain)
}

func TestTrelloDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	trelloCred := TrelloCredential{
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
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
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
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
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
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
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
		},
	}

	os.Unsetenv("CLICKUP_TOKEN")
	newCreds, err := clickUpCred.Resolve(context.TODO())
	assert.Nil(err)

	newClickUpCreds := newCreds.(*ClickUpCredential)
	assert.Equal("", *newClickUpCreds.APIToken)

	os.Setenv("CLICKUP_TOKEN", "pk_616_L5H36X3CXXXXXXXWEAZZF0NM5")

	newCreds, err = clickUpCred.Resolve(context.TODO())
	assert.Nil(err)

	newClickUpCreds = newCreds.(*ClickUpCredential)
	assert.Equal("pk_616_L5H36X3CXXXXXXXWEAZZF0NM5", *newClickUpCreds.APIToken)
}

func TestPagerDutyDefaultCredential(t *testing.T) {
	assert := assert.New(t)

	pagerDutyCred := PagerDutyCredential{
		HclResourceImpl: HclResourceImpl{
			ShortName: "default",
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
