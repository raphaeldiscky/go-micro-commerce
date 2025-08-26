# Redis Cluster

Manually create the cluster

````bash
docker exec -it redis-1 redis-cli -a supersecret --cluster create \
 redis-1:6379 redis-2:6379 redis-3:6379 redis-4:6379 redis-5:6379 redis-6:6379 \
 --cluster-replicas 1```
````
