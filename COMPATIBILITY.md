# Redis Compatibility

Status of all Redis 6.2 core commands (plus JSON and Bloom Filter module commands)
in this implementation.

## Legend

- ✅ — implemented
- 🚫 — not implemented (missing case in switch or no handler)

---

## String commands

| Command | Status | Notes |
|---------|--------|-------|
| APPEND | 🚫 | |
| DECR | ✅ | |
| DECRBY | ✅ | |
| GET | ✅ | |
| GETDEL | ✅ | |
| GETEX | 🚫 | |
| GETRANGE | 🚫 | |
| GETSET | ✅ | |
| INCR | ✅ | |
| INCRBY | ✅ | |
| INCRBYFLOAT | 🚫 | |
| MGET | ✅ | |
| MSET | 🚫 | |
| MSETNX | 🚫 | |
| PSETEX | 🚫 | |
| SET | ✅ | |
| SETEX | ✅ | |
| SETNX | ✅ | |
| SETRANGE | 🚫 | |
| STRLEN | ✅ | |
| SUBSTR | ✅ | |

---

## Hash commands

| Command | Status | Notes |
|---------|--------|-------|
| HDEL | 🚫 | |
| HEXISTS | 🚫 | |
| HGET | 🚫 | |
| HGETALL | 🚫 | |
| HINCRBY | 🚫 | |
| HINCRBYFLOAT | 🚫 | |
| HKEYS | 🚫 | |
| HLEN | 🚫 | |
| HMGET | 🚫 | |
| HMSET | 🚫 | |
| HRANDFIELD | 🚫 | |
| HSCAN | 🚫 | |
| HSET | 🚫 | |
| HSETNX | 🚫 | |
| HSTRLEN | 🚫 | |
| HVALS | 🚫 | |

---

## List commands

| Command | Status | Notes |
|---------|--------|-------|
| BLMOVE | 🚫 | |
| BLPOP | 🚫 | |
| BRPOP | 🚫 | |
| BRPOPLPUSH | 🚫 | |
| LINDEX | ✅ | |
| LINSERT | ✅ | |
| LLEN | ✅ | |
| LMOVE | 🚫 | |
| LPOP | ✅ | |
| LPOS | 🚫 | |
| LPUSH | ✅ | |
| LPUSHX | ✅ | |
| LRANGE | ✅ | |
| LREM | ✅ | |
| LSET | ✅ | |
| LTRIM | ✅ | |
| RPOP | ✅ | |
| RPOPLPUSH | 🚫 | |
| RPUSH | ✅ | |
| RPUSHX | ✅ | |

---

## Set commands

| Command | Status | Notes |
|---------|--------|-------|
| SADD | ✅ | |
| SCARD | ✅ | |
| SDIFF | ✅ | |
| SDIFFSTORE | ✅ | |
| SINTER | ✅ | |
| SINTERSTORE | ✅ | |
| SISMEMBER | ✅ | |
| SMEMBERS | ✅ | |
| SMISMEMBER | 🚫 | |
| SMOVE | ✅ | |
| SPOP | ✅ | |
| SRANDMEMBER | ✅ | |
| SREM | ✅ | |
| SSCAN | 🚫 | |
| SUNION | ✅ | |
| SUNIONSTORE | ✅ | |

---

## Sorted set commands

| Command | Status | Notes |
|---------|--------|-------|
| BZPOPMAX | 🚫 | |
| BZPOPMIN | 🚫 | |
| ZADD | 🚫 | |
| ZCARD | 🚫 | |
| ZCOUNT | 🚫 | |
| ZDIFF | 🚫 | |
| ZDIFFSTORE | 🚫 | |
| ZINCRBY | 🚫 | |
| ZINTER | 🚫 | |
| ZINTERSTORE | 🚫 | |
| ZLEXCOUNT | 🚫 | |
| ZPOPMAX | 🚫 | |
| ZPOPMIN | 🚫 | |
| ZRANDMEMBER | 🚫 | |
| ZRANGE | 🚫 | |
| ZRANGEBYLEX | 🚫 | |
| ZRANGEBYSCORE | 🚫 | |
| ZRANK | 🚫 | |
| ZREM | 🚫 | |
| ZREMRANGEBYLEX | 🚫 | |
| ZREMRANGEBYRANK | 🚫 | |
| ZREMRANGEBYSCORE | 🚫 | |
| ZREVRANGE | 🚫 | |
| ZREVRANGEBYLEX | 🚫 | |
| ZREVRANGEBYSCORE | 🚫 | |
| ZREVRANK | 🚫 | |
| ZSCAN | 🚫 | |
| ZSCORE | 🚫 | |
| ZUNION | 🚫 | |
| ZUNIONSTORE | 🚫 | |

