package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type Secret struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
}

func handler(ctx context.Context, event interface{}) (interface{}, error) {
	secretArn := "arn:aws:secretsmanager:ap-northeast-1:449671225256:secret:rdspass-OdNnsx"

	// AWS SDKセッションの作成
	sess := session.Must(session.NewSession())
	sm := secretsmanager.New(sess)

	// シークレットの取得
	secretValue, err := sm.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})
	if err != nil {
		log.Fatalf("Error retrieving secret: %s", err)
	}
	fmt.Println(secretValue)
	var secret Secret
	err = json.Unmarshal([]byte(*secretValue.SecretString), &secret)
	if err != nil {
		log.Fatalf("Error unmarshalling secret: %s", err)
	}

	// 新しいパスワードを生成（ここでは単純なサンプルとして "NewPassword123!" とする）
	newPassword := "NewPassword123!"

	// RDS Admin APIを使ってパスワードを変更する
	rdsSvc := rds.New(sess)

	dbInstanceIdentifier, err := findDBInstanceIdentifier(rdsSvc, secret.Host)
	if err != nil {
		log.Fatalf("Error finding DB instance identifier: %v", err)
	}
	_, err = rdsSvc.ModifyDBInstance(&rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		MasterUserPassword:   aws.String(newPassword),
		ApplyImmediately:     aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("Error modifying DB instance: %s", err)
	}

	// シークレットを更新
	secret.Password = newPassword
	updatedSecretString, err := json.Marshal(secret)
	if err != nil {
		log.Fatalf("Error marshalling updated secret: %s", err)
	}

	_, err = sm.UpdateSecret(&secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(secretArn),
		SecretString: aws.String(string(updatedSecretString)),
	})
	if err != nil {
		log.Fatalf("Error updating secret: %s", err)
	}

	return fmt.Sprintf("Successfully rotated secret: %s", secretArn), nil
}

// findDBInstanceIdentifier関数：DBのホスト名を引数にしてDBInstanceIdentifierを取得する。
func findDBInstanceIdentifier(svc *rds.RDS, hostName string) (string, error) {
	// Describe all DB instances
	result, err := svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{})
	if err != nil {
		return "", fmt.Errorf("failed to describe DB instances: %v", err)
	}

	// Iterate through the DB instances to find the matching host name
	for _, dbInstance := range result.DBInstances {
		if strings.Contains(aws.StringValue(dbInstance.Endpoint.Address), hostName) {
			return aws.StringValue(dbInstance.DBInstanceIdentifier), nil
		}
	}

	return "", fmt.Errorf("no DB instance found with host name: %s", hostName)
}

func main() {
	lambda.Start(handler)
}
