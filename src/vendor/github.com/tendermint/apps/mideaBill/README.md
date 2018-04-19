# PersistentMideaBillApplication

1) persistence of state across app restarts (using Tendermint's ABCI-Handshake mechanism)
2) validator set changes

The state is persisted in leveldb along with the last block committed,
and the Handshake allows any necessary blocks to be replayed.
Validator set changes are effected using the following transaction format:

```
val:pubkey1/power1,addr2/power2,addr3/power3"
```

where `power1` is the new voting power for the validator with `pubkey1` (possibly a new one).
There is no sybil protection against new validators joining. 
Validators can be removed by setting their power to `0`.

