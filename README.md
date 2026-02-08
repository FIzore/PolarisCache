# 🚀 LCache (PolarisCache)

![Go Version](https://img.shields.io/badge/go-1.21%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)

> A high-performance distributed cache system based on Golang.
>
> 基于 Go 语言实现的高性能分布式缓存系统，支持 LRU-2 淘汰、一致性哈希、SingleFlight 防击穿及 Etcd 服务注册发现。

## 📖 Introduction (项目介绍)

LCache 是一个轻量级、高可用的分布式缓存系统。它旨在解决高并发场景下的缓存击穿与数据一致性问题。
项目吸取了 GroupCache 的设计思想，并在此基础上进行了深度优化，引入了 LRU-K 算法、分段锁以及基于 Etcd 的动态节点管理，使其更适用于现代微服务架构。

## ✨ Key Features (核心特性)

- 多级缓存架构: 支持进程内缓存（Hot）与分布式节点缓存（Cold），有效降低网络开销。
- 智能淘汰算法: 实现了 LRU 及 LRU-2 (Two-Queues) 算法，针对不同访问模式优化缓存命中率。
- 高并发防护:
    - SingleFlight: 请求合并机制，杜绝热点 Key 击穿（Thundering Herd）。
    - Sharded Locks: 采用分段锁机制（默认 256 分片），显著减少锁争用。
- 分布式协调:
    - 一致性哈希 (Consistent Hashing): 支持虚拟节点，确保数据在节点增删时均匀迁移。
    - 服务发现: 集成 Etcd，实现节点的自动注册、发现与健康检查。
- 高性能通信: 节点间采用 Protobuf + gRPC 进行高效通信。

## 🛠️ Architecture (架构设计)

```text
                            +-------------+
                            |  Client     |
                            +------+------+
                                   |
                                   v
+-----------------------------------------------------------------------+
|  LCache Node (Coordinator)                                            |
|                                                                       |
|  +----------------+    +------------------+    +------------------+   |
|  | HTTP/RPC Server|<---| Consistent Hash  |--->|  Peer Picker     |   |
|  +----------------+    | (Virtual Nodes)  |    | (Etcd Discovery) |   |
|                        +------------------+    +---------+--------+   |
|                                                          |            |
|  +----------------+    +------------------+              |            |
|  | SingleFlight   |<---|   Local Cache    |              |            |
|  | (Anti-Stampede)|    | (LRU / LRU-2)    |              |            |
|  +-------+--------+    +------------------+              |            |
|          |                                               |            |
+----------|-----------------------------------------------|------------+
           |                                               |
           v (Cache Miss)                                  v (Remote Fetch)
    +-------------+                                 +-------------+
    |  Backend DB |                                 | Remote Peer |
    +-------------+                                 +-------------+