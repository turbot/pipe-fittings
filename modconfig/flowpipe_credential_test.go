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