---

## Stream commands

| Command | Status | Notes |
|---------|--------|-------|
| XACK | 🚫 | Not implementing |
| XADD | 🚫 | Not implementing |
| XAUTOCLAIM | 🚫 | Not implementing |
| XCLAIM | 🚫 | Not implementing |
| XDEL | 🚫 | Not implementing |
| XGROUP CREATE | 🚫 | Not implementing |
| XGROUP CREATECONSUMER | 🚫 | Not implementing |
| XGROUP DELCONSUMER | 🚫 | Not implementing |
| XGROUP DESTROY | 🚫 | Not implementing |
| XGROUP SETID | 🚫 | Not implementing |
| XINFO CONSUMERS | 🚫 | Not implementing |
| XINFO GROUPS | 🚫 | Not implementing |
| XINFO STREAM | 🚫 | Not implementing |
| XLEN | 🚫 | Not implementing |
| XPENDING | 🚫 | Not implementing |
| XRANGE | 🚫 | Not implementing |
| XREAD | 🚫 | Not implementing |
| XREADGROUP | 🚫 | Not implementing |
| XREVRANGE | 🚫 | Not implementing |
| XSETID | 🚫 | Not implementing |
| XTRIM | 🚫 | Not implementing |

---

## Bitmap commands

| Command | Status | Notes |
|---------|--------|-------|
| BITCOUNT | 🚫 | |
| BITFIELD | 🚫 | |
| BITFIELD_RO | 🚫 | |
| BITOP | 🚫 | |
| BITPOS | 🚫 | |
| GETBIT | 🚫 | |
| SETBIT | 🚫 | |

---

## HyperLogLog commands

| Command | Status | Notes |
|---------|--------|-------|
| PFADD | ✅ | |
| PFCOUNT | ✅ | |
| PFDEBUG | 🚫 | |
| PFMERGE | ✅ | |
| PFSELFTEST | 🚫 | |

---

## Geospatial commands

| Command | Status | Notes |
|---------|--------|-------|
| GEOADD | 🚫 | Not implementing |
| GEODIST | 🚫 | Not implementing |
| GEOHASH | 🚫 | Not implementing |
| GEOPOS | 🚫 | Not implementing |
| GEORADIUS | 🚫 | Not implementing |
| GEORADIUS_RO | 🚫 | Not implementing |
| GEORADIUSBYMEMBER | 🚫 | Not implementing |
| GEORADIUSBYMEMBER_RO | 🚫 | Not implementing |
| GEOSEARCH | 🚫 | Not implementing |
| GEOSEARCHSTORE | 🚫 | Not implementing |

---

## Pub/Sub commands

| Command | Status | Notes |
|---------|--------|-------|
| PSUBSCRIBE | ✅ | |
| PUBLISH | ✅ | |
| PUBSUB CHANNELS | 🚫 | |
| PUBSUB NUMPAT | 🚫 | |
| PUBSUB NUMSUB | 🚫 | |
| PUNSUBSCRIBE | 🚫 | |
| SPUBLISH | 🚫 | |
| SSUBSCRIBE | 🚫 | |
| SUBSCRIBE | ✅ | |
| SUNSUBSCRIBE | 🚫 | |
| UNSUBSCRIBE | 🚫 | |

---

## Transaction commands

| Command | Status | Notes |
|---------|--------|-------|
| DISCARD | 🚫 | |
| EXEC | 🚫 | |
| MULTI | 🚫 | |
| UNWATCH | 🚫 | |
| WATCH | 🚫 | |

---

## Scripting commands

| Command | Status | Notes |
|---------|--------|-------|
| EVAL | 🚫 | Not implementing |
| EVAL_RO | 🚫 | Not implementing |
| EVALSHA | 🚫 | Not implementing |
| EVALSHA_RO | 🚫 | Not implementing |
| SCRIPT DEBUG | 🚫 | Not implementing |
| SCRIPT EXISTS | 🚫 | Not implementing |
| SCRIPT FLUSH | 🚫 | Not implementing |
| SCRIPT KILL | 🚫 | Not implementing |
| SCRIPT LOAD | 🚫 | Not implementing |

---

