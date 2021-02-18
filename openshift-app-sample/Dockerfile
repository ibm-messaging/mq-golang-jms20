 FROM golang:1.15 as builder
 ENV APP_USER app
 ENV APP_HOME /go/src/openshift-app-sample
 RUN mkdir -p /opt/mqm \
   && chmod a+rx /opt/mqm
 # Location of the downloadable MQ client package \
 ENV RDURL="https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/redist" \
    RDTAR="IBM-MQC-Redist-LinuxX64.tar.gz" \
    VRMF=9.2.0.1
 # Install the MQ client from the Redistributable package. This also contains the
 # header files we need to compile against. Setup the subset of the package
 # we are going to keep - the genmqpkg.sh script removes unneeded parts
 ENV genmqpkg_incnls=1 \
     genmqpkg_incsdk=1 \
     genmqpkg_inctls=1
 RUN cd /opt/mqm \
   && curl -LO "$RDURL/$VRMF-$RDTAR" \
   && tar -zxf ./*.tar.gz \
   && rm -f ./*.tar.gz \
   && bin/genmqpkg.sh -b /opt/mqm
 RUN groupadd $APP_USER && useradd -m -g $APP_USER -l $APP_USER
 RUN mkdir -p $APP_HOME && chown -R $APP_USER:$APP_USER $APP_HOME
 WORKDIR $APP_HOME
 USER $APP_USER
 COPY src/ .
 RUN go build -o openshift-app-sample

 FROM golang:1.15
 ENV APP_USER app
 ENV APP_HOME /go/src/openshift-app-sample
 RUN groupadd $APP_USER && useradd -m -g $APP_USER -l $APP_USER
 RUN mkdir -p $APP_HOME
 # Create the directories the client expects to be present
 RUN mkdir -p /IBM/MQ/data/errors \
   && mkdir -p /.mqm \
   && chmod -R 777 /IBM \
   && chmod -R 777 /.mqm
 WORKDIR $APP_HOME
 COPY --chown=0:0 --from=builder $APP_HOME/openshift-app-sample $APP_HOME
 COPY --chown=0:0 --from=builder /opt/mqm /opt/mqm
 USER $APP_USER
 CMD ["./openshift-app-sample"]