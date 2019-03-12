# JMS 2.0 style interface for IBM MQ applications in Golang

This repository provides a developer-friendly [JMS 2.0](https://javaee.github.io/jms-spec/pages/JMS20FinalRelease) style programming interface to enable Golang applications to send and receive messages through IBM MQ.

- Developers with experience using JMS in Java will be immediately familiar with using this JMS style Golang API
- All Golang developers benefit from a simple messaging API that is based on industry best practice and experience, with no prior knowledge of Java or JMS required!

For many years developers have been able to write messaging applications in Golang that connect to IBM MQ using the [mq-golang](https://github.com/ibm-messaging/mq-golang) module. That module is very powerful as it exposes the traditional IBM MQ ["MQI" (Message Queueing Interface)](https://www.ibm.com/support/knowledgecenter/en/SSFKSJ_latest/com.ibm.mq.dev.doc/q025720_.htm) but can be difficult to pick up for new developers.

This `mq-golang-jms20` module provides a simplified interface for sending and receiving messages with IBM MQ in Golang
by providing a client library that implements a subset of the JMS 2.0 programming API. This makes developing messaging applications in Golang very
straightforward as shown in the samples below, and allows you to make use of the existing documentation and collateral
for developing applications in JMS 2.0.

If you're not familiar with IBM MQ then you'll also find the [MQ Essentials tutorial](https://developer.ibm.com/messaging/learn-mq/mq-tutorials/getting-started-mq/) on the [Learn MQ](https://developer.ibm.com/messaging/learn-mq/) site important to understand how IBM MQ solves key problems for your application solution.

Note for experienced MQ / JMS developers: This repository provides a JMS style programming interface, but there is no use of Java as part of the implementation. It also does not use the IBM MQ Java client, or IBM MQ JMS client. The implementation is written in Golang and builds upon the [mq-golang](https://github.com/ibm-messaging/mq-golang) module, which itself uses Cgo to invoke the MQ C client library and communicate with the queue manager.


# Table of Contents   

* [Code samples](#code-samples)
* [Getting Started](#getting-started)
* [Comments on mapping JMS 2.0 to Golang](#comments-on-mapping-jms-20-to-golang)
* [Contributing](#contributing)
* [Licensing](#licensing)


## Code samples
The following samples demonstrate how simple it is to achieve common programming scenarios
in this JMS 2.0 style interface in Golang. Note that there are additional working samples in the
[testcases included in this repo](#more-detailed-code-samples).


### Send and receive a message containing a text string
(from [sample_sendreceive_test.go](sample_sendreceive_test.go))
Note that for illustration purposes this sample only has limited error handling, which you should never do in production application code! Please see the TestSampleSendReceiveWithErrorHandling function for an equivalent sample that demonstrates good practice for error handling.
```
// Create a ConnectionFactory using details stored in some external property files
cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
if cfErr != nil {
  // Handle error here!
}

// Creates a connection to the queue manager.
// We use a "defer" call to make sure that the connection is closed at the end of the method
context, ctxErr := cf.CreateContext()
if ctxErr != nil {
  // Handle error here!
}
if context != nil {
  defer context.Close()
}

// Create a Queue object that points at an IBM MQ queue
queue := context.CreateQueue("DEV.QUEUE.1")

// Send a message to the queue that contains the specified text string
context.CreateProducer().SendString(queue, "My first message")

// Create a consumer, using Defer to make sure it gets closed at the end of the method
consumer, conErr := context.CreateConsumer(queue)
if conErr != nil {
  // Handle error here!
}
if consumer != nil {
  defer consumer.Close()
}

// Receive a message from the queue and return the string from the message body
rcvBody := consumer.ReceiveStringBodyNoWait()

if rcvBody != nil {
  fmt.Println("Received text string: " + *rcvBody)
} else {
  fmt.Println("No message received")
}
```

### Send a non-persistent message
(from [deliverymode_test.go](deliverymode_test.go))
```
msgBody = "My non-persistent message"
err3 := context.CreateProducer().SetDeliveryMode(jms20subset.DeliveryMode_NON_PERSISTENT).SendString(queue, msgBody)
```

### Error handling
(from [sample_errorhandling_test.go](sample_errorhandling_test.go))
```
// Create a ConnectionFactory using some property files
cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
assert.Nil(t, cfErr)

// Set a value we know will cause a failure
cf.UserName = "wrong_user"

// Creates a connection to the queue manager
context, err := cf.CreateContext()
assert.NotNil(t, err)
if context != nil {
  defer context.Close()
}

// Check the error code that comes back in the Error (JMS uses a String for the code)
assert.Equal(t, "2035", err.GetErrorCode())
assert.Equal(t, "MQRC_NOT_AUTHORIZED", err.GetReason())
```

### More detailed code samples
Other sample code can be found in the testcase files as follows. When writing your own applications you will
generally replace the various "assert" calls that test the successful execution of the application logic with
your own error handling or logging.
* Creating a ConnectionFactory - [connectionfactory_test.go](connectionfactory_test.go)
* Send/receive a text string - [sample_sendreceive_test.go](sample_sendreceive_test.go)
* Send a message as Persistent or NonPersistent - [deliverymode_test.go](deliverymode_test.go)
* Get by CorrelationID - [getbycorrelid_test.go](getbycorrelid_test.go)
* Request/reply messaging pattern - [requestreply_test.go](requestreply_test.go)
* Sending a message that expires after a period of time - [timetolive_test.go](timetolive_test.go)
* Handle error codes returned by the queue manager - [sample_errorhandling_test.go](sample_errorhandling_test.go)

As normal with Go, you can run any individual testcase by executing a command such as;
```
go test -run TestSampleSendReceiveWithErrorHandling
```


## Getting started

### Installing the pre-requisites
The IBM MQ client on which this library depends is supported on Linux and Windows, and is [now available for development use on MacOS](https://developer.ibm.com/messaging/2019/02/05/ibm-mq-macos-toolkit-for-developers/)).

1. Install Golang
    - This library has been validated with Golang v1.11.4. If you don't have Golang installed on your system you can [download it here](https://golang.org/doc/install) for MacOS, Linux or Windows
2. Install 'dep' to manage the dependent packages
    - See [Installation instructions](https://github.com/golang/dep#installation) for details
3. Install the MQ Client library
    - If you have a full MQ server with a queue manager installed on your machine then you already have the client library
    - If you don't have a queue manager installed on your machine then you can download the "redistributable client" library for IBM MQ 9.1.1 CD or higher for [Linux](https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/redist/9.1.1.0-IBM-MQC-Redist-LinuxX64.tar.gz), [Windows](https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/redist/9.1.1.0-IBM-MQC-Redist-Win64.zip) or [MacOS](https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/mactoolkit/IBM-MQ-Toolkit-Mac-x64-9.1.1.0.tar.gz)
      - Simply unzip the archive and make a note of the installation location. For ease of configuration you may wish to unzip the archive into the default install IBM MQ location for your platform
      - Note that v9.1.1 (CD) or higher of the MQ client library is required as it includes header files that are not present in v9.1.0 LTS or below.
4. Git clone this project to download this JMS style implementation onto your workstation
  ```
  # Update and set the GOPATH variable to match your workspace
  export GOPATH=/home/myuser/workspace

  # Clone this MQ JMS Golang repo into your local workspace
  git clone https://github.com/ibm-messaging/mq-golang-jms20.git $GOPATH/src/github.com/ibm-messaging/mq-golang-jms20
  ```
5. Deploy an IBM MQ queue manager
    - If you have an existing queue manager then you can continue to use that
    - You can also deploy a queue manager using one of the following simple approaches
      - Select the Lite plan to deploy a free queue manager using the [IBM MQ on Cloud service](https://cloud.ibm.com/catalog/services/mq) (IBM SaaS offering)
      - Deploy IBM MQ for Developers for free in a container using the sample Docker container as described in the [Ready, Set, Connect - Docker tutorial](https://developer.ibm.com/messaging/learn-mq/mq-tutorials/mq-connect-to-queue-manager/#docker)
      - Install IBM MQ for Developers for free on [Windows](https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqadv/mqadv_dev911_windows.zip), [Linux](https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqadv/mqadv_dev911_linux_x86-64.tar.gz) or [Ubuntu](https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqadv/mqadv_dev911_ubuntu_x86-64.tar.gz)

### Configuring your environment

First you must configure your command console environment as described in the [mq-golang Getting Started instructions](https://github.com/ibm-messaging/mq-golang#getting-started) so that the necessary flags are set;
```
# Configure your Go environment variables (update to match your own setup)
export GOROOT=/usr/local/go
export GOPATH=/home/myuser/workspace
export PATH=$PATH:$GOROOT/bin

# Set the CGO flags to allow the compilation of the Go/C client interface
export CGO_LDFLAGS_ALLOW="-Wl,-rpath.*"
```


**If your client install is not located in the default installation location**, for example `/opt/mqm` then you also need to set the follow environment variables to point at your installation location. For example on Linux or MacOS;
```
export MQ_INSTALLATION_PATH=$HOME/9.1.1.0-IBM-MQC-Redist-LinuxX64
export CGO_CFLAGS="-I$MQ_INSTALLATION_PATH/inc"
export CGO_LDFLAGS="-L$MQ_INSTALLATION_PATH/lib64 -Wl,-rpath,$MQ_INSTALLATION_PATH/lib64"
```

Download the Golang modules on which this project depends by running the dep command;
```
cd $GOPATH/src/github.com/ibm-messaging/mq-golang-jms20/
dep ensure
```

Confirm the settings are correct by compiling the MQ JMS Golang package, for example as follows; (no errors will be shown if successful)
```
cd $GOPATH/src/github.com/ibm-messaging/mq-golang-jms20/mqjms/
go build
```

### Verify the installation by executing the tests
This project includes a series of tests that validate the successful operation of the Golang JMS style client library.

The test cases use the `CreateConnectionFactoryFromDefaultJSONFiles` method to obtain details of a queue manager to connect to from two JSON files in your `/Downloads` directory;
- `connection_info.json` contains information like the hostname/port/channel
  - If you are using the MQ on Cloud service you can download a pre-populated file directly from the queue manager details page as [described here](https://cloud.ibm.com/docs/services/mqcloud/mqoc_jms_tls.html#connection_info-json)
  - Otherwise you can insert details of your own queue manager into [this sample file](./config-samples/connection_info.json) and copy it to your `/Downloads` directory
- `apiKey.json` contains the Application username and password that will be used to connect to your queue manager
  - If you are using the MQ on Cloud service you can download a pre-populated file directly from the Application Permissions tab in the service console as [described here](https://cloud.ibm.com/docs/services/mqcloud/mqoc_jms_tls.html#apikey-json)
  - Otherwise you can insert details of your own queue manager into [this sample file](./config-samples/apiKey.json) and copy it to your `/Downloads` directory

Once you have added the details of your queue manager and user credentials into the two JSON files and placed them in your `/Downloads` directory you are ready to run the test, which is done in the same way as any other Go tests.

Note that the tests require the queues `DEV.QUEUE.1` and `DEV.QUEUE.2` to be defined on your queue manager, be empty of messages and be accessible to the application username you are using. This will be the case by default for queue managers provisioned through the MQ on Cloud service, but may require manual configuration for queue managers you have created through other means.
```
> cd $GOPATH/src/github.com/ibm-messaging/mq-golang-jms20/
> go test -v

=== RUN   TestLoadCFFromJSON
--- PASS: TestLoadCFFromJSON (0.59s)
...
...
PASS
ok  	github.com/ibm-messaging/mq-golang-jms20	11.308s
```


### Writing your own Golang application that talks to IBM MQ
Writing your own application to talk to IBM MQ is simple - as shown in the [sample_sendreceive_test.go](sample_sendreceive_test.go) sample. Simply import this module into your source file, and get started!
```
import (
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"

)
```

The first thing you'll need to do is create a ConnectionFactory object so that you can connect to your queue manager. There are two ways of doing this as follows;
1. Populate a ConnectionFactory object using properties outside your program
    - This is a similar approach to doing a JNDI lookup in a JMS application in Java - it is good practice because it avoids hardcoding details like hostnames, ports, usernames and passwords in your application
    - This approach also allows the details to be updated without having to recompile the whole application - for example when you promote your application from development to production
    - The `CreateConnectionFactoryFromDefaultJSONFiles` method shown in the samples loads the values from two JSON files on the filesystem
    - Similarly you could implement a utility method to download the details from an HTTP server or some other location when the application starts up
2. Create a new ConnectionFactory object and set the variables in your application code
    - This is a quick and easy way to get started, but less desirable for production quality applications
    - You can hardcode the values in your source file, or perhaps look them up from environment variables as shown below
```
cf := mqjms.ConnectionFactoryImpl{
  QMName:      "QM_ONE",
  Hostname:    "random.hostname.com",
  PortNumber:  1414,
  ChannelName: "SYSTEM.APP.SVRCONN",
  UserName:    os.Getenv("MQ_SAMP_USERID"),
  Password:    os.Getenv("MQ_SAMP_PASSWORD"),
}
```

Once you have successfully created a connection to the queue manager (using `cf.CreateContext()`) you can send and receive messages in whatever way your application requires.

**Top tip**: Don't forget to include the necessary error handling in your application code - the world has too many `// This should never happen` statements that ended up being triggered for the first time in Production!


## Comments on mapping JMS 2.0 to Golang
We have attempted to keep this rendering of JMS 2.0 into Golang as close to the original Java
JMS interface and method names as possible, in order that it be immediately familiar to anyone who has used
JMS, and also to allow users to use JMS documentation and have a good understanding of what
to expect when using this Golang module. However Golang is a different language than Java so there are some
areas where the exact spelling has diverged a little from the Java form.

* Use of JMSRuntimeException
  * In Java, JMS 2.0 has converted all exceptions to be subclasses of RuntimeException which means that you do not have to explicitly write code to catch them, and instead they will be propagated up the stack if you do not write any error handling
  * Golang has a strong preference for enforcing error checking so we have implemented some checked errors in the Golang interfaces, but also tried to omit returning errors from some methods that should typically be safe to call without error checking in well written applications
  * This also has an effect in the amount of "method chaining" that is replicated in the Golang JMS interfaces, since you can only chain method calls if the method returns a single return object (and Golang errors are returned rather than "thrown")
* Automatically closing objects
  * JMS 2.0 makes use of java.lang.AutoCloseable to automatically close objects
  * Golang doesn't have a direct equivalent so we recommend using "defer" to ensure that objects are automatically closed when the function completes
* Method overloading
  * JMS 2.0 makes extensive use of method overloading in Java to define multiple methods with the same name but different parameters (for example the five different "send" methods on a [JMSProducer](https://github.com/eclipse-ee4j/jms-api/blob/master/src/main/java/javax/jms/JMSProducer.java#L87))
  * Golang doesn't allow method overloading so we have introduced slightly different methods names, such as Send and SendString in the [Golang JMSProducer object](./jms20subset/JMSProducer.go)
* Generics
  * Similarly, JMS 2.0 has used Generics in Java to allow you to receive a [message body directly without casting](https://javaee.github.io/jms-spec/pages/JMS20MeansLessCode#receiving-synchronously-can-receive-mesage-payload-directly)
  * In the Golang rendering we simulate that by introducing a differently named method for each supported data type as in the [Golang JMSConsumer object](./jms20subset/JMSConsumer.go)


## Contributing
We love to receive your input - if you find a bug, please [raise an Issue](https://github.com/ibm-messaging/mq-golang-jms20/issues). Even better, you can submit a Pull Request to fix the bug or contribute additional functionality to this module, such as implementing an additional piece of the JMS 2.0 specification into this Golang style client library.

Contributions to this package must be made under the terms of the IBM Contributor License Agreement, found in the [CLA file](CLA.md) of this repository. When submitting a pull request, you must include a statement stating you accept the terms in the CLA.


## Licensing
- All content found in this repository is licensed under the Eclipse Public License. In particular;
  - The JMS 2.0 interfaces are licensed under the Eclipse Public License from [eclipse-ee4j/jms-api](https://github.com/eclipse-ee4j/jms-api/blob/master/LICENSE.md), which permits derivative works such as the re-spelling of JMS into Golang as found in the `/jms20subset` directory here
  - The remainder of the content in this repository, including the MQ specific implementation of those JMS 2.0 interfaces in Golang (as found in `/mqjms`) is also licensed under the Eclipse Public License

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
