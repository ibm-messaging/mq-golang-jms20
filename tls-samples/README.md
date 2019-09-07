# Running the TLS tests
The TLS connection tests found in [tls_connections_test.go](../tls_connections_test.go)
demonstrate how to configure two different patterns of TLS;
- Anonymous ("one way") TLS to encrypt traffic between the client application
and the queue manager
- Mutual TLS authentication, where the client application also authenticates itself
to the queue manager by providing a client certificate

In order to successfully run the TLS examples you must first apply some specific
configuration to your queue manager using the following three steps;


## Step 1: Define the queue manager channels
The tests assume that two new channels have been created on the queue manager - one
to demonstrate anonymous TLS, and the other for mutual TLS.

You can define the necessary channels by executing the following runmqsc commands
against your queue manager. Note that if your queue manager is configured to block
all access by default (for example a queue manager hosted by the [IBM MQ on Cloud service](https://cloud.ibm.com/catalog/services/mq) or using the developer configuration for the [public MQ Docker container](https://github.com/ibm-messaging/mq-container/blob/6c72c894f775752a8ec8944d73dde4264711f0ff/incubating/mqadvanced-server-dev/10-dev.mqsc.tpl#L42))
then you will also need to execute the `SET CHLAUTH` command associated with each
channel, otherwise you will receive a `MQRC_NOT_AUTHORIZED` after the TLS
connection has been established.
```
DEFINE CHANNEL(TLS.ANON.SVRCONN) CHLTYPE(SVRCONN) SSLCIPH(ANY_TLS12) SSLCAUTH(OPTIONAL) REPLACE
SET CHLAUTH(TLS.ANON.SVRCONN) TYPE(ADDRESSMAP) ADDRESS('*') USERSRC(CHANNEL) CHCKCLNT(REQUIRED) DESCR('Allow applications to connect') ACTION(REPLACE)

DEFINE CHANNEL(TLS.MUTUAL.SVRCONN) CHLTYPE(SVRCONN) SSLCIPH(ANY_TLS12) SSLCAUTH(REQUIRED) REPLACE
SET CHLAUTH(TLS.MUTUAL.SVRCONN) TYPE(ADDRESSMAP) ADDRESS('*') USERSRC(CHANNEL) CHCKCLNT(REQUIRED) DESCR('Allow applications to connect') ACTION(REPLACE)
```


## Step 2: Import the queue manager certificate into the client key store
The [anon-tls.kdb](anon-tls.kdb) and [mutual-tls.kdb](mutual-tls.kdb) files are
keystores that contain certificates that are used when the test applications
connect to the queue manager.

The keystores have been configured to contain the public certificate for the DigiCert Root
certificate authority (CA) so that the tests will automatically work against any
queue manager that presents a certificate issued by DigiCert (such as queue
managers hosted by the IBM MQ on Cloud service), however if your queue manager
uses a certificate issued by a different certificate authority then you will need
to import the public part of that certificate into both of the keystores in order
for the tests to pass.

To do this, obtain a copy of the public certificate for your queue manager in PEM
format, and import it into each of the keystores as follows;
```
gsk8capicmd -cert -add -db anon-tls.kdb -file MyQMgrCert.pem -label MyQMgrCert -stashed -type kdb -format ascii
gsk8capicmd -cert -add -db mutual-tls.kdb -file MyQMgrCert.pem -label MyQMgrCert -stashed -type kdb -format ascii
```
Note that the `gsk8capicmd` command is included in the [IBM MQ redistributable client installation](https://ibm.biz/mq91cdredistclients)
in the `./gskit8/bin` directory, or if you have a local queue manager installation
then you can substitute that command for the `ikeycmd` command instead.


## Step 3: Import the client certificate into the queue manager key store
For the mutual TLS scenario you must also configure the queue manager so that it
trusts the certificate that will be presented by the client application. The client
will present the certificate with the label `SampleClientA` in [mutual-tls.kdb](mutual-tls.kdb),
but for convenience the same certificate is also included in this directory
as [SampleClientA.pem](SampleClientA.pem).

The technique for importing the client certificate into the queue manager keystore
depends on the type of queue manager you are using, for example;
- For IBM MQ on Cloud queue managers you can import the certificate into the Trust
Store as [described here](https://cloud.ibm.com/docs/services/mqcloud?topic=mqcloud-mqoc_jms_tls#twoway_mqoc_jms_tls)
- For software installed MQ queue managers on Linux, use the `runmqckm` command [as described here](https://www.ibm.com/support/knowledgecenter/en/SSFKSJ_9.1.0/com.ibm.mq.sec.doc/q012830_.htm)

Note that in both cases you must refresh the queue manager security configuration
in order for the new certificate to take effect - for example by executing the
runmqsc command `REFRESH SECURITY TYPE(SSL)` or equivalent.


## Reference: Creating your own key store files
The [anon-tls.kdb](anon-tls.kdb) and [mutual-tls.kdb](mutual-tls.kdb) files in
this directory have been pre-defined for use with the TLS tests defined in this
repository, however you will create equivalent files for use with your own
applications.

The following commands demonstrate how the pre-defined files were created using
the `gsk8capicmd` command. As noted above this command is supplied with the IBM
MQ redistributable client installation, but if you have a local queue manager
installation then you can substitute the `ikeycmd` command instead.

```
# Create the anon-tls key store
gsk8capicmd -keydb -create -db anon-tls.kdb -pw myKeystorePW -type kdb -expire 0 -stash

# Import the DigiCert Root CA certificate so that queue managers with certificates
# issued by that CA will be trusted by default
gsk8capicmd -cert -add -db anon-tls.kdb -file DigiCertRootCA.pem -label DigiCertRootCA -stashed -type kdb -format ascii

# Create the mutual-tls key store
gsk8capicmd -keydb -create -db mutual-tls.kdb -pw myMutualKeystorePW -type kdb -expire 0 -stash

# Import the DigiCert Root CA certificate so that queue managers with certificates
# issued by that CA will be trusted by default
gsk8capicmd -cert -add -db mutual-tls.kdb -file DigiCertRootCA.pem -label DigiCertRootCA -stashed -type kdb -format ascii

# Create a new self-signed private key and certificate inside the mutual-tls keystore.
# The client will present this certificate to identify itself to the queue manager
gsk8capicmd -cert -create -db mutual-tls.kdb -pw myMutualKeystorePW -sig_alg SHA256WithRSA -expire 3650 -label SampleClientA -dn "O=Sample Client Corporation, C=UK"

# Extract the public part of the self-signed certificate so that it can be added
# to the queue manager keystore in order that the queue manager trust the client
gsk8capicmd -cert -extract -db mutual-tls.kdb -pw myMutualKeystorePW -label SampleClientA -target SampleClientA.pem
```
