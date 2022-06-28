#!/bin/bash
set -o errexit -o nounset -o pipefail -eu

DIR="$HOME/test-contracts"
mkdir -p $DIR

echo "-----------------------"
echo "## Add new CosmWasm contract via gov proposal"
wget "https://github.com/CosmWasm/wasmd/raw/14688c09855ee928a12bcb7cd102a53b78e3cbfb/x/wasm/keeper/testdata/hackatom.wasm" -q -O $DIR/hackatom.wasm 
VAL1_KEY=$($CHAIN_BIN keys show -a val1)
RESP=$($CHAIN_BIN tx gov submit-proposal wasm-store "$DIR/hackatom.wasm" \
  --title "hackatom" \
  --description "hackatom test contact" \
  --deposit 10000stake \
  --run-as $VAL1_KEY \
  --instantiate-everybody "true" \
  --keyring-backend test \
  --from val1 --gas auto --fees 10000stake -y \
  --chain-id $CHAIN_ID \
  -b block -o json --gas-adjustment 1.5)
echo "$RESP"
PROPOSAL_ID=$(echo "$RESP" | jq -r '.logs[0].events[] | select(.type == "submit_proposal") | .attributes[] | select(.key == "proposal_id") | .value')

echo "### Query proposal prevote"
$CHAIN_BIN q gov proposal $PROPOSAL_ID -o json | jq

echo "### Vote proposal"
$CHAIN_BIN tx gov vote $PROPOSAL_ID yes --from val1 --yes --chain-id $CHAIN_ID \
    --fees 500stake --gas auto --gas-adjustment 1.5 -b block -o json | jq
$CHAIN_BIN tx gov vote $PROPOSAL_ID yes --from test1 --yes --chain-id $CHAIN_ID \
    --fees 500stake --gas auto --gas-adjustment 1.5 -b block -o json | jq
$CHAIN_BIN tx gov vote $PROPOSAL_ID yes --from test2 --yes --chain-id $CHAIN_ID \
    --fees 500stake --gas auto --gas-adjustment 1.5 -b block -o json | jq

echo "### Query proposal postvote"
$CHAIN_BIN q gov proposal $PROPOSAL_ID -o json | jq

