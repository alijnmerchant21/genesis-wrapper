# Genesis Wrapper

This repository is a tool to prepare genesis file for Crescent

## Build

```bash
make install
```

## Usage

Initialize chain and prepare genesis file

```bash
wrapper init test
wrapper prepare-genesis testnet mooncat-1-1
```

## Testing Methodology (Reference)

### Build

Git clone the `crescent` repository and build

```bash
git clone https://github.com/crescent-network/crescent.git
make install
```

### Boostrap

```bash
BINARY=crescentd
CHAIN_ID=mooncat-1-1
RPC_PORT=26657
GRPC_PORT=9090
ALICE="guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
BOB="friend excite rough reopen cover wheel spoon convince island path clean monkey play snow number walnut pull lock shoot hurry dream divide concert discover"
ALICE_COINS=1500000000000stake,1000000000000000uatom,1000000000000000uusd
BOB_COINS=1500000000000stake,1000000000000000uatom,1000000000000000uusd

echo $ALICE | $BINARY keys add alice --recover --keyring-backend=test 
echo $BOB | $BINARY keys add bob --recover --keyring-backend=test 
$BINARY add-genesis-account $($BINARY keys show alice --keyring-backend test -a) $ALICE_COINS 
$BINARY add-genesis-account $($BINARY keys show bob --keyring-backend test -a) $BOB_COINS 
$BINARY gentx alice 1000000000stake --chain-id $CHAIN_ID --keyring-backend test
$BINARY collect-gentxs

sed -i '' 's/enable = false/enable = true/g' $HOME/.crescent/config/app.toml
sed -i '' 's/swagger = false/swagger = true/g' $HOME/.crescent/config/app.toml

$BINARY start
```
