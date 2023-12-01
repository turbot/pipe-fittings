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

	newCreds, err := slackCred.Resolve(context.TODO())
	assert.Nil(err)

	newSlackCreds := newCreds.(*SlackCredential)
	assert.Nil(newSlackCreds.Token)

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

	newCreds, err := abuseIPDBCred.Resolve(context.TODO())
	assert.Nil(err)

	newAbuseIPDBCreds := newCreds.(*AbuseIPDBCredential)
	assert.Nil(newAbuseIPDBCreds.APIKey)

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

	newCreds, err := sendGridCred.Resolve(context.TODO())
	assert.Nil(err)

	newSendGridCreds := newCreds.(*SendGridCredential)
	assert.Nil(newSendGridCreds.APIKey)

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

	newCreds, err := virusTotalCred.Resolve(context.TODO())
	assert.Nil(err)

	newVirusTotalCreds := newCreds.(*VirusTotalCredential)
	assert.Nil(newVirusTotalCreds.APIKey)

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

	newCreds, err := zendeskCred.Resolve(context.TODO())
	assert.Nil(err)

	newZendeskCreds := newCreds.(*ZendeskCredential)
	assert.Nil(newZendeskCreds.Subdomain)
	assert.Nil(newZendeskCreds.Email)
	assert.Nil(newZendeskCreds.Token)

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

	newCreds, err := oktaCred.Resolve(context.TODO())
	assert.Nil(err)

	newOktaCreds := newCreds.(*OktaCredential)
	assert.Nil(newOktaCreds.APIToken)
	assert.Nil(newOktaCreds.Domain)

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

	newCreds, err := trelloCred.Resolve(context.TODO())
	assert.Nil(err)

	newTrelloCreds := newCreds.(*TrelloCredential)
	assert.Nil(newTrelloCreds.APIKey)
	assert.Nil(newTrelloCreds.Token)

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

	newCreds, err := uptimeRobotCred.Resolve(context.TODO())
	assert.Nil(err)

	newUptimeRobotCreds := newCreds.(*UptimeRobotCredential)
	assert.Nil(newUptimeRobotCreds.APIKey)

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

	newCreds, err := urlscanCred.Resolve(context.TODO())
	assert.Nil(err)

	newUrlscanCreds := newCreds.(*UrlscanCredential)
	assert.Nil(newUrlscanCreds.APIKey)

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

	newCreds, err := clickUpCred.Resolve(context.TODO())
	assert.Nil(err)

	newClickUpCreds := newCreds.(*ClickUpCredential)
	assert.Nil(newClickUpCreds.APIToken)

	os.Setenv("CLICKUP_TOKEN", "pk_616_L5H36X3CXXXXXXXWEAZZF0NM5")

	newCreds, err = clickUpCred.Resolve(context.TODO())
	assert.Nil(err)

	newClickUpCreds = newCreds.(*ClickUpCredential)
	assert.Equal("pk_616_L5H36X3CXXXXXXXWEAZZF0NM5", *newClickUpCreds.APIToken)
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
