测试

```bash
docker run -it --rm --name etcd3 -p 2379:2379 -p 2380:2380 -e ETCD_ROOT_PASSWORD=123456 bitnami/etcd:3.5.3
```