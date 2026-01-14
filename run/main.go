package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	// 关键修改：引用你自己的包名
	lcache "github.com/FIzore/PolarisCache"
)

func main() {
	// 1. 解析命令行参数
	var port int
	var nodeID string
	flag.IntVar(&port, "port", 8001, "节点监听端口")
	flag.StringVar(&nodeID, "node", "A", "节点标识符")
	flag.Parse()

	addr := fmt.Sprintf("localhost:%d", port) // 建议加上 localhost
	log.Printf("[节点 %s] 启动, 地址: %s", nodeID, addr)

	// 2. 创建节点 (Server) - 用于被别人访问
	// 对应截图 line 23
	node, err := lcache.NewServer(addr, "polaris-cache",
		lcache.WithEtcdEndpoints([]string{"localhost:2379"}),
		lcache.WithDialTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("创建节点失败: %v", err)
	}

	// 3. 创建节点选择器 (Picker) - 用于去找别人
	// 对应截图 line 32: 这里原版用的是 NewClientPicker，不是 NewEtcdPicker
	picker, err := lcache.NewClientPicker(addr)
	if err != nil {
		log.Fatalf("创建节点选择器失败: %v", err)
	}

	// 4. 创建缓存组
	// 对应截图 line 38
	group := lcache.NewGroup("test", 2<<20, lcache.GetterFunc(
		func(ctx context.Context, key string) ([]byte, error) {
			log.Printf("[节点 %s] 触发数据源加载: key=%s", nodeID, key)
			return []byte(fmt.Sprintf("节点%s的数据源值", nodeID)), nil
		}),
	)

	// 5. 注册节点选择器
	// 对应截图 line 46: 原版方法名是 RegisterPeers
	group.RegisterPeers(picker)

	// 6. 启动节点服务
	go func() {
		log.Printf("[节点 %s] 开始启动服务...", nodeID)
		if err := node.Start(); err != nil {
			log.Fatalf("启动节点失败: %v", err)
		}
	}()

	// 等待节点注册完成
	log.Printf("[节点 %s] 等待节点注册...", nodeID)
	time.Sleep(5 * time.Second)

	ctx := context.Background()

	// 7. 设置本节点的特定键值对 (测试 Set)
	localKey := fmt.Sprintf("key_%s", nodeID)
	// 修正：截图里的 fmt.Sprintf 格式字符串修正
	localValue := []byte(fmt.Sprintf("这是节点%s的数据", nodeID))

	fmt.Printf("\n=== 节点%s: 设置本地数据 ===\n", nodeID)
	err = group.Set(ctx, localKey, localValue)
	if err != nil {
		log.Fatalf("设置本地数据失败: %v", err)
	}
	fmt.Printf("[节点%s]: 设置键 %s 成功\n", nodeID, localKey)

	// 等待其他节点也完成设置
	// 截图 line 75 是 30秒，为了测试快一点我改成了 10秒，你可以改回去
	log.Printf("[节点 %s] 等待其他节点准备就绪...", nodeID)
	time.Sleep(10 * time.Second)

	// 8. 打印当前已发现的节点 (验证 Etcd 是否工作)
	// 对应截图 line 78
	picker.PrintPeers()

	// 9. 测试获取本地数据
	fmt.Printf("\n=== 节点%s: 获取本地数据 ===\n", nodeID)
	fmt.Printf("直接查询本地缓存...\n") // 修正了原本的 Pritnf 拼写错误

	// 打印统计信息
	stats := group.Stats()
	fmt.Printf("缓存统计: %+v\n", stats)

	if val, err := group.Get(ctx, localKey); err == nil {
		fmt.Printf("[节点%s]: 获取本地键 %s 成功: %s\n", nodeID, localKey, val.String())
	} else {
		fmt.Printf("[节点%s]: 获取本地键失败: %v\n", nodeID, err)
	}

	// 10. 测试获取其他节点的数据 (验证分布式能力)
	// 对应截图 line 95
	otherKeys := []string{"key_A", "key_B", "key_C"}
	for _, key := range otherKeys {
		if key == localKey {
			continue
		}
		fmt.Printf("\n=== 节点%s: 尝试获取远程数据 %s ===\n", nodeID, key)
		log.Printf("[节点%s] 开始查找键 %s 的远程节点", nodeID, key)

		if val, err := group.Get(ctx, key); err == nil {
			fmt.Printf("[节点%s]: 获取远程键 %s 成功: %s\n", nodeID, key, val.String())
		} else {
			fmt.Printf("[节点%s]: 获取远程键失败: %v\n", nodeID, err)
		}
	}

	// 保持程序运行
	select {}
}