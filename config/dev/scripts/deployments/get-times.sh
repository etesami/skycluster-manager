#!/bin/bash

timestamps=$(kubectl get ilptask -o json | jq -r '
.items[] | select(.status.result=="Completed") | .metadata.annotations | 
{
  completion_time: .["skycluster-manager.savitestbed.ca/completion-time"], 
  creation_time: .["skycluster-manager.savitestbed.ca/creation-time"]
} | 
  "\(.completion_time) \(.creation_time)"
')

total_time=0
line_count=0
time_diffs=()

while read -r line; do
  completion_time_org=$(echo $line | awk '{print $1}')
  creation_time_org=$(echo $line | awk '{print $2}')
  completion_time=$(date -d $completion_time_org +%s)
  creation_time=$(date -d $creation_time_org +%s)
  time_diff=$((completion_time - creation_time))
  # add time_diff to time_diffs array
  time_diffs+=("$time_diff")
  echo "$completion_time_org" "$creation_time_org" "$time_diff"

  # Accumulate total time and increment line count
  total_time=$((total_time + time_diff))
  line_count=$((line_count + 1))

done <<< "$timestamps"

# Calculate average if there are lines to avoid division by zero
if [ "$line_count" -gt 0 ]; then
  avg_time=$((total_time / line_count))
  echo "Average time difference: $avg_time seconds and the array is: ${time_diffs[@]}"
else
  echo "No lines to process"
fi