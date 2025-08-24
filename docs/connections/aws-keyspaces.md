# AWS Keyspaces

_**DISCLAIMER**_
> As AWS Keyspaces is not real Apache Cassandra but more of emulation, many features aren't supported. That isn't something that can be fixed at the plugin's side but just not implemented at the AWS side. At the moment of Feb'2021 UDFs and IN operator aren't supported: `cassandra-backend-datasource: Error while processing a query: IN is not yet supported`

* [Setup IAM permissions](https://docs.aws.amazon.com/keyspaces/latest/devguide/accessing.html#SettingUp.IAM)
* [Generate service-specific credentials](https://docs.aws.amazon.com/keyspaces/latest/devguide/programmatic.credentials.html#programmatic.credentials.ssc)
* [Select an endpoint](https://docs.aws.amazon.com/keyspaces/latest/devguide/programmatic.endpoints.html)
* Download the certificate `curl https://certs.secureserver.net/repository/sf-class2-root.crt -O`
* Get the `sf-class2-root.crt` file to grafana
* Add it as a CA certificate in the datasource configuration
