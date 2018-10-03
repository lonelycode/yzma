# YzmaDB: a multi-master k/v store

YzmaDB is a key/value store that uses the SWIM gossip protocol (thanks [hashicorp/memberlist](https://github.com/hashicorp/memberlist) and an [Observed-Removed Set CRDT](https://en.wikipedia.org/wiki/Conflict-free_replicated_data_type#OR-Set_(Observed-Removed_Set)) implemented in [BoltDB](https://github.com/etcd-io/bbolt) to provide a multi-master, fully-replicated k/v store that can handle node scaling and shrinkage. 

New nodes joining an active cluster will re-sync the oplog at start from the node they are joining and nodes that have disconnected and reconnect will do the same, the nature of the OR-Set CRDT should ensure data integrity. The nodes do not make use of sharding, all nodes contain all data

It is worth noting that OR-Sets will prefer addition operations over removals in the case of a set merge. In the case of multiple adds (without corresponding removals), the underlying set manager will fall back to a last-write-wins (LWW) to determine the surfaced value.

It is possible to disable LWW fallback, but it is not configurable at th moment, the data of collisioned writes *is* retained and can be surfaced allowing the client to determine the best value to use, however this hasn't been implemented as a client interface yet, just rest assured the data is there).

## Alpha and Unstable

This is an experimental, alpha and unstable application mainly developed to satisfy curiosity. 

*Feedback is greatly appreciated, as are PR's and bug reports!*

## Installation

    go install github.com/lonelycode/yzma

## Configuration

config files are stored in `/etc/yzma/yzma.json` by default but can be set with the `-c` flag.

Sample `yzma.json`:

```json
{
  "API": {
    "Bind": "0.0.0.0:8080"
  },
  "Peering": {
    "Name": "kronk",
    "BindPort": 37001,
    "BindAddr": "0.0.0.0",
    "AdvertisePort": 37001,
    "AdvertiseAddress": "127.0.0.1",
    "Federation": {
      "NodeName": "kronk",
      "APIIngress": "127.0.0.1:37002"
    }
  },
  "Server": {
    "DBPath": "dat.db"
  }
}
```

This will expose the web API on port `8080`, and the peering server will run on ports `37001` and `37002` (the latter is computed, but you can see it is in the advertise address).

Give your nodes unique names, though they will automatically append a UUID to ensure that nodes run as a set or cluster remain unique in the member list, it helps you identify clusters in the logs.

## Start the store:

To start with, just start you initial node:

    ./yzma -c yzma.json 

Then bring up a second node and join it to the cluster (this can also be done with an API call if you want to handle the clustering later)

    ./yzma -c yzma2.json -j 127.0.0.1:37001
    
 Note that for the second node you'll need a new config file with different port values and DB file name.
 
You should now see some output on both nodes that the nodes have joined, you should now be able to add, delete and retrieve data from the store using the API.

> The API only support JSON payloads at the moment, so if you need to store arbitrary data types, wrap them in a JSON map.

## HTTP API

### Creating / deleting keys

Currently there is only a simple HTTP API, you with the following methods:

    GET | POST | DELETE /keys/{key}
    
For example:

    curl -X POST -d @dat.json http://localhost:8080/keys/foo

The payload should be a JSON object, it will automatically be decoded and made available in the payload object, like so:

    curl -X GET http://localhost:8081/keys/foo 
    
    {"Status":"ok","Error":"","Data":{"foo":"bar"}}
    
### Joining and leaving a cluster

    POST /cluster/join
    
    curl -X POST -d '{"Peers": ["127.0.0.1:37001"]}' http://localhost:8080/cluster/join
    
    POST /cluster/leave
        
    curl -X POST http://localhost:8080/cluster/leave
    
## Improvements

Some things that I'd like to investigate further:

- [ ] Compress the oplog so that replication of large data sets can be faster when new nodes join
- [ ] Have nodes only update from an oplog ID to make the replication process faster
- [ ] Move encoding of data on-disk to a binary format, it's JSON at the moment for convenience and switching to msgpack introduces weird decoding issues  
- [ ] Add a CLI for easier testing
        
### Disclaimer

This software is provided as is, and comes with absolute zero warranty.


    
 