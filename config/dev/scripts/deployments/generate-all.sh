#!/bin/bash

if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]  ; then
  echo "Usage: $0 <action> <compNum/providerNum> <testCase>"
  exit 1
fi

ACTION=$1

PROVIDER_NUM=$2
APP_NUM=$PROVIDER_NUM

TEST_CASE=$3

if [ "$ACTION" == "apply" ] && [ "$PROVIDER_NUM" -gt 100 ]; then
  SLEEP_TIME=30
elif [ "$ACTION" == "apply" ] && [ "$PROVIDER_NUM" -gt 50 ]; then
  SLEEP_TIME=15
elif [ "$ACTION" == "apply" ] && [ "$PROVIDER_NUM" -le 50 ]; then
  SLEEP_TIME=10
else
  SLEEP_TIME=1
fi

if [ "$TEST_CASE" == "test1" ]; then
  echo "Generating deployment for test case 1"
  ./generate-deployment.sh $ACTION app1 $PROVIDER_NUM && sleep $SLEEP_TIME && \
  ./generate-deployment.sh $ACTION app2 $PROVIDER_NUM && sleep $SLEEP_TIME && \
  ./generate-deployment.sh $ACTION app3 $PROVIDER_NUM && sleep $SLEEP_TIME && \
  ./generate-deployment.sh $ACTION app4 $PROVIDER_NUM && sleep $SLEEP_TIME && \
  ./generate-deployment.sh $ACTION app5 $PROVIDER_NUM && sleep $SLEEP_TIME && \
  ./generate-deployment.sh $ACTION app6 $PROVIDER_NUM && sleep $SLEEP_TIME && \
  ./generate-deployment.sh $ACTION app7 $PROVIDER_NUM && sleep $SLEEP_TIME && \
  ./generate-deployment.sh $ACTION app8 $PROVIDER_NUM && sleep $SLEEP_TIME && \
  ./generate-deployment.sh $ACTION app9 $PROVIDER_NUM && sleep $SLEEP_TIME && \
  ./generate-deployment.sh $ACTION app10 $PROVIDER_NUM

elif [ "$TEST_CASE" == "test2" ]; then
  echo "Generating deployment for test case 2"
  for i in {1..10}
  do
    ./generate-deployment-app.sh $ACTION app$i $APP_NUM && sleep $SLEEP_TIME
  done
fi