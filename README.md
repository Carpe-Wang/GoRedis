# GoRedis

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

### AOF

* 主要通过项目下的appendonly.aof记录操作信息。
* 再重启系统时进行LoadAof防止重启机器信息丢失。