## Connection commands

| Command | Status | Notes |
|---------|--------|-------|
| AUTH | 🚫 | |
| CLIENT CACHING | 🚫 | |
| CLIENT GETNAME | 🚫 | |
| CLIENT GETREDIR | 🚫 | |
| CLIENT ID | ✅ | |
| CLIENT INFO | ✅ | |
| CLIENT KILL | 🚫 | |
| CLIENT LIST | 🚫 | |
| CLIENT NO-EVICT | 🚫 | |
| CLIENT NO-TOUCH | 🚫 | |
| CLIENT PAUSE | 🚫 | |
| CLIENT REPLY | 🚫 | |
| CLIENT SETNAME | 🚫 | |
| CLIENT TRACKING | 🚫 | |
| CLIENT TRACKINGINFO | 🚫 | |
| CLIENT UNBLOCK | 🚫 | |
| CLIENT UNPAUSE | 🚫 | |
| ECHO | 🚫 | |
| HELLO | 🚫 | |
| PING | ✅ | |
| QUIT | ✅ | |
| RESET | 🚫 | |
| SELECT | ✅ | |

---

## Server commands

| Command | Status | Notes |
|---------|--------|-------|
| ACL CAT | 🚫 | |
| ACL DELUSER | 🚫 | |
| ACL DRYRUN | 🚫 | |
| ACL GENPASS | 🚫 | |
| ACL GETUSER | 🚫 | |
| ACL LIST | 🚫 | |
| ACL LOAD | 🚫 | |
| ACL LOG | 🚫 | |
| ACL SAVE | 🚫 | |
| ACL SETUSER | 🚫 | |
| ACL USERS | 🚫 | |
| ACL WHOAMI | 🚫 | |
| BGREWRITEAOF | 🚫 | |
| BGSAVE | ✅ | |
| COMMAND | 🚫 | |
| COMMAND COUNT | 🚫 | |
| COMMAND DOCS | 🚫 | |
| COMMAND GETKEYS | 🚫 | |
| COMMAND GETKEYSANDFLAGS | 🚫 | |
| COMMAND INFO | 🚫 | |
| COMMAND LIST | 🚫 | |
| CONFIG GET | 🚫 | |
| CONFIG RESETSTAT | 🚫 | |
| CONFIG REWRITE | 🚫 | |
| CONFIG SET | 🚫 | |
| DBSIZE | ✅ | |
| DEBUG | 🚫 | |
| FAILOVER | 🚫 | |
| FLUSHALL | ✅ | |
| FLUSHDB | ✅ | |
| INFO | 🚫 | |
| LASTSAVE | 🚫 | |
| LATENCY DOCTOR | 🚫 | |
| LATENCY GRAPH | 🚫 | |
| LATENCY HISTORY | 🚫 | |
| LATENCY LATEST | 🚫 | |
| LATENCY RESET | 🚫 | |
| LOLWUT | 🚫 | |
| MEMORY DOCTOR | 🚫 | |
| MEMORY MALLOC-STATS | 🚫 | |
| MEMORY PURGE | 🚫 | |
| MEMORY STATS | 🚫 | |
| MEMORY USAGE | 🚫 | |
| MODULE LIST | 🚫 | |
| MODULE LOAD | 🚫 | |
| MODULE LOADEX | 🚫 | |
| MODULE UNLOAD | 🚫 | |
| MONITOR | 🚫 | |
| PSYNC | 🚫 | |
| REPLCONF | 🚫 | |
| REPLICAOF | 🚫 | |
| RESTORE-ASKING | 🚫 | |
| ROLE | 🚫 | |
| SAVE | 🚫 | |
| SHUTDOWN | 🚫 | |
| SLAVEOF | 🚫 | |
| SLOWLOG GET | 🚫 | |
| SLOWLOG LEN | 🚫 | |
| SLOWLOG RESET | 🚫 | |
| SWAPDB | 🚫 | |
| SYNC | 🚫 | |
| TIME | 🚫 | |

---

## Cluster commands

