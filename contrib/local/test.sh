#!/bin/bash
set -o errexit -o nounset -o pipefail -eu

DIR="$HOME/test-contracts"
mkdir -p $DIR

CODE_ID=1
CONTRACT="persistence14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sjvz4fk"

# echo "-----------------------"
# echo "## Create new contract instance"
# INIT="{\"verifier\":\"$($CHAIN_BIN keys show val1 -a --keyring-backend test)\", \"beneficiary\":\"$($CHAIN_BIN keys show test1 -a)\"}"
# $CHAIN_BIN tx wasm instantiate "$CODE_ID" "$INIT" --admin="$($CHAIN_BIN keys show val1 -a --keyring-backend test)" \
#   --from val1 --amount "10000stake" --label "local0.1.0" --gas-adjustment 1.5 --fees "10000stake" \
#   --gas "auto" -y --chain-id $CHAIN_ID -b block -o json | jq

# CONTRACT=$($CHAIN_BIN query wasm list-contract-by-code "$CODE_ID" -o json | jq -r '.contracts[-1]')
# echo "* Contract address: $CONTRACT"

echo "-----------------------"
echo "## Execute contract $CONTRACT"
MSG='{"release":{}}'
$CHAIN_BIN tx wasm execute "$CONTRACT" "$MSG" \
  --from val1 --gas-adjustment 1.5 --fees "10000stake" \
  --gas "auto" -y --chain-id $CHAIN_ID -b block -o json | jq


#echo "### Query all"
#RESP=$($CHAIN_BIN query wasm contract-state all "$CONTRACT" -o json)
#echo "$RESP" | jq
#echo "### Query smart"
#$CHAIN_BIN query wasm contract-state smart "$CONTRACT" '{"verifier":{}}' -o json | jq
#echo "### Query raw"
#KEY=$(echo "$RESP" | jq -r ".models[0].key")
#$CHAIN_BIN query wasm contract-state raw "$CONTRACT" "$KEY" -o json | jq

