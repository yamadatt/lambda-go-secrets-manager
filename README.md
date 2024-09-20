## このリポジトリは？

SECRETS Managerに設定している以下のパスワードを変更するためのlambda。SECRETS Managerのローテーションで呼び出される。

素振り用。

- RDS（Mysql）
- SECRETS Manager

golangで実装し、コンテナ化してlambdaで動かしている。

SAMテンプレートを使用しているため、不要なAPIゲートウエイがある。これは削除してOK。

## 使ったコマンド

テストするために頻繁に使用したコマンド。

ビルドとテスト。

```bash
sam build;sam local invoke
```

テストできたらデプロイ。「本当にデプロイする？」で```yes```を選択するのが面倒なので以下のようなコマンドでオプションを使用した。

```bash
sam deploy --no-confirm-changeset --no-fail-on-empty-changeset
```

最初だけ使用したコマンド。SECRETS Managerからlambdaを実行する権限を付与できなかったため、以下のコマンドを使用した。（tempalte.yamlで権限付与までできるのが良い。これは次に課題。）

```bash
aws lambda add-permission \
  --function-name go-secrets-manager-HelloWorldFunction-LcdtCOcuQAth \
  --principal secretsmanager.amazonaws.com \
  --statement-id SecretsManagerInvocation \
  --action lambda:InvokeFunction
```

後片付けのコマンド。全て```yes```で答えてOK。

```bash
sam delete
```

## 工夫ポイント

RDSのパスワードを変更するためには```DBInstanceIdentifier```が必要。しかし、AWS SECRETS Managerから取得できないので、ホスト名から```DBInstanceIdentifier```を取得するための関数を作った。

