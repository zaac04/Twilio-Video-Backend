package iam

import (
	"context"
	"fmt"

	AppConfig "stargazer/video-recording/config"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func HasPermissions() error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	stc := sts.NewFromConfig(cfg)
	out, err := stc.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})

	if err != nil {
		return fmt.Errorf("error calling sts: %s", err.Error())
	}

	svc := iam.NewFromConfig(cfg)
	actions := []string{"s3:CopyObject", "s3:HeadObject"}

	resourceArns := []string{
		fmt.Sprintf("arn:aws:s3:::%s", AppConfig.App.S3_BUCKET_NAME),
		fmt.Sprintf("arn:aws:s3:::%s/*", AppConfig.App.S3_BUCKET_NAME),
	}

	result, err := svc.SimulatePrincipalPolicy(context.TODO(), &iam.SimulatePrincipalPolicyInput{
		PolicySourceArn: out.Arn,
		ActionNames:     actions,
		ResourceArns:    resourceArns,
	})

	if err != nil {
		return fmt.Errorf("failed to simulate policy: %v", err)
	}

	for _, evaluation := range result.EvaluationResults {
		if evaluation.EvalDecision != "allowed" {
			return fmt.Errorf("permission denied for action: %s", *evaluation.EvalActionName)
		}
	}

	fmt.Printf("Provided AWS Credentials with arn %s have the required permissions.\n", *out.Arn)
	return nil
}
