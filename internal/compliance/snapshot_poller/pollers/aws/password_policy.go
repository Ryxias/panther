package aws

/**
 * Panther is a scalable, powerful, cloud-native SIEM written in Golang/React.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"go.uber.org/zap"

	apimodels "github.com/panther-labs/panther/api/gateway/resources/models"
	awsmodels "github.com/panther-labs/panther/internal/compliance/snapshot_poller/models/aws"
	pollermodels "github.com/panther-labs/panther/internal/compliance/snapshot_poller/models/poller"
	"github.com/panther-labs/panther/internal/compliance/snapshot_poller/pollers/utils"
)

// Set as variables to be overridden in testing
var (
	IAMClientFunc = setupIAMClient
)

func setupIAMClient(sess *session.Session, cfg *aws.Config) interface{} {
	cfg.MaxRetries = aws.Int(MaxRetries)
	return iam.New(sess, cfg)
}

// PollPasswordPolicyResource polls a password policy and returns it as a resource
func PollPasswordPolicyResource(
	pollerResourceInput *awsmodels.ResourcePollerInput,
	_ *utils.ParsedResourceID,
	_ *pollermodels.ScanEntry,
) interface{} {

	snapshot, err := PollPasswordPolicy(pollerResourceInput)
	if err != nil || snapshot == nil {
		return nil
	}
	return snapshot[0].Attributes
}

// getPasswordPolicy returns the password policy for the account
func getPasswordPolicy(svc iamiface.IAMAPI) (*iam.PasswordPolicy, error) {
	out, err := svc.GetAccountPasswordPolicy(&iam.GetAccountPasswordPolicyInput{})
	if err != nil {
		return nil, err
	}

	return out.PasswordPolicy, nil
}

// PollPasswordPolicy gathers information on all PasswordPolicy in an AWS account.
func PollPasswordPolicy(pollerInput *awsmodels.ResourcePollerInput) ([]*apimodels.AddResourceEntry, error) {
	zap.L().Debug("starting Password Policy resource poller")
	sess := session.Must(session.NewSession(&aws.Config{}))
	creds, err := AssumeRoleFunc(pollerInput, sess)
	if err != nil {
		return nil, err
	}

	iamSvc := IAMClientFunc(sess, &aws.Config{Credentials: creds}).(iamiface.IAMAPI)

	anyExist := true
	passwordPolicy, getErr := getPasswordPolicy(iamSvc)
	if getErr != nil {
		if awsErr, ok := getErr.(awserr.Error); ok {
			switch awsErr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				anyExist = false
			default:
				utils.LogAWSError("IAM.GetPasswordPolicy", err)
			}
		}
	}

	resourceID := utils.GenerateResourceID(
		pollerInput.AuthSourceParsedARN.AccountID,
		"",
		awsmodels.PasswordPolicySchema,
	)

	genericFields := awsmodels.GenericResource{
		ResourceID:   aws.String(resourceID),
		ResourceType: aws.String(awsmodels.PasswordPolicySchema),
	}
	genericAWSFields := awsmodels.GenericAWSResource{
		AccountID: aws.String(pollerInput.AuthSourceParsedARN.AccountID),
		Name:      aws.String(awsmodels.PasswordPolicySchema),
		Region:    aws.String(awsmodels.GlobalRegion),
	}

	if anyExist && passwordPolicy != nil {
		return []*apimodels.AddResourceEntry{{
			Attributes: &awsmodels.PasswordPolicy{
				GenericResource:    genericFields,
				GenericAWSResource: genericAWSFields,
				AnyExist:           anyExist,
				PasswordPolicy:     *passwordPolicy,
			},
			ID:              apimodels.ResourceID(resourceID),
			IntegrationID:   apimodels.IntegrationID(*pollerInput.IntegrationID),
			IntegrationType: apimodels.IntegrationTypeAws,
			Type:            awsmodels.PasswordPolicySchema,
		}}, nil
	}

	return []*apimodels.AddResourceEntry{{
		Attributes: &awsmodels.PasswordPolicy{
			GenericResource:    genericFields,
			GenericAWSResource: genericAWSFields,
			AnyExist:           anyExist,
		},
		ID:              apimodels.ResourceID(resourceID),
		IntegrationID:   apimodels.IntegrationID(*pollerInput.IntegrationID),
		IntegrationType: apimodels.IntegrationTypeAws,
		Type:            awsmodels.PasswordPolicySchema,
	}}, nil
}
