 FROM golang:1.15 as builder
 ENV APP_USER app
 ENV APP_HOME /go/src/openshift-app-sample
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
 WORKDIR $APP_HOME
 COPY --chown=0:0 --from=builder $APP_HOME/openshift-app-sample $APP_HOME
 USER $APP_USER
 CMD ["./openshift-app-sample"]