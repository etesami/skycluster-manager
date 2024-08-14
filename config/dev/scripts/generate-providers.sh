#!/usr/bin/env bash

help="Usage: 
    ./generate.sh [--vs] [--pv] [--attr] [--pv-num=<pv-num>] [--vs-num=<vs-num>]
Options:
    -h --help     Show this screen.
    --vs          Generate VirtualServices, <vs-num> and <pv-num> are required. 
    --pv          Generate Providers, <pv-num> is required.
    --attr        Generates ProviderAttributes, <pv-num> is required.
    <pv-num>      Number of providers.
    <vs-num>      Number of virtual services."
eval "$(docopts -A args -h "$help" : "$@")"

PROVIDER_NUM="${args[--pv-num]}"
VS_NUM="${args[--vs-num]}"
PV="${args[--pv]}"
VS="${args[--vs]}"
ATTR="${args[--attr]}"
output_dir="./output"
echo $PV $VS $ATTR  $PROVIDER_NUM $VS_NUM

# Function to generate random latency value
generate_latency() {
    echo "$((RANDOM % 100))ms"
}

# Function to generate random egress data cost value
generate_egress_data_cost() {
    echo "\"$((RANDOM % 100))\""
}

generate_providers_names() {
  local num=$1
  local providers=()
  for i in $(seq 1 ${num}); do
    # Generate Provider name
    provider_name="provider-$i"
    providers+=("$provider_name")
  done
  echo "${providers[@]}"
}

shuffle_array() {
  local array=("${!1}")  # Input array is passed by reference
  local i tmp size max rand
  
  size=${#array[@]}
  max=$(( 32768 / size * size ))

  # Fisher-Yates shuffle algorithm
  for ((i=size-1; i>0; i--)); do
    rand=$((RANDOM % (i+1)))
    tmp=${array[i]}
    array[i]=${array[rand]}
    array[rand]=$tmp
  done
  
  # Print the shuffled array as space-separated values
  echo "${array[@]}"
}

# Types array
types=("cloud" "edge" "nte")

# Regions array
regions=("west1" "west2" "east1" "east2" "north1" "south1")

# Directory to store the generated YAML files
mkdir -p "$output_dir"

# Generate Providers
if [[ $PV = "true" && -n $PROVIDER_NUM ]]; then
  echo "Generating Providers..., number of providers: ${PROVIDER_NUM}"
  group_counter=1
  file_counter=1
  providers=()
  for ((i=1; i<=PROVIDER_NUM; i++)); do
      # Randomly select a type
      type=${types[$RANDOM % ${#types[@]}]}
      
      # Randomly select a region
      region=${regions[$RANDOM % ${#regions[@]}]}

      # Generate Provider name
      provider_name="provider-$i"
      providers+=("$provider_name")

      # Generate Provider YAML
      cat <<EOF >> "$output_dir/provider_g$group_counter.yaml"
apiVersion: core.skycluster-manager.savitestbed.ca/v1alpha1
kind: Provider
metadata:
  name: $provider_name
spec:
  name: $provider_name
  region: $region
  zone: default
  type: $type
---
EOF

      # Increment the file counter
      ((file_counter++))

      # If file_counter exceeds 10, reset it and increment the group counter
      if (( file_counter > 10 )); then
          file_counter=1
          ((group_counter++))
      fi
  done
fi

if [[ $ATTR = "true" && -n $PROVIDER_NUM ]]; then
  echo "Generating ProviderAttributes..."
  providers=($(generate_providers_names $PROVIDER_NUM))
  file_counter=1
  group_counter=1
  for src in "${providers[@]}"; do
      for dst in "${providers[@]}"; do
          if [ "$src" != "$dst" ]; then
              cat <<EOF >> "$output_dir/providerattr-g$group_counter.yaml"
apiVersion: core.skycluster-manager.savitestbed.ca/v1alpha1
kind: ProviderAttribute
metadata:
  name: providerattr-$src-$dst
spec:
  providerReference:
    name: $src
    namespace: default
  providerMetrics:
    - dstProviderName: $dst
      latency: $(generate_latency)
      egressDataCost: $(generate_egress_data_cost)
---
EOF
            # Increment the file counter
            ((file_counter++))

            # If file_counter exceeds 10, reset it and increment the group counter
            if (( file_counter > 100 )); then
                file_counter=1
                ((group_counter++))
            fi
          fi
      done
  done
fi


if [[ $VS = "true"  && -n $VS_NUM && -n $PROVIDER_NUM ]]; then
  echo "Generating VirtualServices..."
  file_counter=1
  group_counter=1
  providers=($(generate_providers_names $PROVIDER_NUM))
  for ((i=1; i<=VS_NUM; i++)); do
      shuffled_providers=$(shuffle_array providers[@])
      # Convert the space-separated string back to an array
      IFS=' ' read -r -a shuffled_providers <<< "$shuffled_providers"
      # Generate Provider YAML
      cat <<EOF >> "$output_dir/vs_g$group_counter.yaml"
apiVersion: core.skycluster-manager.savitestbed.ca/v1alpha1
kind: VirtualService
metadata:
  name: vs-$i
spec:
  name: vs-$i
  vservicecosts:
EOF
      random_index=$((RANDOM % ${PROVIDER_NUM}))
      echo "Random index: $random_index"
      for ((j=0; j<=random_index; j++)); do
        cat <<EOF >> "$output_dir/vs_g$group_counter.yaml"
    - providerName: ${shuffled_providers[$j]}
      cost: $(generate_egress_data_cost)
EOF
      done
      cat <<EOF >> "$output_dir/vs_g$group_counter.yaml"
---
EOF

      # Increment the file counter
      ((file_counter++))

      # If file_counter exceeds 10, reset it and increment the group counter
      if (( file_counter > 10 )); then
          file_counter=1
          ((group_counter++))
      fi
  done
fi