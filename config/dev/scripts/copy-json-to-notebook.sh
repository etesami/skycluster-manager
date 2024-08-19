#!/bin/bash

kubectl get cm my-skyapp1-providerattr -o jsonpath="{.data.my-skyapp1-providerattr}" > /home/ubuntu/docs/Python/skycluster/files-ilp/provAttr 
kubectl get cm my-skyapp1-providers -o jsonpath="{.data.my-skyapp1-providers}" > /home/ubuntu/docs/Python/skycluster/files-ilp/providers
kubectl get cm my-skyapp1-edges -o jsonpath="{.data.my-skyapp1-edges}" > /home/ubuntu/docs/Python/skycluster/files-ilp/edges
kubectl get cm my-skyapp1-tasks -o jsonpath="{.data.my-skyapp1-tasks}" > /home/ubuntu/docs/Python/skycluster/files-ilp/tasks
kubectl get cm my-skyapp1-vservices -o jsonpath="{.data.my-skyapp1-vservices}" > /home/ubuntu/docs/Python/skycluster/files-ilp/vservices