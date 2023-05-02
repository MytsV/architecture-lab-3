#!/bin/bash
pos=0.0
step=0.05
interval=0.01
flag=1.0

curl -X POST http://localhost:17000 -d "green"
curl -X POST http://localhost:17000 -d "figure $pos $pos"
curl -X POST http://localhost:17000 -d "update"
sleep $interval

while true; do
  if (( $(echo "$pos >= $flag" |bc ) ));
    then
      pos=0.0
      curl -X POST http://localhost:17000 -d "reset"
      curl -X POST http://localhost:17000 -d "green"
      curl -X POST http://localhost:17000 -d "figure $pos $pos"
      curl -X POST http://localhost:17000 -d "update"
  fi

  pos=$(echo "$pos + $step" |bc )
  curl -X POST http://localhost:17000 -d "move $step $step"
  curl -X POST http://localhost:17000 -d "update"
  sleep $interval
done