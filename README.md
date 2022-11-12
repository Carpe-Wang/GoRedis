# GoRedis(该项目为本人2023年河南工业大学本科毕业课设)

### 项目简介

* 使用go语言实现Redis的基础命令，比如set，get和AOF等基础功能。

### 使用方法

* 单机启动

  * 直接启动程序即可。

* 集群启动

  * 通过`go build`命令生成可执行文件。

  * 把可执行文件单独放到一个文件夹，注：文件夹中需要包含配置(redis.conf)。

  * port不可重复，self为自己的ip+端口号，peers为集群中另外的ip+端口号。

  * ```
    bind 0.0.0.0
    port 6379
    
    appendonly yes
    appendfilename appendonly.aof
    
    self 127.0.0.1:6379
    peers 127.0.0.1:6380
    ```

  * 之后macOS或者linux进入相对应的文件夹执行`./goRedis`即可运行

  * 运行之后可以看到`[INFO][server.go:42] 2022/10/09 10:36:47 bind: 0.0.0.0:6380, start listening...`。

  * 通过网络调试助手连接TCP客户端`[INFO][server.go:71] 2022/10/09 10:37:30 accept link`。

  * 即可正常发送命令。
### 日志
* 日志信息信息主要记录项目启动时间。
* 是否有连接，连接时间等。

### AOF

* 主要通过项目下的appendonly.aof记录操作信息。
* 再重启系统时进行LoadAof防止重启机器信息丢失。

### 注：简易先看一下Resp协议，方便后期发送请求
>*3/r/n$3/r/nset/r/n$4/r/nname/r/n$5/r/npdudo/r/n


### 注释：mac
> 因为我本身使用的是macOS，在使用RESP的时候没办法直接用网络调试助手发送tcp，只能采用telnet命令。
> 注意重点就是需要先下载Homebrew，才能在mac的终端上执行命令。


## 成果展示
* 启动
  * ![img.png](img/img.png)
* set测试
  * ![img_1.png](img/img_1.png)
* AOF效果
  * ![img_3.png](img/img_3.png)
* Get测试
  * ![img_2.png](img/img_2.png)

--------------------------------------------
# 毕设准备
### 为什么选择本课题
* 因为看到Http3弃用TCP作为传输层协议通信方式，于是我想用自己的方式记录一下TCP。
* 对于后端软件开发来说，Redis是一个仅次于数据库的至关重要的部分。并且可以弥补多次请求数据库带来的IO请求的时间浪费。

### Redis为什么快
* 因为在计算机组成原理中，我们可以了解到，访问时间的快慢排序为：寄存器<L1,L2,L3缓存<内存<固态硬盘<机械硬盘。在Redis中，存储在内存中，然后持久化是存储在硬盘上，所以速度也会比传统数据库要更快。
* Redis的多路复用“复用”指的是复用同一个线程。采用多路 I/O 复用技术可以让单个线程高效的处理多个连接请求（尽量减少网络 IO 的时间消耗），且 Redis 在内存中操作数据的速度非常快，也就是说内存内的操作不会成为影响Redis性能的瓶颈，主要由以上几点造就了 Redis 具有很高的吞吐量。
* 完全基于内存，绝大部分请求是纯粹的内存操作，非常快速。数据存在内存中，类似于HashMap，HashMap的优势就是查找和操作的时间复杂度都是O(1)；

### 和传统数据库的区别
* Redis是非关系型数据库，是基于键值对的对应关系，用于超大规模数据的存储。
* nosql数据库将数据存储于缓存之中，关系型数据库将数据存储在硬盘中，自然查询速度远不及nosql数据库。
*  性能NOSQL是基于键值对的，可以想象成表中的主键和值的对应关系，而且不需要经过SQL层的解析，所以性能非常高。

# 设计细节
### 简易TCP服务器
* TCP相关配置参考`config.config`和`redis.conf`配置文件
* 项目中源码

```go
func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		// closing handler refuse new connection
		_ = conn.Close()
	}

	client := &EchoClient{
		Conn: conn,
	}
	h.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		// may occurs: client EOF, client timeout, handler early close
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("connection close")
				h.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}
```
>其中Context包主要用来传递超时时间，目前环境，参数等，Conn主要用来处理

```go
type Config struct {
	Address    string        `yaml:"address"`
	MaxConnect uint32        `yaml:"max-connect"`
	Timeout    time.Duration `yaml:"timeout"`
}
```
>在tcp包下的server层定义结构体，用来设定一些链接细节。

```go
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	ListenAndServe(listener, handler, closeChan)
	return nil
}
```
> 