| Command | Status | Notes |
|---------|--------|-------|
| ASKING | 🚫 | |
| CLUSTER ADDSLOTS | 🚫 | |
| CLUSTER ADDSLOTSRANGE | 🚫 | |
| CLUSTER BUMPEPOCH | 🚫 | |
| CLUSTER COUNT-FAILURE-REPORTS | 🚫 | |
| CLUSTER COUNTKEYSINSLOT | 🚫 | |
| CLUSTER DELSLOTS | 🚫 | |
| CLUSTER DELSLOTSRANGE | 🚫 | |
| CLUSTER FAILOVER | 🚫 | |
| CLUSTER FLUSHSLOTS | 🚫 | |
| CLUSTER FORGET | 🚫 | |
| CLUSTER GETKEYSINSLOT | 🚫 | |
| CLUSTER INFO | 🚫 | |
| CLUSTER KEYSLOT | 🚫 | |
| CLUSTER LINKS | 🚫 | |
| CLUSTER MEET | 🚫 | |
| CLUSTER MIGRATION | 🚫 | |
| CLUSTER MYID | 🚫 | |
| CLUSTER MYSHARDID | 🚫 | |
| CLUSTER NODES | 🚫 | |
| CLUSTER REPLICAS | 🚫 | |
| CLUSTER REPLICATE | 🚫 | |
| CLUSTER RESET | 🚫 | |
| CLUSTER SAVECONFIG | 🚫 | |
| CLUSTER SET-CONFIG-EPOCH | 🚫 | |
| CLUSTER SETSLOT | 🚫 | |
| CLUSTER SHARDS | 🚫 | |
| CLUSTER SLAVES | 🚫 | |
| CLUSTER SLOTS | 🚫 | |
| READONLY | 🚫 | |
| READWRITE | 🚫 | |

---

## Generic (keys) commands

| Command | Status | Notes |
|---------|--------|-------|
| COPY | 🚫 | |
| DEL | ✅ | |
| DUMP | 🚫 | |
| EXISTS | ✅ | |
| EXPIRE | ✅ | |
| EXPIREAT | 🚫 | |
| EXPIRETIME | 🚫 | |
| KEYS | 🚫 | |
| MIGRATE | 🚫 | |
| MOVE | ✅ | |
| OBJECT ENCODING | 🚫 | |
| OBJECT FREQ | 🚫 | |
| OBJECT IDLETIME | 🚫 | |
| OBJECT REFCOUNT | 🚫 | |
| PERSIST | 🚫 | |
| PEXPIRE | 🚫 | |
| PEXPIREAT | 🚫 | |
| PEXPIRETIME | 🚫 | |
| PTTL | ✅ | |
| RANDOMKEY | 🚫 | |
| RENAME | ✅ | |
| RENAMENX | ✅ | |
| RESTORE | 🚫 | |
| SCAN | 🚫 | |
| SORT | 🚫 | |
| SORT_RO | 🚫 | |
| TOUCH | 🚫 | |
| TTL | ✅ | |
| TYPE | ✅ | |
| UNLINK | 🚫 | |
| WAIT | 🚫 | |

---

## JSON module commands

| Command | Status | Notes |
|---------|--------|-------|
| JSON.ARRAPPEND | ✅ | |
| JSON.ARRINDEX | ✅ | |
| JSON.ARRINSERT | ✅ | |
| JSON.ARRLEN | ✅ | |
| JSON.ARRPOP | ✅ | |
| JSON.ARRTRIM | ✅ | |
| JSON.CLEAR | ✅ | |
| JSON.DEBUG | 🚫 | |
| JSON.DEBUG MEMORY | 🚫 | |
| JSON.DEL | ✅ | |
| JSON.FORGET | 🚫 | |
| JSON.GET | ✅ | |
| JSON.MERGE | 🚫 | |
| JSON.MGET | ✅ | |
| JSON.MSET | 🚫 | |
| JSON.NUMINCRBY | ✅ | |
| JSON.NUMMULTBY | ✅ | |
| JSON.OBJKEYS | ✅ | |
| JSON.OBJLEN | ✅ | |
| JSON.RESP | ✅ | |
| JSON.SET | ✅ | |
| JSON.STRAPPEND | ✅ | |
| JSON.STRLEN | ✅ | |
| JSON.TOGGLE | 🚫 | |
| JSON.TYPE | ✅ | |

---

## Bloom Filter module commands

| Command | Status | Notes |
|---------|--------|-------|
| BF.ADD | ✅ | |
| BF.CARD | 🚫 | |
| BF.EXISTS | ✅ | |
| BF.INFO | ✅ | |
| BF.INSERT | ✅ | |
| BF.LOADCHUNK | 🚫 | |
| BF.MADD | ✅ | |
| BF.MEXISTS | ✅ | |
| BF.RESERVE | ✅ | |
| BF.SCANDUMP | 🚫 | |
