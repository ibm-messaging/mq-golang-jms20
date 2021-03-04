# IBM MQ Golang application sample for OpenShift
This sample provides a working template that you can use to build your own Golang
application container image that runs in Red Hat OpenShift under the most secure
"Restricted SCC" and uses either the
[IBM MQ Golang](https://github.com/ibm-messaging/mq-golang) or
[IBM MQ Golang JMS](https://github.com/ibm-messaging/mq-golang-jms20) libraries to connect
to an IBM MQ queue manager.

You can build and run your application using the following simple steps;
1. Modify the sample application code to add your business logic
2. Build the application into a Docker container image locally
3. Push the container image to your OpenShift cluster
4. Configure the variables and run the application!


Let's get started!


## Step 1: Modify the application file to add your code
Add your application logic into the [openshift-app-sample/src/main.go](./src/main.go) file using either the
[IBM MQ Golang](https://github.com/ibm-messaging/mq-golang) or
[IBM MQ Golang JMS](https://github.com/ibm-messaging/mq-golang-jms20) interfaces depending
on your preference.

The file contains all the basic details necessary to create a
connection to a queue manager, so you can simply add your logic to send and/or receive
messages at the bottom of the function.


## Step 2: Build the application into a Docker container image locally
Open a command shell to the same directory as the Dockerfile and execute the following command;
```bash
# Make sure you are in the directory where the Dockerfile is
cd $YOUR_GITHUB_PATH/ibm-messaging/mq-golang-jms20/openshift-app-sample

# Run the dockerfile build to create the container image
docker build -t golang-app -f Dockerfile .
```


## Step 3: Push the container image to your OpenShift cluster
This step will vary depending on what sort of OpenShift cluster you are going to deploy to - 
for example a registry that is part of the cluster or an external registry that can be
accessed by the cluster.

In this example we are using the IBM Cloud Container Registry (in London) that is hosted
at `uk.icr.io`.

```bash
# Tag the new image against the target registry
docker tag golang-app uk.icr.io/golang-sample/golang-app:1.0

# Push the updated image to your registry
docker push uk.icr.io/golang-sample/golang-app:1.0
```

**Note:** If you are iterating over this step multiple times while testing your application then
don't forget to increase the tag version each time as image caching on cluster may mean
that your updates don't actually get used when you expect them to!


## Step 4: Configure the variables and run the application!
Next we will do a one-time set up of some basic objects on the cluster in order to
supply configuration settings to the application. 

```bash
# Create a service account that we can use to deploy using the Restricted SCC
oc apply -f ./yaml/sa-pod-deployer.yaml

# Create a config map containing the details of your queue manager.
#
# If your queue manager is in the same OpenShift cluster then the hostname will be the
# name of the "service".
oc create configmap qmgr-details \
    --from-literal=HOSTNAME=mydynamichostname \
    --from-literal=PORT=34567 \
    --from-literal=QMNAME=QM100 \
    --from-literal=CHANNELNAME=SYSTEM.DEF.SVRCONN

# If necessary, create a secret to hold the username and password your application should
# use to authenticate to the queue manager.
oc create secret generic qmgr-credentials \
    --from-literal=USERNAME=appuser100 \
    --from-literal=PASSWORD='password100'
```

Now run the application!
```bash
# Before you run this command, be sure to update the "image" attribute on
# line 8 of pod-sample.yaml so that it matches your image tag version from step 3.
oc apply -f ./yaml/pod-sample.yaml --as=my-service-account
```

Congratulations on running your MQ Golang application in OpenShift - and happy messaging!