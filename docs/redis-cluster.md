# Redis Cluster

Manually create the cluster

```bash
docker exec -it redis-1 redis-cli -a supersecret --cluster create \
  redis-1:6379 redis-2:6379 redis-3:6379 redis-4:6379 redis-5:6379 redis-6:6379 \
  --cluster-replicas 1 --cluster-yes
```

Check cluster info

```bash
docker exec -it redis-1 redis-cli -c -a supersecret cluster info
```

Check cluster nodes

```bash
docker exec -it redis-1 redis-cli -c -a supersecret cluster nodes
```

Check slot distribution

```bash
docker exec -it redis-1 redis-cli -c -a supersecret cluster slots
```
